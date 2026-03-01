package tne

import (
	"encoding/json"
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/resultsDB"
	"prismServer/utils"
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
	stop          bool
	statusMonitor chan RTStatus
}

// Internal Helpers

func (clm *CableLossMeasurement) notify(msg string) {
	clm.statusMonitor <- RTStatus{Message: msg, Success: true}
}

func (clm *CableLossMeasurement) finish(msg string, success bool) {
	clm.statusMonitor <- RTStatus{
		Message:   msg,
		Success:   success,
		Error:     !success,
		Completed: true,
	}
	close(clm.statusMonitor)
}

func (clm *CableLossMeasurement) setError(msg string) {
	clm.finish(msg, false)
}

func (clm *CableLossMeasurement) check(resp utils.CommandResponse, errMsg string) bool {
	if !resp.Success {
		clm.setError(fmt.Sprintf("%s: %s", errMsg, resp.ErrorMessage))
		return false
	}
	if clm.stop {
		clm.setError("Measurement aborted by user")
		return false
	}
	return true
}

func (clm *CableLossMeasurement) saveResults(points []MeasurementPoint, testType, cableName string, length float64) bool {
	jsonData, err := json.MarshalIndent(points, "", " ")
	if err != nil {
		clm.setError("Failed to process measurement data")
		return false
	}

	if strings.EqualFold(testType, "PM_REF") {
		if resultsDB.CheckIfCableLossPMReferenceExists() {
			resultsDB.UpdateCableLossPMReference(clm.startDate, clm.startTime, string(jsonData))
		} else {
			resultsDB.InsertCableLoss(clm.startDate, clm.startTime, "PM", 0, string(jsonData))
		}
	} else {
		if !resultsDB.InsertCableLoss(clm.startDate, clm.startTime, cableName, length, string(jsonData)) {
			clm.setError("Failed to save results to database")
			return false
		}
	}
	return true
}

// Public API

func (clm *CableLossMeasurement) GetStatusMonitor() chan RTStatus {
	return clm.statusMonitor
}

func (clm *CableLossMeasurement) Initialize(pmChannel, deviceProfile string, frequencies []string) {
	clm.pmChannel = pmChannel
	clm.deviceProfile = deviceProfile
	clm.frequencies = frequencies
	clm.statusMonitor = make(chan RTStatus, 20)
	clm.stop = false

	pmName, okPm := database.GetPMFromDeviceProfile(deviceProfile)
	sgName, okSg := database.GetSGFromDeviceProfile(deviceProfile)
	if !okPm || !okSg {
		clm.setError("Unable to resolve devices from profile")
		return
	}

	if !clm.pm.LoadDevice(pmName) || !clm.sg.LoadDevice(sgName) {
		clm.setError("Failed to load device drivers")
	}
}

func (clm *CableLossMeasurement) Stop() {
	clm.stop = true
}

func (clm *CableLossMeasurement) prepareSession() []string {
	now := time.Now()
	clm.startDate = now.Format("02-01-2006")
	clm.startTime = now.Format("15:04:05")

	allFreqNames, _ := database.GetLossMeasurementFrequencyNames()
	var allFreqs []string
	var selectedFreqs []string

	for _, name := range allFreqNames {
		val, ok := database.GetFrequencyForLossMeasurement(name)
		if !ok {
			continue
		}
		fStr := fmt.Sprintf("%.2f", val)
		allFreqs = append(allFreqs, fStr)
		if slices.Index(clm.frequencies, name) != -1 {
			selectedFreqs = append(selectedFreqs, fStr)
		}
	}

	clm.frequencies = selectedFreqs
	return allFreqs
}

