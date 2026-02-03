package utilities

import (
	"encoding/json"
	"fmt"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/resultsDB"
	"strconv"
	"strings"
	"time"
)

type TVACCableLossMeasurement struct {
	pmChannel     string
	deviceProfile string
	frequencies   []string
	pm            driver.PM
	sg            driver.SG
	startDate     string
	startTime     string
	reference     string
	newCable      bool
	testPhase     string
	currentStatus [][]string
	stop          bool
	statusMonitor chan MeasurementStatus
}

type MeasurementStatus struct {
	CurrentStatus [][]string
	Message       string
	Completed     bool
	Success       bool
}

func (tclm *TVACCableLossMeasurement) GetStatusMonitor() chan MeasurementStatus {
	return tclm.statusMonitor
}

type cableLossMeasured struct {
	Frequency []string
	Measured  []string
}

func (tclm *TVACCableLossMeasurement) Initialize(pmChannel string, deviceProfile string, testPhase string) {
	tclm.pmChannel = pmChannel
	tclm.deviceProfile = deviceProfile
	tclm.testPhase = testPhase

	tclm.statusMonitor = make(chan MeasurementStatus, 20)
	tclm.loadDevices()
	tclm.stop = false
}

func (tclm *TVACCableLossMeasurement) Stop() {
	tclm.stop = true
}

func (tclm *TVACCableLossMeasurement) loadDevices() bool {
	pmName, ok := database.GetPMFromDeviceProfile(tclm.deviceProfile)
	if !ok {
		return false
	}
	sgName, ok := database.GetSGFromDeviceProfile(tclm.deviceProfile)
	if !ok {
		return false
	}
	ok = tclm.pm.LoadDevice(pmName)
	if !ok {
		return false
	}
	ok = tclm.sg.LoadDevice(sgName)
	if !ok {
		return false
	}
	return true
}

func (tclm *TVACCableLossMeasurement) startMeasurement() []string {
	now := time.Now()
	tclm.startDate = now.Format("02-01-2006")
	tclm.startTime = now.Format("15:04:05")
	frequencies, ok := database.GetLossMeasurementFrequencyNames()
	var freqs = make([]string, 0)
	var f float64

	for _, freq := range frequencies {
		f, ok = database.GetFrequencyForLossMeasurement(freq)
		if !ok {
			continue
		}
		freqs = append(freqs, fmt.Sprintf("%.2f", f))
	}
	tclm.frequencies = freqs
	tclm.currentStatus = make([][]string, 0)
	tclm.currentStatus = append(tclm.currentStatus, freqs)
	measured := make([]string, len(freqs))
	tclm.currentStatus = append(tclm.currentStatus, measured)
	if !ok {
		var measure = MeasurementStatus{
			Completed:     true,
			Success:       false,
			Message:       "Unable to get details from Database",
			CurrentStatus: make([][]string, 0),
		}
		tclm.statusMonitor <- measure
		close(tclm.statusMonitor)
		return nil
	}
	return freqs
}

func (tclm *TVACCableLossMeasurement) measureForFrequencies(frequencies []string, offset map[string]float64) bool {
	response := tclm.pm.SetChAAverageOff()
	if !response.Success {
		tclm.setError("Unable to communicate with PM")
		return false
	}
	response = tclm.pm.SetChBAverageOff()
	if !response.Success {
		tclm.setError("Unable to communicate with PM")
		return false
	}
	response = tclm.sg.SetModOff()
	if !response.Success {
		tclm.setError("Unable to communicate with SG")
		return false
	}
	defer func() {
		tclm.pm.SetChAAverageOn()
		tclm.pm.SetChBAverageOn()
		tclm.sg.SetRFOff()
	}()

	for i, freq := range frequencies {
		if tclm.stop {
			tclm.setError("User Aborted")
			return false
		}

		var measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Measuring Loss for " + freq + " Hz",
			CurrentStatus: make([][]string, 0),
		}
		tclm.statusMonitor <- measure

		f, _ := strconv.ParseFloat(freq, 64)
		response = tclm.sg.SetFrequency(f)
		if !response.Success {
			tclm.setError("Unable to communicate with SG")
			return false
		}
		power := offset[freq]
		power = power * -1
		response = tclm.sg.SetPower(power)
		if !response.Success {
			tclm.setError("Unable to communicate with SG")
			return false
		}
		response = tclm.sg.SetRFOn()
		if !response.Success {
			tclm.setError("Unable to communicate with SG")
			return false
		}
		if strings.EqualFold(tclm.pmChannel, "A") {
			response = tclm.pm.SetChannelA(f)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return false
			}
			response = tclm.pm.GetPowerChannelA(true)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return false
			}
			if response.Result["ChannelBPower"].Value < -60 {
				tclm.setError("Power read is less than -60dBm. Check PM Connection")
				return false
			}
			tclm.currentStatus[1][i] = fmt.Sprintf("%.2f", response.Result["Power"].Value)
		} else {
			response = tclm.pm.SetChannelB(f)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return false
			}
			response = tclm.pm.GetPowerChannelB(true)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return false
			}
			if response.Result["ChannelBPower"].Value < -60 {
				tclm.setError("Power read is less than -60dBm. Check PM Connection")
				return false
			}
			tclm.currentStatus[1][i] = fmt.Sprintf("%.2f", response.Result["Power"].Value)
		}
		response = tclm.sg.SetRFOff()
		if !response.Success {
			tclm.setError("Unable to communicate with SG")
			return false
		}
	}
	return true
}

