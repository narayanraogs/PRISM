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

type GTxAttnMeasurement struct {
	deviceProfile   string
	rxName          string
	linkedRxs       []string
	spectrumProfile string
	component       string
	maxPower        float64
	minPower        float64
	stepSize        float64
	frequency       int64
	sa              driver.SA
	gtx             driver.GTX
	currentStatus   [][]string
	statusMonitor   chan AttnMeasurementStatus
	deviations      []CorrectedDeviation
	stop            bool
}

func (gtx *GTxAttnMeasurement) GetStatusMonitor() chan AttnMeasurementStatus {
	return gtx.statusMonitor
}

func (gtx *GTxAttnMeasurement) Initialize(deviceProfile string, rxName string, spectrumProfile string,
	component string, maxPower float64, minPower float64, stepSize float64) {
	gtx.deviceProfile = deviceProfile
	gtx.rxName = rxName
	gtx.spectrumProfile = spectrumProfile
	gtx.component = component
	gtx.maxPower = maxPower
	gtx.minPower = minPower
	gtx.stepSize = stepSize
	gtx.currentStatus = make([][]string, 0)
	gtx.statusMonitor = make(chan AttnMeasurementStatus, 20)
	ok := gtx.loadDevices()
	if !ok {
		gtx.setError("Unable to Load devices")
		return
	}
	ok = gtx.loadDetails()
	if !ok {
		gtx.setError("Unable to Load Details")
		return
	}
	header := make([]string, 0)
	header = append(header, "Sl. No", "Set Power", "Measured Power", "Deviation")
	gtx.currentStatus = append(gtx.currentStatus, header)
	gtx.stop = false
}

func (gtx *GTxAttnMeasurement) Stop() {
	gtx.stop = true
}

func (gtx *GTxAttnMeasurement) loadDevices() bool {
	saName, ok := database.GetSAFromDeviceProfile(gtx.deviceProfile)
	if !ok {
		return false
	}
	gtxName, ok := database.GetGTxFromDeviceProfile(gtx.deviceProfile)
	if !ok {
		return false
	}
	ok = gtx.sa.LoadDevice(saName)
	if !ok {
		return false
	}
	ok = gtx.gtx.LoadDevice(gtxName)
	if !ok {
		return false
	}
	return true
}

func (gtx *GTxAttnMeasurement) loadDetails() bool {
	freq, ok := database.GetRxFrequency(gtx.rxName)
	if !ok {
		return false
	}
	rxs, ok := database.GetAllRxWithFrequency(freq)
	if !ok {
		return false
	}
	gtx.frequency = freq
	gtx.linkedRxs = make([]string, 0)
	gtx.linkedRxs = append(gtx.linkedRxs, rxs...)
	return true
}

func (gtx *GTxAttnMeasurement) setError(message string) {
	var measure = AttnMeasurementStatus{
		Completed: true,
		Error:     true,
		Message:   message,
	}
	gtx.statusMonitor <- measure
	close(gtx.statusMonitor)
}