func (clm *CableLossMeasurement) measureForFrequencies(freqs []string, offset map[string]float64, measureAll bool) ([]float64, bool) {
	if !clm.check(clm.pm.SetChAAverageOff(), "PM: ChA average off") {
		return nil, false
	}
	if !clm.check(clm.pm.SetChBAverageOff(), "PM: ChB average off") {
		return nil, false
	}
	if !clm.check(clm.sg.SetModOff(), "SG: mod off") {
		return nil, false
	}

	defer clm.pm.SetChAAverageOn()
	defer clm.pm.SetChBAverageOn()
	defer clm.sg.SetRFOff()

	var results []float64

	for _, fStr := range freqs {
		if !measureAll && slices.Index(clm.frequencies, fStr) == -1 {
			results = append(results, math.NaN())
			continue
		}

		clm.notify(fmt.Sprintf("Measuring loss for %s Hz", fStr))
		f, _ := strconv.ParseFloat(fStr, 64)

		if !clm.check(clm.sg.SetFrequency(f), "SG: set frequency") {
			return nil, false
		}

		power := offset[fStr] * -1
		if !clm.check(clm.sg.SetPower(power), "SG: set power") {
			return nil, false
		}
		if !clm.check(clm.sg.SetRFOn(), "SG: RF on") {
			return nil, false
		}

		var pResp utils.CommandResponse
		if strings.EqualFold(clm.pmChannel, "A") {
			clm.pm.SetChannelA(f)
			pResp = clm.pm.GetPowerChannelA(true)
		} else {
			clm.pm.SetChannelB(f)
			pResp = clm.pm.GetPowerChannelB(true)
		}

		if !clm.check(pResp, "PM: read power") {
			return nil, false
		}

		val := pResp.Result["Power"].Value
		if val < -60 {
			clm.setError("Power too low (<-60dBm). Check connections")
			return nil, false
		}

		results = append(results, val)
		clm.sg.SetRFOff()
	}

	return results, true
}

func (clm *CableLossMeasurement) MeasurePMReference() {
	allFreqs := clm.prepareSession()
	if len(allFreqs) == 0 {
		clm.setError("No frequencies found in database")
		return
	}

	offsets := make(map[string]float64)
	for _, f := range allFreqs {
		offsets[f] = 0
	}

	results, ok := clm.measureForFrequencies(allFreqs, offsets, true)
	if !ok {
		return
	}

	var points []MeasurementPoint
	for i, fStr := range allFreqs {
		f, _ := strconv.ParseFloat(fStr, 64)
		points = append(points, MeasurementPoint{Frequency: f, Loss: results[i]})
	}

	if clm.saveResults(points, "PM_REF", "", 0) {
		clm.finish("Reference measurement completed", true)
	}
}

func (clm *CableLossMeasurement) MeasureCableLoss(cableName, cableLength string) {
	allFreqs := clm.prepareSession()
	if len(allFreqs) == 0 {
		clm.setError("No frequencies found in database")
		return
	}

	pmRef := clm.getPMReference()
	if pmRef == nil {
		clm.setError("PM reference not found. Please run reference measurement first")
		return
	}

	results, ok := clm.measureForFrequencies(allFreqs, pmRef, false)
	if !ok {
		return
	}

	var points []MeasurementPoint
	for i, fStr := range allFreqs {
		if math.IsNaN(results[i]) {
			continue
		}
		f, _ := strconv.ParseFloat(fStr, 64)
		points = append(points, MeasurementPoint{Frequency: f, Loss: results[i]})
	}

	length, _ := strconv.ParseFloat(cableLength, 64)
	if clm.saveResults(points, "CABLE", cableName, length) {
		clm.finish("Cable measurement completed", true)
	}
}

func (clm *CableLossMeasurement) getPMReference() map[string]float64 {
	ref, ok := resultsDB.GetCableLossPMReference()
	if !ok {
		return nil
	}

	var points []MeasurementPoint
	if err := json.Unmarshal([]byte(ref), &points); err != nil {
		return nil
	}

	pmMap := make(map[string]float64)
	for _, p := range points {
		pmMap[fmt.Sprintf("%.2f", p.Frequency)] = p.Loss
	}
	return pmMap
}