func (tclm *TVACCableLossMeasurement) MeasurePMReference() {
	frequencies := tclm.startMeasurement()
	if frequencies == nil {
		return
	}

	offset := make(map[string]float64)
	for _, freq := range frequencies {
		offset[freq] = 0
	}

	ok := tclm.measureForFrequencies(frequencies, offset)
	if !ok {
		return
	}

	var measurement cableLossMeasured
	measurement.Frequency = tclm.currentStatus[0]
	measurement.Measured = tclm.currentStatus[1]
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}
	exist := resultsDB.CheckIfTVACCableLossPMReferenceExists()
	if exist {
		resultsDB.UpdateTVACCableLossPMReference(tclm.startDate, tclm.startTime, tclm.reference, string(jsonData))
	} else {
		resultsDB.InsertTVACCableLoss(tclm.startDate, tclm.startTime, "PM", "", tclm.reference, string(jsonData))
	}

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tclm.statusMonitor <- measure
	close(tclm.statusMonitor)
}

func (tclm *TVACCableLossMeasurement) MeasureTVACReference(cableName string, testPhase string) {
	frequencies := tclm.startMeasurement()
	if frequencies == nil {
		return
	}
	cableReference := tclm.getTVACPMReference()
	if cableReference == nil {
		tclm.setError("Unable to read Cable Reference, Rerun Cable Reference")
		return
	}

	ok := tclm.measureForFrequencies(frequencies, cableReference)
	if !ok {
		return
	}

	var measurement cableLossMeasured
	measurement.Frequency = tclm.currentStatus[0]
	measurement.Measured = tclm.currentStatus[1]
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}

	resultsDB.InsertTVACCableLoss(tclm.startDate, tclm.startTime, cableName, tclm.testPhase, "1", string(jsonData))

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tclm.statusMonitor <- measure
	close(tclm.statusMonitor)

}

func (tclm *TVACCableLossMeasurement) getTVACPMReference() map[string]float64 {
	var tbr = make(map[string]float64)
	ref, ok := resultsDB.GetTVACCableLossPMReference()
	if !ok {
		return nil
	}
	var pmMeasured cableLossMeasured
	err := json.Unmarshal([]byte(ref), &pmMeasured)
	if err != nil {
		return nil
	}
	for i, freq := range pmMeasured.Frequency {
		value, err := strconv.ParseFloat(pmMeasured.Measured[i], 64)
		if err != nil {
			continue
		}
		tbr[freq] = value
	}
	return tbr
}

func (tclm *TVACCableLossMeasurement) getTVACCableReference(cableName string) map[string]float64 {
	var tbr = make(map[string]float64)
	ref, ok := resultsDB.GetTVACCableReference(cableName)
	if !ok {
		return nil
	}
	var refCableMeasured cableLossMeasured
	err := json.Unmarshal([]byte(ref), &refCableMeasured)
	if err != nil {
		return nil
	}
	for i, freq := range refCableMeasured.Frequency {
		value, err := strconv.ParseFloat(refCableMeasured.Measured[i], 64)
		if err != nil {
			continue
		}
		tbr[freq] = value
	}
	return tbr
}

func (tclm *TVACCableLossMeasurement) MeasureTVACCableLoss(cableName string, testPhase string) {
	frequencies := tclm.startMeasurement()
	if frequencies == nil {
		return
	}
	pmReference := tclm.getTVACPMReference()
	if pmReference == nil {
		tclm.setError("Unable to read PM Reference, Rerun PM Reference")
		return
	}

	ok := tclm.measureForFrequencies(frequencies, pmReference)
	if !ok {
		return
	}

	var measurement cableLossMeasured
	measurement.Frequency = tclm.currentStatus[0]
	measurement.Measured = tclm.currentStatus[1]
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}

	resultsDB.InsertTVACCableLoss(tclm.startDate, tclm.startTime, cableName, tclm.testPhase, "0", string(jsonData))

	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Computing Difference",
		CurrentStatus: make([][]string, 0),
	}
	tclm.statusMonitor <- measure

	var computed cableLossMeasured
	computed.Frequency = make([]string, 0)
	computed.Measured = make([]string, 0)

	refLosses := tclm.getTVACCableReference(cableName)
	for i, frequency := range measurement.Frequency {
		refLoss := refLosses[frequency]
		measuredLoss, _ := strconv.ParseFloat(measurement.Measured[i], 64)
		diff := measuredLoss - refLoss
		computed.Frequency = append(computed.Frequency, frequency)
		computed.Measured = append(computed.Measured, fmt.Sprintf("%0.2f", diff))
	}

	jsonData, err = json.MarshalIndent(computed, "", " ")
	if err != nil {
		return
	}

	resultsDB.InsertTVACCableLoss(tclm.startDate, tclm.startTime, cableName, tclm.testPhase, "Diff", string(jsonData))

	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tclm.statusMonitor <- measure

	close(tclm.statusMonitor)
}

func (tclm *TVACCableLossMeasurement) setError(message string) {
	var measure = MeasurementStatus{
		Completed:     true,
		Success:       false,
		Message:       message,
		CurrentStatus: make([][]string, 0),
	}
	tclm.statusMonitor <- measure
	close(tclm.statusMonitor)
}
