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

type SGPowerMeasurement struct {
	deviceProfile   string
	rxName          string
	linkedRxs       []string
	spectrumProfile string
	maxPower        float64
	minPower        float64
	stepSize        float64
	frequency       int64
	sa              driver.SA
	sg              driver.SG
	currentStatus   [][]string
	statusMonitor   chan AttnMeasurementStatus
	deviations      []CorrectedDeviation
	stop            bool
}

func (sg *SGPowerMeasurement) GetStatusMonitor() chan AttnMeasurementStatus {
	return sg.statusMonitor
}

func (sg *SGPowerMeasurement) Initialize(deviceProfile string, rxName string, spectrumProfile string,
	maxPower float64, minPower float64, stepSize float64) {
	sg.deviceProfile = deviceProfile
	sg.rxName = rxName
	sg.spectrumProfile = spectrumProfile
	sg.maxPower = maxPower
	sg.minPower = minPower
	sg.stepSize = stepSize
	sg.currentStatus = make([][]string, 0)
	sg.statusMonitor = make(chan AttnMeasurementStatus, 20)
	ok := sg.loadDevices()
	if !ok {
		sg.setError("Unable to load devices")
		return
	}
	ok = sg.loadDetails()
	if !ok {
		sg.setError("Unable to load details")
		return
	}
	header := make([]string, 0)
	header = append(header, "Sl. No", "Set Power", "Measured Power", "Deviation")
	sg.currentStatus = append(sg.currentStatus, header)
	sg.stop = false
}

func (sg *SGPowerMeasurement) Stop() {
	sg.stop = true
}

func (sg *SGPowerMeasurement) loadDevices() bool {
	saName, ok := database.GetSAFromDeviceProfile(sg.deviceProfile)
	if !ok {
		return false
	}
	sgName, ok := database.GetSGFromDeviceProfile(sg.deviceProfile)
	if !ok {
		return false
	}
	ok = sg.sa.LoadDevice(saName)
	if !ok {
		return false
	}
	ok = sg.sg.LoadDevice(sgName)
	if !ok {
		return false
	}
	return true
}

func (sg *SGPowerMeasurement) loadDetails() bool {
	freq, ok := database.GetRxFrequency(sg.rxName)
	if !ok {
		return false
	}
	rxs, ok := database.GetAllRxWithFrequency(freq)
	if !ok {
		return false
	}
	sg.frequency = freq
	sg.linkedRxs = make([]string, 0)
	sg.linkedRxs = append(sg.linkedRxs, rxs...)
	return true
}

func (sg *SGPowerMeasurement) setError(message string) {
	var measure = AttnMeasurementStatus{
		Completed: true,
		Error:     true,
		Message:   message,
		HasData:   false,
	}
	sg.statusMonitor <- measure
	close(sg.statusMonitor)
}

func (sg *SGPowerMeasurement) StartMeasurement() {
	var measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "SG Power Measurement Started",
		HasData:   false,
	}
	sg.statusMonitor <- measure
	response := sg.sa.SetAlignmentOff()
	if !response.Success {
		sg.setError("Unable to communicate with SA")
		return
	}
	response = sg.sg.SetFrequency(float64(sg.frequency))
	if !response.Success {
		sg.setError("Unable to communicate with SG")
		return
	}

	defer func() {
		sg.sa.SetAlignmentOn()
		sg.sa.SystemPreset()
		sg.sg.SetRFOff()
	}()

	response = sg.sg.SetModOff()
	if !response.Success {
		sg.setError("Unable to communicate with SG")
		return
	}

	spectrum, ok := database.GetSpectrumProfile(sg.spectrumProfile)
	if !ok {
		sg.setError("Unable to Read Spectrum from Database")
		return
	}

	response = sg.sa.SetSpectrum(spectrum.CenterFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		sg.setError("Unable to communicate with SA")
		return
	}

	response = sg.sg.SetPower(0)
	if !response.Success {
		sg.setError("Unable to communicate with SG")
		return
	}

	response = sg.sg.SetRFOn()
	if !response.Success {
		sg.setError("Unable to communicate with SG")
		return
	}

	response = sg.sa.SetReferenceNominal()
	if !response.Success {
		sg.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = sg.sa.GetMaxMarkerValue()
	if !response.Success {
		sg.setError("Unable to communicate with SA")
		return
	}
	cableLoss := response.Result["MarkerY"].Value
	slNo := 0
	slNoStr := strconv.Itoa(slNo)

	for powerSet := sg.minPower; powerSet <= sg.maxPower; powerSet = powerSet + sg.stepSize {
		if sg.stop {
			sg.setError("Measurement Aborted by User")
			return
		}
		powerStr := fmt.Sprintf("%.3f", powerSet)

		response = sg.sg.SetPower(powerSet)
		if !response.Success {
			sg.setError("SG Power Cannot be set to " + powerStr)
			return
		}

		time.Sleep(1000 * time.Millisecond)

		response = sg.sa.GetMaxMarkerValue()
		if !response.Success {
			sg.setError("SA Power Cannot be read")
			return
		}
		actualPower := response.Result["MarkerY"].Value - cableLoss

		actualPowerStr := fmt.Sprintf("%.3f", actualPower)
		slNo = slNo + 1
		slNoStr = strconv.Itoa(slNo)

		difference := actualPower - powerSet
		differenceStr := fmt.Sprintf("%.3f", difference)
		row := make([]string, 0)
		row = append(row, slNoStr, powerStr, actualPowerStr, differenceStr)
		sg.currentStatus = append(sg.currentStatus, row)
		measure = AttnMeasurementStatus{
			Completed:     false,
			Error:         false,
			PlotDeviation: true,
			Message:       "Completed Measurement for " + powerStr,
		}
		measure.AddData(slNo, powerSet, actualPower, difference)
		sg.statusMonitor <- measure
	}
	measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "Saving Results",
	}
	sg.statusMonitor <- measure

	var csv strings.Builder
	for _, row := range sg.currentStatus {
		csv.WriteString(strings.Join(row, ","))
		csv.WriteString("\n")
	}

	for _, rx := range sg.linkedRxs {
		fileName := utils.Config.BaseFolder + "/.resources/sg-" + rx + ".csv"
		err := os.WriteFile(fileName, []byte(csv.String()), 0755)
		if err != nil {
			fmt.Println("Cannot write file", fileName)
		}
	}

	var requried = make([]float64, 0)
	var measured = make([]float64, 0)
	var difference = make([]float64, 0)
	for i, row := range sg.currentStatus {
		if i == 0 {
			continue
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
	correctedStruct = utils.GetCorrectedProfile(measuredStruct, 0, sg.stepSize)

	sg.deviations = make([]CorrectedDeviation, 0)
	for i := 0; i < len(requried); i++ {
		var temp CorrectedDeviation
		temp.SetValue = requried[i]
		temp.MeasuredDeviation = difference[i]
		temp.CorrectedDeviation = correctedStruct.GetDeviation(requried[i])
		sg.deviations = append(sg.deviations, temp)
	}

	measure = AttnMeasurementStatus{
		Completed: true,
		Error:     false,
		Message:   "Measurement Completed",
	}
	sg.statusMonitor <- measure
	close(sg.statusMonitor)
}

func (sg *SGPowerMeasurement) GetCorrectedDeviations() []CorrectedDeviation {
	return sg.deviations
}
