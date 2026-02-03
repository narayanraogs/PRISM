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
	statusMonitor   chan MeasurementStatus
	stop            bool
}

func (sg *SGPowerMeasurement) GetStatusMonitor() chan MeasurementStatus {
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
	sg.statusMonitor = make(chan MeasurementStatus, 20)
	sg.loadDevices()
	sg.loadDetails()
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
	var measure = MeasurementStatus{
		Completed:     true,
		Success:       false,
		Message:       message,
		CurrentStatus: make([][]string, 0),
	}
	sg.statusMonitor <- measure
	close(sg.statusMonitor)
}

func (sg *SGPowerMeasurement) StartMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "SG Power Measurement Started",
		CurrentStatus: make([][]string, 0),
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
	slNo := 1
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
		slNoStr = strconv.Itoa(slNo)
		slNo = slNo + 1
		difference := actualPower - powerSet
		differenceStr := fmt.Sprintf("%.3f", difference)
		row := make([]string, 0)
		row = append(row, slNoStr, powerStr, actualPowerStr, differenceStr)
		sg.currentStatus = append(sg.currentStatus, row)
		measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Completed Measurement for " + powerStr,
			CurrentStatus: make([][]string, 0),
		}
		measure.CurrentStatus = append(measure.CurrentStatus, row)
		sg.statusMonitor <- measure
	}
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: make([][]string, 0),
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

	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	sg.statusMonitor <- measure
	close(sg.statusMonitor)
}
