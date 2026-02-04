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

type TVACCableLossRecord struct {
	SlNo         int                `json:"slNo"`
	CableName    string             `json:"cableName"`
	CycleName    string             `json:"cycleName"`
	Phase        string             `json:"phase"`
	Date         string             `json:"date"`
	Time         string             `json:"time"`
	IsReference  bool               `json:"isReference"`
	Measurements []MeasurementPoint `json:"measurements"`
}

type MeasurementPoint struct {
	Frequency float64 `json:"frequency"` // GHz
	Loss      float64 `json:"loss"`      // Absolute Loss (dB)
	Delta     float64 `json:"delta"`     // Difference from reference (dB)
}

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
	stop          bool
	statusMonitor chan MeasurementStatus
}

type MeasurementStatus struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
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
	tclm.stop = false
	tclm.reference = "1"

	if !tclm.loadDevices() {
		tclm.statusMonitor <- MeasurementStatus{
			Message: "Failed to load device profiles for " + deviceProfile,
			Error:   true,
		}
		close(tclm.statusMonitor)
		return
	}
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
	if !ok {
		var measure = MeasurementStatus{
			Message: "Unable to get details from Database",
			Error:   true,
		}
		tclm.statusMonitor <- measure
		close(tclm.statusMonitor)
		return nil
	}
	return freqs
}

func (tclm *TVACCableLossMeasurement) measureForFrequencies(frequencies []string, offset map[string]float64) ([]float64, bool) {
	response := tclm.pm.SetChAAverageOff()
	if !response.Success {
		tclm.setError("Unable to communicate with PM")
		return nil, false
	}
	response = tclm.pm.SetChBAverageOff()
	if !response.Success {
		tclm.setError("Unable to communicate with PM")
		return nil, false
	}
	response = tclm.sg.SetModOff()
	if !response.Success {
		tclm.setError("Unable to communicate with SG")
		return nil, false
	}
	defer func() {
		tclm.pm.SetChAAverageOn()
		tclm.pm.SetChBAverageOn()
		tclm.sg.SetRFOff()
	}()
	var measuredLosses = make([]float64, 0)

	for _, freq := range frequencies {
		if tclm.stop {
			tclm.setError("User Aborted")
			return nil, false
		}

		var measure = MeasurementStatus{
			Message: "Measuring Loss for " + freq + " Hz",
			Error:   false,
		}
		tclm.statusMonitor <- measure

		f, _ := strconv.ParseFloat(freq, 64)
		response = tclm.sg.SetFrequency(f)
		if !response.Success {
			tclm.setError("Unable to communicate with SG")
			return nil, false
		}
		power := offset[freq]
		power = power * -1
		response = tclm.sg.SetPower(power)
		if !response.Success {
			tclm.setError("Unable to communicate with SG when setting power")
			return nil, false
		}
		response = tclm.sg.SetRFOn()
		if !response.Success {
			tclm.setError("Unable to communicate with SG when setting RF")
			return nil, false
		}
		if strings.EqualFold(tclm.pmChannel, "A") {
			response = tclm.pm.SetChannelA(f)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return nil, false
			}
			response = tclm.pm.GetPowerChannelA(true)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return nil, false
			}
			if response.Result["Power"].Value < -60 {
				tclm.setError("Power read is less than -60dBm. Check PM Connection")
				return nil, false
			}
			measuredLosses = append(measuredLosses, response.Result["Power"].Value)
		} else {
			response = tclm.pm.SetChannelB(f)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return nil, false
			}
			response = tclm.pm.GetPowerChannelB(true)
			if !response.Success {
				tclm.setError("Unable to communicate with PM")
				return nil, false
			}
			if response.Result["Power"].Value < -60 {
				tclm.setError("Power read is less than -60dBm. Check PM Connection")
				return nil, false
			}
			measuredLosses = append(measuredLosses, response.Result["Power"].Value)
		}
		response = tclm.sg.SetRFOff()
		if !response.Success {
			tclm.setError("Unable to communicate with SG")
			return nil, false
		}
	}
	return measuredLosses, true
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

	var measurements []MeasurementPoint

	pmOffset, ok := tclm.measureForFrequencies(frequencies, offset)
	if !ok {
		return
	}

	for i, freq := range frequencies {
		var m MeasurementPoint
		m.Frequency, _ = strconv.ParseFloat(freq, 64)
		m.Loss = pmOffset[i]
		m.Delta = 0.0
		measurements = append(measurements, m)
	}
	jsonData, err := json.MarshalIndent(measurements, "", " ")
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
		Error:   false,
		Message: "Measurement Completed",
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

	losses, ok := tclm.measureForFrequencies(frequencies, cableReference)
	if !ok {
		return
	}

	var measurement []MeasurementPoint = make([]MeasurementPoint, 0)
	for i, freq := range frequencies {
		var m MeasurementPoint
		m.Frequency, _ = strconv.ParseFloat(freq, 64)
		m.Loss = losses[i]
		m.Delta = 0.0
		measurement = append(measurement, m)
	}
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}
	resultsDB.InsertTVACCableLoss(tclm.startDate, tclm.startTime, cableName, tclm.testPhase, "1", string(jsonData))

	var measure = MeasurementStatus{
		Error:   false,
		Message: "Measurement Completed",
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
	var pmMeasured []MeasurementPoint
	err := json.Unmarshal([]byte(ref), &pmMeasured)
	if err != nil {
		return nil
	}
	for _, m := range pmMeasured {
		tbr[fmt.Sprintf("%.2f", m.Frequency)] = m.Loss
	}
	return tbr
}

func (tclm *TVACCableLossMeasurement) getTVACCableReference(cableName string) map[string]float64 {
	var tbr = make(map[string]float64)
	ref, ok := resultsDB.GetTVACCableReference(cableName)
	if !ok {
		return nil
	}
	var refCableMeasured []MeasurementPoint
	err := json.Unmarshal([]byte(ref), &refCableMeasured)
	if err != nil {
		return nil
	}
	for _, m := range refCableMeasured {
		tbr[fmt.Sprintf("%.2f", m.Frequency)] = m.Loss
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
	refLosses := tclm.getTVACCableReference(cableName)

	losses, ok := tclm.measureForFrequencies(frequencies, pmReference)
	if !ok {
		return
	}

	var measurement []MeasurementPoint = make([]MeasurementPoint, 0)

	for i, freq := range frequencies {
		var m MeasurementPoint
		m.Frequency, _ = strconv.ParseFloat(freq, 64)
		m.Loss = losses[i]
		m.Delta = losses[i] - refLosses[freq]
		measurement = append(measurement, m)
	}

	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}

	resultsDB.InsertTVACCableLoss(tclm.startDate, tclm.startTime, cableName, tclm.testPhase, "0", string(jsonData))

	measure := MeasurementStatus{
		Error:   false,
		Message: "Measurement Completed",
	}
	tclm.statusMonitor <- measure

	close(tclm.statusMonitor)
}

func (tclm *TVACCableLossMeasurement) setError(message string) {
	var measure = MeasurementStatus{
		Message: message,
		Error:   true,
	}
	tclm.statusMonitor <- measure
	close(tclm.statusMonitor)
}
