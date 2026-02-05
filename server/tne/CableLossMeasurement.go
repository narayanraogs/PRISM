package tne

import (
	"encoding/json"
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/resultsDB"
	"slices"
	"strconv"
	"strings"
	"time"
)

type CableLossRecord struct {
	SlNo         int                `json:"slNo"`
	CableName    string             `json:"cableName"`
	Length       float64            `json:"length"`
	Date         string             `json:"date"`
	Time         string             `json:"time"`
	Measurements []MeasurementPoint `json:"measurements"` // JSON data from DB
}

type MeasurementPoint struct {
	Frequency float64 `json:"frequency"` // Value in GHz
	Loss      float64 `json:"loss"`      // Value in dB
}

type CableLossMeasurement struct {
	pmChannel     string
	deviceProfile string
	frequencies   []string
	pm            driver.PM
	sg            driver.SG
	startDate     string
	startTime     string
	currentStatus [][]string
	stop          bool
	statusMonitor chan RTStatus
}

type RTStatus struct {
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
	Success   bool   `json:"success"`
	Error     bool   `json:"error"`
}

type MeasurementStatus struct {
	CurrentStatus [][]string
	Message       string
	Completed     bool
	Success       bool
}

func (clm *CableLossMeasurement) GetStatusMonitor() chan RTStatus {
	return clm.statusMonitor
}

type cableLossMeasured struct {
	Frequency []string
	Measured  []string
}

func (clm *CableLossMeasurement) Initialize(pmChannel string, deviceProfile string, frequencies []string) {
	clm.pmChannel = pmChannel
	clm.deviceProfile = deviceProfile
	clm.frequencies = frequencies
	clm.statusMonitor = make(chan RTStatus, 20)
	ok := clm.loadDevices()
	if !ok {
		clm.setError("Unable to load devices")
		return
	}
	clm.stop = false
}

func (clm *CableLossMeasurement) Stop() {
	clm.stop = true
}

func (clm *CableLossMeasurement) loadDevices() bool {
	pmName, ok := database.GetPMFromDeviceProfile(clm.deviceProfile)
	if !ok {
		return false
	}
	sgName, ok := database.GetSGFromDeviceProfile(clm.deviceProfile)
	if !ok {
		return false
	}
	ok = clm.pm.LoadDevice(pmName)
	if !ok {
		return false
	}
	ok = clm.sg.LoadDevice(sgName)
	if !ok {
		return false
	}
	return true
}

func (clm *CableLossMeasurement) startMeasurement() []string {
	now := time.Now()
	clm.startDate = now.Format("02-01-2006")
	clm.startTime = now.Format("15:04:05")
	frequencies, ok := database.GetLossMeasurementFrequencyNames()
	var freqs = make([]string, 0)
	var f float64
	var freqsToBeMeasured = make([]string, 0)

	for _, freq := range frequencies {
		f, ok = database.GetFrequencyForLossMeasurement(freq)
		if !ok {
			continue
		}
		freqs = append(freqs, fmt.Sprintf("%.2f", f))
		if slices.Index(clm.frequencies, freq) != -1 {
			freqsToBeMeasured = append(freqsToBeMeasured, fmt.Sprintf("%.2f", f))
		}
	}
	clm.frequencies = freqsToBeMeasured
	clm.currentStatus = make([][]string, 0)
	clm.currentStatus = append(clm.currentStatus, freqs)
	measured := make([]string, len(freqs))
	clm.currentStatus = append(clm.currentStatus, measured)
	if !ok {
		var measure = RTStatus{
			Completed: true,
			Success:   false,
			Error:     true,
			Message:   "Unable to get details from Database",
		}
		clm.statusMonitor <- measure
		close(clm.statusMonitor)
		return nil
	}
	return freqs
}

