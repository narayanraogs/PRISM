package tne

import (
	"fmt"
	"os"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type TSMAttnMeasurement struct {
	deviceProfile    string
	rxName           string
	linkedRxs        []string
	spectrumProfile  string
	tsmConfiguration string
	maxPower         float64
	minPower         float64
	stepSize         float64
	frequency        int64
	tsmConfig        database.TSMConfiguration
	sa               driver.SA
	tsm              driver.TSM
	sg               driver.SG
	currentStatus    [][]string
	statusMonitor    chan AttnMeasurementStatus
	deviations       []CorrectedDeviation

	stop bool
}

type CorrectedDeviation struct {
	SetValue           float64
	MeasuredDeviation  float64
	CorrectedDeviation float64
}

type AttnMeasurementStatus struct {
	SlNo          int
	SetAttn       float64
	MeasuredAttn  float64
	Deviation     float64
	HasData       bool
	Completed     bool
	Error         bool
	Message       string
	PlotDeviation bool
}

func (t *AttnMeasurementStatus) AddData(slNo int, setAttn float64, measured float64, deviation float64) {
	t.SlNo = slNo
	t.SetAttn = setAttn
	t.MeasuredAttn = measured
	t.Deviation = deviation
	t.HasData = true
}

func (tsm *TSMAttnMeasurement) GetStatusMonitor() chan AttnMeasurementStatus {
	return tsm.statusMonitor
}

func (tsm *TSMAttnMeasurement) Initialize(deviceProfile string, rxName string, spectrumProfile string, tsmConfiguration string,
	maxPower float64, minPower float64, stepSize float64) {
	tsm.deviceProfile = deviceProfile
	tsm.rxName = rxName
	tsm.spectrumProfile = spectrumProfile
	tsm.tsmConfiguration = tsmConfiguration
	tsm.maxPower = maxPower
	tsm.minPower = minPower
	tsm.stepSize = stepSize
	tsm.currentStatus = make([][]string, 0)
	tsm.currentStatus = append(tsm.currentStatus, []string{"Sl No", "Set Attn", "Measured Attn", "Deviation"})
	tsm.statusMonitor = make(chan AttnMeasurementStatus, 20)
	ok := tsm.loadDevices()
	if !ok {
		tsm.setError("Unable to load devices")
		return
	}
	ok = tsm.loadDetails()
	if !ok {
		tsm.setError("Unable to load details")
		return
	}
	tsm.stop = false
}

func (tsm *TSMAttnMeasurement) Stop() {
	tsm.stop = true
}

func (tsm *TSMAttnMeasurement) loadDevices() bool {
	saName, ok := database.GetSAFromDeviceProfile(tsm.deviceProfile)
	if !ok {
		return false
	}
	sgName, ok := database.GetSGFromDeviceProfile(tsm.deviceProfile)
	if !ok {
		return false
	}
	tsmName, ok := database.GetTSMFromDeviceProfile(tsm.deviceProfile)
	if !ok {
		return false
	}
	ok = tsm.sa.LoadDevice(saName)
	if !ok {
		return false
	}
	ok = tsm.sg.LoadDevice(sgName)
	if !ok {
		return false
	}
	ok = tsm.tsm.LoadDevice(tsmName)
	if !ok {
		return false
	}
	return true
}

func (tsm *TSMAttnMeasurement) loadDetails() bool {
	freq, ok := database.GetRxFrequency(tsm.rxName)
	if !ok {
		return false
	}
	rxs, ok := database.GetAllRxWithFrequency(freq)
	if !ok {
		return false
	}
	tsm.frequency = freq
	tsm.linkedRxs = make([]string, 0)
	tsm.linkedRxs = append(tsm.linkedRxs, rxs...)
	paths, ok := database.GetTSMPathDetails(tsm.tsmConfiguration)
	if !ok {
		return false
	}
	tsm.tsmConfig = paths
	return true
}

func (tsm *TSMAttnMeasurement) setError(message string) {
	var measure = AttnMeasurementStatus{
		Completed: true,
		Error:     true,
		Message:   message,
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMAttnMeasurement) StartMeasurement() {
	var measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "TSM Power Measurement Started",
	}
	tsm.statusMonitor <- measure
	response := tsm.sa.SetAlignmentOff()
	if !response.Success {
		tsm.setError("Unable to communicate with SA")
		return
	}
	response = tsm.tsm.SetDriverStatus(tsm.tsmConfig.UplinkToSC.String)
	if !response.Success {
		tsm.setError("Unable to communicate with TSM")
		return
	}
	time.Sleep(500 * time.Millisecond)
	response = tsm.tsm.SetAttn(int(tsm.tsmConfig.AttnNumber.Int64), 0)
	if !response.Success {
		tsm.setError("Unable to communicate with TSM")
		return
	}
	response = tsm.sg.SetPower(0)
	if !response.Success {
		tsm.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		tsm.tsm.SetDriverStatus(tsm.tsmConfig.TerminateUplink.String)
		time.Sleep(500 * time.Millisecond)
		tsm.sa.SetAlignmentOn()
		tsm.sa.SystemPreset()
		tsm.sg.SetRFOff()
		tsm.tsm.SetAttn(int(tsm.tsmConfig.AttnNumber.Int64), 0)
	}()
	response = tsm.sg.SetFrequency(float64(tsm.frequency))
	if !response.Success {
		tsm.setError("Unable to communicate with SG")
		return
	}
	response = tsm.sg.SetModOff()
	if !response.Success {
		tsm.setError("Unable to communicate with SG")
		return
	}
	response = tsm.sg.SetRFOn()
	if !response.Success {
		tsm.setError("Unable to communicate with SG")
		return
	}

	spectrum, ok := database.GetSpectrumProfile(tsm.spectrumProfile)
	if !ok {
		tsm.setError("Unable to Read Spectrum from Database")
		return
	}

	response = tsm.sa.SetSpectrum(spectrum.CenterFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		tsm.setError("Unable to communicate with SA")
		return
	}

	if tsm.tsmConfig.ExcludePad.Valid {
		response = tsm.tsm.SetDriverStatus(tsm.tsmConfig.ExcludePad.String)
		if !response.Success {
			tsm.setError("Unable to communicate with TSM")
			return
		}
	}

	measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "Measuring initial power",
	}
	tsm.statusMonitor <- measure

	response = tsm.sa.SetReferenceNominal()
	if !response.Success {
		tsm.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = tsm.sa.GetMaxMarkerValue()
	if !response.Success {
		tsm.setError("Unable to communicate with SA")
		return
	}
	initialPower := response.Result["MarkerY"].Value
	slNo := 0

	if tsm.tsmConfig.IncludePad.Valid {
		measure = AttnMeasurementStatus{
			Completed: false,
			Error:     false,
			Message:   "Measuring Fixed Pad Attenuation",
		}
		tsm.statusMonitor <- measure

		response = tsm.tsm.SetDriverStatus(tsm.tsmConfig.IncludePad.String)
		if !response.Success {
			tsm.setError("Unable to communicate with TSM")
			return
		}
		time.Sleep(200 * time.Millisecond)
		response = tsm.sa.GetMaxMarkerValue()
		if !response.Success {
			tsm.setError("Unable to communicate with SA")
			return
		}
		power := response.Result["MarkerY"].Value - initialPower
		slNo = slNo + 1
		response = tsm.tsm.SetDriverStatus(tsm.tsmConfig.ExcludePad.String)
		if !response.Success {
			tsm.setError("Unable to communicate with TSM")
			return
		}
		measure = AttnMeasurementStatus{
			Completed:     false,
			Error:         false,
			Message:       "Completed Measuring Fixed Pad Attenuation",
			PlotDeviation: false,
		}
		measure.AddData(slNo, 0, power, 0)
		tsm.currentStatus = append(tsm.currentStatus, []string{fmt.Sprintf("%d", slNo), "FixedPad", fmt.Sprintf("%.3f", power), "-"})
		tsm.statusMonitor <- measure
		time.Sleep(200 * time.Millisecond)
	}

	for attn := tsm.minPower; attn <= tsm.maxPower; attn = attn + tsm.stepSize {
		if tsm.stop {
			tsm.setError("User Aborted")
			return
		}
		setAttn := fmt.Sprintf("%.3f", attn)
		response = tsm.tsm.SetAttn(int(tsm.tsmConfig.AttnNumber.Int64), attn)
		if !response.Success {
			if !response.Success {
				tsm.setError("Unable to communicate with TSM")
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
		response = tsm.sa.GetMaxMarkerValue()
		if !response.Success {
			if !response.Success {
				tsm.setError("Unable to communicate with TSM")
				return
			}
		}
		power := response.Result["MarkerY"].Value
		actualAttn := initialPower - power
		difference := attn - actualAttn
		slNo = slNo + 1
		measure = AttnMeasurementStatus{
			Completed:     false,
			Error:         false,
			Message:       "Completed Measuring for " + setAttn + " dB",
			PlotDeviation: true,
		}
		measure.AddData(slNo, attn, actualAttn, difference)
		tsm.statusMonitor <- measure
		tsm.currentStatus = append(tsm.currentStatus, []string{fmt.Sprintf("%d", slNo), setAttn, fmt.Sprintf("%.3f", actualAttn), fmt.Sprintf("%.3f", difference)})
		time.Sleep(200 * time.Millisecond)
	}
	measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "Saving Results",
		HasData:   false,
	}
	tsm.statusMonitor <- measure
	var requried = make([]float64, 0)
	var measured = make([]float64, 0)
	var difference = make([]float64, 0)
	var csv strings.Builder
	var fixed float64
	for i, row := range tsm.currentStatus {
		csv.WriteString(strings.Join(row, ","))
		csv.WriteString("\n")
		if i == 0 {
			continue
		}
		if i == 1 {
			fixed, _ = strconv.ParseFloat(row[2], 64)
		}
		tempR, _ := strconv.ParseFloat(row[1], 64)
		tempM, _ := strconv.ParseFloat(row[2], 64)
		tempD, _ := strconv.ParseFloat(row[3], 64)
		requried = append(requried, tempR)
		measured = append(measured, tempM)
		difference = append(difference, tempD)
	}

	var measuredStruct utils.TSMAttnProvider
	measuredStruct.RequiredAttn = requried
	measuredStruct.MeasuredAttn = measured
	measuredStruct.Difference = difference
	var correctedStruct utils.TSMAttnProvider
	correctedStruct = utils.GetCorrectedProfile(measuredStruct, fixed, tsm.stepSize)

	tsm.deviations = make([]CorrectedDeviation, 0)
	for i := 1; i < len(requried); i++ {
		var temp CorrectedDeviation
		temp.SetValue = requried[i]
		temp.MeasuredDeviation = difference[i]
		temp.CorrectedDeviation = correctedStruct.GetDeviation(requried[i])
		tsm.deviations = append(tsm.deviations, temp)
	}

	for _, rx := range tsm.linkedRxs {
		fileName := utils.Config.BaseFolder + "/.resources/tsm-" + rx + ".csv"
		err := os.WriteFile(fileName, []byte(csv.String()), 0755)
		if err != nil {
			fmt.Println("Cannot write file", fileName)
		}
	}

	measure = AttnMeasurementStatus{
		Completed: true,
		Error:     false,
		Message:   "Measurement Completed",
		HasData:   false,
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMAttnMeasurement) GetCorrectedDeviations() []CorrectedDeviation {
	return tsm.deviations
}