func (gtx *GTxAttnMeasurement) StartMeasurement() {
	var measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "GTx Power Measurement Started",
	}
	gtx.statusMonitor <- measure
	response := gtx.sa.SetAlignmentOff()
	if !response.Success {
		gtx.setError("Unable to communicate with SA")
		return
	}
	response = gtx.gtx.SetPower(gtx.component, 0)
	if !response.Success {
		gtx.setError("Unable to communicate with GTx")
		return
	}
	time.Sleep(200 * time.Millisecond)

	defer func() {
		gtx.sa.SetAlignmentOn()
		gtx.sa.SystemPreset()
		gtx.gtx.SetCarrierOff(gtx.component)
	}()

	response = gtx.gtx.SetModulationOff(gtx.component)
	if !response.Success {
		gtx.setError("Unable to communicate with GTx")
		return
	}
	time.Sleep(200 * time.Millisecond)
	response = gtx.gtx.SetCarrierOn(gtx.component)
	if !response.Success {
		gtx.setError("Unable to communicate with GTx")
		return
	}
	time.Sleep(200 * time.Millisecond)
	spectrum, ok := database.GetSpectrumProfile(gtx.spectrumProfile)
	if !ok {
		gtx.setError("Unable to Read Spectrum from Database")
		return
	}

	response = gtx.sa.SetSpectrum(spectrum.CenterFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		gtx.setError("Unable to communicate with SA")
		return
	}

	response = gtx.sa.SetReferenceNominal()
	if !response.Success {
		gtx.setError("Carrier Not found")
		return
	}

	time.Sleep(1 * time.Second)

	response = gtx.sa.GetMaxMarkerValue()
	if !response.Success {
		gtx.setError("Unable to communicate with SA")
		return
	}
	initalPower := response.Result["MarkerY"].Value
	slNo := 1

	for powerSet := gtx.minPower; powerSet <= gtx.maxPower; powerSet = powerSet + gtx.stepSize {
		if gtx.stop {
			gtx.setError("Measurement Aborted by User")
			return
		}
		powerStr := fmt.Sprintf("%.3f", powerSet)
		response = gtx.gtx.SetPower(gtx.component, powerSet)
		if !response.Success {
			gtx.setError("GTx Power Cannot be set to " + powerStr)
			return
		}

		time.Sleep(200 * time.Millisecond)

		response = gtx.sa.GetMaxMarkerValue()
		if !response.Success {
			gtx.setError("SA Power Cannot be read")
			return
		}
		actualPower := response.Result["MarkerY"].Value - initalPower

		actualPowerStr := fmt.Sprintf("%.3f", actualPower)
		slNo = slNo + 1
		slNoStr := strconv.Itoa(slNo)

		difference := actualPower - powerSet
		differenceStr := fmt.Sprintf("%.3f", difference)
		row := make([]string, 0)
		row = append(row, slNoStr, powerStr, actualPowerStr, differenceStr)
		gtx.currentStatus = append(gtx.currentStatus, row)

		measure = AttnMeasurementStatus{
			Completed:     false,
			Error:         false,
			Message:       "Completed Measurement for " + powerStr,
			PlotDeviation: true,
		}
		measure.AddData(slNo, powerSet, actualPower, difference)
		gtx.statusMonitor <- measure
	}
	measure = AttnMeasurementStatus{
		Completed: false,
		Error:     false,
		Message:   "Saving Results",
		HasData:   false,
	}
	gtx.statusMonitor <- measure

	var csv strings.Builder
	for _, row := range gtx.currentStatus {
		csv.WriteString(strings.Join(row, ","))
		csv.WriteString("\n")
	}

	for _, rx := range gtx.linkedRxs {
		fileName := utils.Config.BaseFolder + "/.resources/gtx-" + rx + ".csv"
		err := os.WriteFile(fileName, []byte(csv.String()), 0755)
		if err != nil {
			fmt.Println("Cannot write file", fileName)
		}
	}

	var requried = make([]float64, 0)
	var measured = make([]float64, 0)
	var difference = make([]float64, 0)
	for i, row := range gtx.currentStatus {
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
	correctedStruct = utils.GetCorrectedProfile(measuredStruct, 0, gtx.stepSize)

	gtx.deviations = make([]CorrectedDeviation, 0)
	for i := 0; i < len(requried); i++ {
		var temp CorrectedDeviation
		temp.SetValue = requried[i]
		temp.MeasuredDeviation = difference[i]
		temp.CorrectedDeviation = correctedStruct.GetDeviation(requried[i])
		gtx.deviations = append(gtx.deviations, temp)
	}
	measure = AttnMeasurementStatus{
		Completed:     true,
		Error:         false,
		Message:       "Measurement Completed",
		HasData:       false,
		PlotDeviation: false,
	}
	gtx.statusMonitor <- measure
	close(gtx.statusMonitor)
}

func (gtx *GTxAttnMeasurement) GetCorrectedDeviations() []CorrectedDeviation {
	return gtx.deviations
}
