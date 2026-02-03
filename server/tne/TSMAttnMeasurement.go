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
	statusMonitor    chan MeasurementStatus
	stop             bool
}

func (tsm *TSMAttnMeasurement) GetStatusMonitor() chan MeasurementStatus {
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
	tsm.statusMonitor = make(chan MeasurementStatus, 20)
	tsm.loadDevices()
	tsm.loadDetails()
	header := make([]string, 0)
	header = append(header, "Sl. No", "Set Attn", "Measured Attn", "Deviation")
	tsm.currentStatus = append(tsm.currentStatus, header)
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
	var measure = MeasurementStatus{
		Completed:     true,
		Success:       false,
		Message:       message,
		CurrentStatus: make([][]string, 0),
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMAttnMeasurement) StartMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "TSM Power Measurement Started",
		CurrentStatus: make([][]string, 0),
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

	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Measuring initial power",
		CurrentStatus: make([][]string, 0),
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
	slNo := 1

	if tsm.tsmConfig.IncludePad.Valid {
		measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Measuring Fixed Pad Attenuation",
			CurrentStatus: make([][]string, 0),
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
		slNoStr := strconv.Itoa(slNo)
		slNo = slNo + 1
		powerStr := fmt.Sprintf("%.3f", power)
		row := make([]string, 0)
		row = append(row, slNoStr, "FixedPad", powerStr, "-")
		tsm.currentStatus = append(tsm.currentStatus, row)
		response = tsm.tsm.SetDriverStatus(tsm.tsmConfig.ExcludePad.String)
		if !response.Success {
			tsm.setError("Unable to communicate with TSM")
			return
		}
		measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Completed Measuring Fixed Pad Attenuation",
			CurrentStatus: make([][]string, 0),
		}
		measure.CurrentStatus = append(measure.CurrentStatus, row)
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
		actualAttnStr := fmt.Sprintf("%.3f", actualAttn)
		difference := attn - actualAttn
		differenceStr := fmt.Sprintf("%.3f", difference)
		slNoStr := strconv.Itoa(slNo)
		slNo = slNo + 1
		row := make([]string, 0)
		row = append(row, slNoStr, setAttn, actualAttnStr, differenceStr)
		tsm.currentStatus = append(tsm.currentStatus, row)
		measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Completed Measuring for " + setAttn + " dB",
			CurrentStatus: make([][]string, 0),
		}
		measure.CurrentStatus = append(measure.CurrentStatus, row)
		tsm.statusMonitor <- measure
	}
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: make([][]string, 0),
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
	//plot measured and corrected in GNU plot and send to client
	fmt.Println(correctedStruct)

	for _, rx := range tsm.linkedRxs {
		fileName := utils.Config.BaseFolder + "/.resources/tsm-" + rx + ".csv"
		err := os.WriteFile(fileName, []byte(csv.String()), 0755)
		if err != nil {
			fmt.Println("Cannot write file", fileName)
		}
	}

	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}