func (clm *CableLossMeasurement) measureForFrequencies(frequencies []string, offset map[string]float64, measureAll bool) ([]float64, bool) {
	response := clm.pm.SetChAAverageOff()
	if !response.Success {
		clm.setError("Unable to communicate with PM")
		return nil, false
	}
	response = clm.pm.SetChBAverageOff()
	if !response.Success {
		clm.setError("Unable to communicate with PM")
		return nil, false
	}
	response = clm.sg.SetModOff()
	if !response.Success {
		clm.setError("Unable to communicate with SG")
		return nil, false
	}
	defer func() {
		clm.pm.SetChAAverageOn()
		clm.pm.SetChBAverageOn()
		clm.sg.SetRFOff()
	}()
	var result = make([]float64, 0)

	for _, freq := range frequencies {
		if clm.stop {
			clm.setError("User Aborted")
			return nil, false
		}
		if !measureAll {
			if slices.Index(clm.frequencies, freq) == -1 {
				result = append(result, math.NaN())
				continue
			}
		}
		var measure = RTStatus{
			Completed: false,
			Success:   true,
			Error:     false,
			Message:   "Measuring Loss for " + freq + " Hz",
		}
		clm.statusMonitor <- measure

		f, _ := strconv.ParseFloat(freq, 64)
		response = clm.sg.SetFrequency(f)
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return nil, false
		}
		power := offset[freq]
		power = power * -1
		response = clm.sg.SetPower(power)
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return nil, false
		}
		response = clm.sg.SetRFOn()
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return nil, false
		}
		if strings.EqualFold(clm.pmChannel, "A") {
			response = clm.pm.SetChannelA(f)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return nil, false
			}
			response = clm.pm.GetPowerChannelA(true)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return nil, false
			}
			if response.Result["Power"].Value < -60 {
				clm.setError("Power read is less than -60dBm. Check PM Connection")
				return nil, false
			}
			result = append(result, response.Result["Power"].Value)
		} else {
			response = clm.pm.SetChannelB(f)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return nil, false
			}
			response = clm.pm.GetPowerChannelB(true)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return nil, false
			}
			if response.Result["Power"].Value < -60 {
				clm.setError("Power read is less than -60dBm. Check PM Connection")
				return nil, false
			}
			result = append(result, response.Result["Power"].Value)
		}
		response = clm.sg.SetRFOff()
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return nil, false
		}
	}

	return result, true
}

func (clm *CableLossMeasurement) MeasurePMReference() {
	frequencies := clm.startMeasurement()
	if frequencies == nil {
		return
	}

	offset := make(map[string]float64)
	for _, freq := range frequencies {
		offset[freq] = 0
	}

	result, ok := clm.measureForFrequencies(frequencies, offset, true)
	if !ok {
		return
	}

	var measurement []MeasurementPoint
	for i, freq := range frequencies {
		var mp MeasurementPoint
		mp.Frequency, _ = strconv.ParseFloat(freq, 64)
		mp.Loss = result[i]
		if !math.IsNaN(mp.Loss) {
			measurement = append(measurement, mp)
		}
	}
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}
	exist := resultsDB.CheckIfCableLossPMReferenceExists()
	if exist {
		resultsDB.UpdateCableLossPMReference(clm.startDate, clm.startTime, string(jsonData))
	} else {
		resultsDB.InsertCableLoss(clm.startDate, clm.startTime, "PM", 0, string(jsonData))
	}

	var measure = RTStatus{
		Completed: true,
		Success:   true,
		Error:     false,
		Message:   "Measurement Completed",
	}
	clm.statusMonitor <- measure
	close(clm.statusMonitor)
}

func (clm *CableLossMeasurement) getPMReference() map[string]float64 {
	var tbr = make(map[string]float64)
	ref, ok := resultsDB.GetCableLossPMReference()
	if !ok {
		return nil
	}
	var pmMeasured []MeasurementPoint
	err := json.Unmarshal([]byte(ref), &pmMeasured)
	if err != nil {
		return nil
	}
	for _, mp := range pmMeasured {
		tbr[fmt.Sprintf("%.2f", mp.Frequency)] = mp.Loss
	}
	return tbr
}

func (clm *CableLossMeasurement) MeasureCableLoss(cableName string, cableLength string) {
	frequencies := clm.startMeasurement()
	if frequencies == nil {
		return
	}
	pmReference := clm.getPMReference()
	if pmReference == nil {
		clm.setError("Unable to read PM Reference, Rerun PM Reference")
		return
	}

	result, ok := clm.measureForFrequencies(frequencies, pmReference, false)
	if !ok {
		return
	}

	var measurement []MeasurementPoint
	for i, freq := range frequencies {
		var mp MeasurementPoint
		mp.Frequency, _ = strconv.ParseFloat(freq, 64)
		mp.Loss = result[i]
		if !math.IsNaN(mp.Loss) {
			measurement = append(measurement, mp)
		}
	}
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		clm.setError("Unable to marshal data to JSON")
		return
	}

	length, _ := strconv.ParseFloat(cableLength, 64)

	fmt.Println("Saving to database, ", clm.startDate, clm.startTime, cableName, length, string(jsonData))
	ok = resultsDB.InsertCableLoss(clm.startDate, clm.startTime, cableName, length, string(jsonData))
	if !ok {
		clm.setError("Unable to save data to database")
		return
	}

	var measure = RTStatus{
		Completed: true,
		Success:   true,
		Error:     false,
		Message:   "Measurement Completed",
	}
	clm.statusMonitor <- measure
	close(clm.statusMonitor)

}

func (clm *CableLossMeasurement) setError(message string) {
	var measure = RTStatus{
		Completed: true,
		Success:   false,
		Error:     true,
		Message:   message,
	}
	clm.statusMonitor <- measure
	close(clm.statusMonitor)
}
