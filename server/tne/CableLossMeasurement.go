package tne

import (
	"encoding/json"
	"fmt"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/resultsDB"
	"slices"
	"strconv"
	"strings"
	"time"
)

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
	statusMonitor chan MeasurementStatus
}

type MeasurementStatus struct {
	CurrentStatus [][]string
	Message       string
	Completed     bool
	Success       bool
}

func (clm *CableLossMeasurement) GetStatusMonitor() chan MeasurementStatus {
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
	clm.statusMonitor = make(chan MeasurementStatus, 20)
	clm.loadDevices()
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
		var measure = MeasurementStatus{
			Completed:     true,
			Success:       false,
			Message:       "Unable to get details from Database",
			CurrentStatus: make([][]string, 0),
		}
		clm.statusMonitor <- measure
		close(clm.statusMonitor)
		return nil
	}
	return freqs
}

func (clm *CableLossMeasurement) measureForFrequencies(frequencies []string, offset map[string]float64, measureAll bool) bool {
	response := clm.pm.SetChAAverageOff()
	if !response.Success {
		clm.setError("Unable to communicate with PM")
		return false
	}
	response = clm.pm.SetChBAverageOff()
	if !response.Success {
		clm.setError("Unable to communicate with PM")
		return false
	}
	response = clm.sg.SetModOff()
	if !response.Success {
		clm.setError("Unable to communicate with SG")
		return false
	}
	defer func() {
		clm.pm.SetChAAverageOn()
		clm.pm.SetChBAverageOn()
		clm.sg.SetRFOff()
		fmt.Println("Restored")
	}()

	for i, freq := range frequencies {
		if clm.stop {
			clm.setError("User Aborted")
			return false
		}
		if !measureAll {
			if slices.Index(clm.frequencies, freq) == -1 {
				clm.currentStatus[1][i] = "-"
				continue
			}
		}
		var measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Measuring Loss for " + freq + " Hz",
			CurrentStatus: make([][]string, 0),
		}
		clm.statusMonitor <- measure

		f, _ := strconv.ParseFloat(freq, 64)
		response = clm.sg.SetFrequency(f)
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return false
		}
		power := offset[freq]
		power = power * -1
		response = clm.sg.SetPower(power)
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return false
		}
		response = clm.sg.SetRFOn()
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return false
		}
		if strings.EqualFold(clm.pmChannel, "A") {
			response = clm.pm.SetChannelA(f)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return false
			}
			response = clm.pm.GetPowerChannelA(true)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return false
			}
			if response.Result["ChannelBPower"].Value < -60 {
				clm.setError("Power read is less than -60dBm. Check PM Connection")
				return false
			}
			clm.currentStatus[1][i] = fmt.Sprintf("%.2f", response.Result["Power"].Value)
		} else {
			response = clm.pm.SetChannelB(f)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return false
			}
			response = clm.pm.GetPowerChannelB(true)
			if !response.Success {
				clm.setError("Unable to communicate with PM")
				return false
			}
			if response.Result["ChannelBPower"].Value < -60 {
				clm.setError("Power read is less than -60dBm. Check PM Connection")
				return false
			}
			clm.currentStatus[1][i] = fmt.Sprintf("%.2f", response.Result["Power"].Value)
		}
		response = clm.sg.SetRFOff()
		if !response.Success {
			clm.setError("Unable to communicate with SG")
			return false
		}
	}

	return true
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

	ok := clm.measureForFrequencies(frequencies, offset, true)
	if !ok {
		return
	}

	var measurement cableLossMeasured
	measurement.Frequency = clm.currentStatus[0]
	measurement.Measured = clm.currentStatus[1]
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

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
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

	ok := clm.measureForFrequencies(frequencies, pmReference, false)
	if !ok {
		return
	}

	var measurement cableLossMeasured
	measurement.Frequency = clm.currentStatus[0]
	measurement.Measured = clm.currentStatus[1]
	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}

	length, _ := strconv.Atoi(cableLength)

	resultsDB.InsertCableLoss(clm.startDate, clm.startTime, cableName, length, string(jsonData))

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	clm.statusMonitor <- measure
	close(clm.statusMonitor)

}

func (clm *CableLossMeasurement) setError(message string) {
	var measure = MeasurementStatus{
		Completed:     true,
		Success:       false,
		Message:       message,
		CurrentStatus: make([][]string, 0),
	}
	clm.statusMonitor <- measure
	close(clm.statusMonitor)
}
