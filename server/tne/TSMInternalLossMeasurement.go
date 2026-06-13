package tne

import (
	"encoding/json"
	"fmt"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/logger"
	"prismServer/resultsDB"
	"prismServer/utils"
	"slices"
	"strconv"
	"strings"
	"time"
)

type TSMInternalLoss struct {
	table              [][]string
	pm                 driver.PM
	sg                 driver.SG
	tsm                driver.TSM
	statusMonitor      chan RTStatus
	pmChannel          string
	stop               bool
	currentMeasurement map[string]float64
}

// Internal Helpers

func (tsm *TSMInternalLoss) notify(msg string) {
	tsm.statusMonitor <- RTStatus{Message: msg, Success: true}
}

func (tsm *TSMInternalLoss) finish(msg string, success bool) {
	tsm.statusMonitor <- RTStatus{
		Message:   msg,
		Success:   success,
		Error:     !success,
		Completed: true,
	}
	close(tsm.statusMonitor)
}

func (tsm *TSMInternalLoss) setError(msg string) {
	logger.Log.Error("TSM Internal Loss Measurement Error", "error", msg)
	tsm.finish(msg, false)
}

func (tsm *TSMInternalLoss) check(resp utils.CommandResponse, errMsg string) bool {
	if !resp.Success {
		tsm.setError(fmt.Sprintf("%s: %s", errMsg, resp.ErrorMessage))
		return false
	}
	if tsm.stop {
		tsm.setError("Measurement aborted by user")
		return false
	}
	return true
}

func (tsm *TSMInternalLoss) updateDB(jsonData string, subType string, inputPort, outputPort string) bool {
	var ok bool
	switch subType {
	case "PM_OFFSET":
		ok = resultsDB.UpdateTSMInternalLossPMOffset(jsonData)
	case "CABLE_LOSS":
		ok = resultsDB.UpdateTSMInternalLossCableLoss(jsonData)
	case "PATH_LOSS":
		ok = resultsDB.UpdateTSMInternalLoss(inputPort, outputPort, jsonData)
	}
	if !ok {
		tsm.setError("Failed to update database results")
	}
	return ok
}

// Public API

func (tsm *TSMInternalLoss) Initialize(deviceProfile string, pmChannel string) bool {
	tsm.statusMonitor = make(chan RTStatus, 20)
	tsm.pmChannel = pmChannel
	tsm.currentMeasurement = make(map[string]float64)
	tsm.stop = false

	pmName, okPm := database.GetPMFromDeviceProfile(deviceProfile)
	sgName, okSg := database.GetSGFromDeviceProfile(deviceProfile)
	tsmName, okTsm := database.GetTSMFromDeviceProfile(deviceProfile)

	if !okPm || !okSg || !okTsm {
		return false
	}

	return tsm.pm.LoadDevice(pmName) && tsm.sg.LoadDevice(sgName) && tsm.tsm.LoadDevice(tsmName)
}

func (tsm *TSMInternalLoss) CreateNewTable() bool {
	tsmConfigs, ok := database.GetTSMConfigurations()
	if !ok {
		return false
	}
	tsm.table = make([][]string, 0)
	for _, t := range tsmConfigs {
		var freqs []float64
		cfgs, ok := database.GetConfigNamesForTSMConfig(t)
		if !ok {
			continue
		}

		for _, cfg := range cfgs {
			config, ok := database.GetConfigurationDetails(cfg)
			if !ok {
				continue
			}

			freq, ok := tsm.getFrequenciesForConfig(config)
			if ok && slices.Index(freqs, freq) == -1 {
				freqs = append(freqs, freq)
			}
		}
		tsm.addToMeasurementTable(t, freqs)
	}

	if ok := tsm.createNewDBEntry(); ok {
		logger.Log.Info("TSM Internal Loss Table created")
	} else {
		logger.Log.Error("Error in TSM Internal Loss Table creation")
	}
	return true
}

func (tsm *TSMInternalLoss) getFrequenciesForConfig(config database.Configuration) (float64, bool) {
	if strings.EqualFold(config.ConfigType, "Tx") {
		freq, ok := database.GetTxFrequency(config.TxName.String)
		return float64(freq), ok
	}
	if strings.EqualFold(config.ConfigType, "Rx") {
		freq, ok := database.GetRxFrequency(config.RxName.String)
		return float64(freq), ok
	}
	if strings.EqualFold(config.ConfigType, "PL") {
		freq, ok := database.GetPLFrequency(config.PayloadName.String)
		return freq, ok
	}
	return 0.0, false
}

func (tsm *TSMInternalLoss) addToMeasurementTable(tsmConfig string, frequencies []float64) bool {
	tsmDetails, ok := database.GetTSMPathDetails(tsmConfig)
	if !ok {
		return false
	}

	inputPort := tsmDetails.InputPortName.String
	var outPorts []string
	var paths []string

	if !strings.EqualFold(strings.TrimSpace(tsmDetails.UplinkToSC.String), "") {
		outPorts = append(outPorts, tsmDetails.OutputPortName.String)
		paths = append(paths, tsmDetails.UplinkToSC.String)
		if tsmDetails.IncludePad.Valid {
			outPorts = append(outPorts, tsmDetails.OutputPortName.String+"-WithPad")
			paths = append(paths, tsmDetails.UplinkToSC.String+"!"+tsmDetails.IncludePad.String+";"+tsmDetails.ExcludePad.String)
		}
		if tsmDetails.UplinkToSA.Valid {
			outPorts = append(outPorts, tsmDetails.SAPortName.String)
			paths = append(paths, tsmDetails.UplinkToSA.String)
		}
		if tsmDetails.UplinkToPM.Valid {
			outPorts = append(outPorts, tsmDetails.PMPortName.String)
			paths = append(paths, tsmDetails.UplinkToPM.String)
		}
	}

	if !strings.EqualFold(strings.TrimSpace(tsmDetails.DownlinkToPM.String), "") {
		outPorts = append(outPorts, tsmDetails.PMPortName.String)
		paths = append(paths, tsmDetails.DownlinkToPM.String)
		outPorts = append(outPorts, tsmDetails.OutputPortName.String)
		paths = append(paths, tsmDetails.DownlinkToDC.String)
		if tsmDetails.DownlinkToSA.Valid {
			outPorts = append(outPorts, tsmDetails.SAPortName.String)
			paths = append(paths, tsmDetails.DownlinkToSA.String)
		}
	}

	for i := range outPorts {
		for _, freq := range frequencies {
			freqStr := fmt.Sprintf("%.2f", freq)
			if !tsm.isDuplicate(inputPort, outPorts[i], paths[i], freqStr) {
				tsm.table = append(tsm.table, []string{inputPort, outPorts[i], paths[i], freqStr})
			}
		}
	}
	return true
}

func (tsm *TSMInternalLoss) isDuplicate(inputPort, outputPort, path, frequency string) bool {
	for _, row := range tsm.table {
		if strings.EqualFold(row[0], inputPort) && strings.EqualFold(row[1], outputPort) &&
			strings.EqualFold(row[2], path) && strings.EqualFold(row[3], frequency) {
			return true
		}
	}
	return false
}

func (tsm *TSMInternalLoss) createNewDBEntry() bool {
	consolidatedMap := make(map[string]*cableLossMeasured)
	var freqList []string

	for _, row := range tsm.table {
		if slices.Index(freqList, row[3]) == -1 {
			freqList = append(freqList, row[3])
		}
		key := row[0] + ":" + row[1] + ":" + row[2]
		if _, ok := consolidatedMap[key]; !ok {
			consolidatedMap[key] = &cableLossMeasured{Frequency: []string{}, Measured: []string{}}
		}
		consolidatedMap[key].Frequency = append(consolidatedMap[key].Frequency, row[3])
		consolidatedMap[key].Measured = append(consolidatedMap[key].Measured, "")
	}

	allCables := cableLossMeasured{Frequency: freqList, Measured: make([]string, len(freqList))}
	jsonData, _ := json.MarshalIndent(allCables, "", " ")

	dbEntry := [][]string{
		{"", "", "PM-Measurement", string(jsonData)},
		{"", "", "Cable-Measurement", string(jsonData)},
	}

	for key, value := range consolidatedMap {
		row := strings.Split(key, ":")
		vData, _ := json.MarshalIndent(value, "", " ")
		dbEntry = append(dbEntry, append(row, string(vData)))
	}

	return resultsDB.CreateNewTSMInternalLoss(dbEntry)
}

func (tsm *TSMInternalLoss) measureForFrequencies(frequencies []string, offset map[string]float64) bool {
	if !tsm.check(tsm.pm.SetChAAverageOff(), "PM: ChA average off") {
		return false
	}
	if !tsm.check(tsm.pm.SetChBAverageOff(), "PM: ChB average off") {
		return false
	}
	if !tsm.check(tsm.sg.SetModOff(), "SG: mod off") {
		return false
	}

	defer tsm.pm.SetChAAverageOn()
	defer tsm.pm.SetChBAverageOn()
	defer tsm.sg.SetRFOff()

	for _, freq := range frequencies {
		tsm.notify(fmt.Sprintf("Measuring loss for %s Hz", freq))
		f, _ := strconv.ParseFloat(freq, 64)

		if !tsm.check(tsm.sg.SetFrequency(f), "SG: frequency set") {
			return false
		}
		if !tsm.check(tsm.sg.SetPower(-1*offset[freq]), "SG: power set") {
			return false
		}
		if !tsm.check(tsm.sg.SetRFOn(), "SG: RF on") {
			return false
		}

		var pResp utils.CommandResponse
		if strings.EqualFold(tsm.pmChannel, "A") {
			tsm.pm.SetChannelA(f)
			pResp = tsm.pm.GetPowerChannelA(true)
		} else {
			tsm.pm.SetChannelB(f)
			pResp = tsm.pm.GetPowerChannelB(true)
		}

		if !tsm.check(pResp, "PM: read power") {
			return false
		}

		val := pResp.Result["Power"].Value
		if val < -60 {
			tsm.setError("Power too low (<-60dBm). Check connection.")
			return false
		}
		logger.Log.Info("TSM Internal Loss Measurement Point", "frequency", freq, "measuredPower", val)
		tsm.currentMeasurement[freq] = val
		tsm.sg.SetRFOff()
	}
	return true
}

func (tsm *TSMInternalLoss) measurePMReference(frequencies []string) {
	offset := make(map[string]float64)
	for _, f := range frequencies {
		offset[f] = 0
	}

	if !tsm.measureForFrequencies(frequencies, offset) {
		return
	}

	var m cableLossMeasured
	for _, f := range frequencies {
		m.Frequency = append(m.Frequency, f)
		m.Measured = append(m.Measured, fmt.Sprintf("%.2f", tsm.currentMeasurement[f]))
	}

	data, _ := json.MarshalIndent(m, "", " ")
	if tsm.updateDB(string(data), "PM_OFFSET", "", "") {
		logger.Log.Info("Completed TSM Internal Loss Measurement", "mode", "PM Reference", "success", true)
		tsm.finish("Reference Measurement Completed", true)
	}
}

func (tsm *TSMInternalLoss) measureCableLoss(pmOffset cableLossMeasured) {
	offset := make(map[string]float64)
	for i, f := range pmOffset.Frequency {
		offset[f], _ = strconv.ParseFloat(pmOffset.Measured[i], 64)
	}

	if !tsm.measureForFrequencies(pmOffset.Frequency, offset) {
		return
	}

	var m cableLossMeasured
	for _, f := range pmOffset.Frequency {
		m.Frequency = append(m.Frequency, f)
		m.Measured = append(m.Measured, fmt.Sprintf("%.2f", tsm.currentMeasurement[f]))
	}

	data, _ := json.MarshalIndent(m, "", " ")
	if tsm.updateDB(string(data), "CABLE_LOSS", "", "") {
		logger.Log.Info("Completed TSM Internal Loss Measurement", "mode", "Cable Loss", "success", true)
		tsm.finish("Cable Loss Measurement Completed", true)
	}
}

func (tsm *TSMInternalLoss) measurePathLoss(inputPort, outputPort string) {
	pathStr, measuredJson, ok := resultsDB.GetTSMInternalMeasuredLoss(inputPort, outputPort)
	if !ok {
		tsm.setError("Failed to retrieve path mnemonic")
		return
	}

	cblJson, ok := resultsDB.GetTSMInternalLossCableLoss()
	if !ok {
		tsm.setError("Failed to retrieve cable loss")
		return
	}

	var ref, cbl cableLossMeasured
	if json.Unmarshal([]byte(measuredJson), &ref) != nil || json.Unmarshal([]byte(cblJson), &cbl) != nil {
		tsm.setError("Failed to parse measurement data")
		return
	}

	paths := strings.Split(pathStr, ";")
	if paths[0] != "" {
		if !tsm.check(tsm.tsm.SetDriverStatus(paths[0]), "TSM: set path") {
			return
		}
		time.Sleep(time.Second)
	}

	if !tsm.check(tsm.tsm.SetAttn(1, 0), "TSM: reset attn 1") {
		return
	}
	if !tsm.check(tsm.tsm.SetAttn(2, 0), "TSM: reset attn 2") {
		return
	}
	time.Sleep(time.Second)

	offsets := make(map[string]float64)
	for _, f := range ref.Frequency {
		offsets[f] = 0
	}

	if !tsm.measureForFrequencies(ref.Frequency, offsets) {
		return
	}

	if len(paths) > 1 {
		tsm.tsm.SetDriverStatus(paths[1])
	}

	var m cableLossMeasured
	for _, f := range ref.Frequency {
		m.Frequency = append(m.Frequency, f)
		loss := 0.0
		if idx := slices.Index(cbl.Frequency, f); idx != -1 {
			loss, _ = strconv.ParseFloat(cbl.Measured[idx], 64)
		}
		m.Measured = append(m.Measured, fmt.Sprintf("%.2f", tsm.currentMeasurement[f]-loss))
	}

	data, _ := json.MarshalIndent(m, "", " ")
	if tsm.updateDB(string(data), "PATH_LOSS", inputPort, outputPort) {
		logger.Log.Info("Completed TSM Internal Loss Measurement", "mode", "Path Loss", "success", true)
		tsm.finish("Path Loss Measurement Completed", true)
	}
}

func (tsm *TSMInternalLoss) MeasureForConfig(mode, inputPort, outputPort string) {
	logger.Log.Info("Starting TSM Internal Loss Measurement", "mode", mode, "inputPort", inputPort, "outputPort", outputPort)
	data, ok := resultsDB.GetTSMInternalLossPMOffset()
	if !ok {
		tsm.setError("Unable to read PM offsets from database")
		return
	}

	var pmMeasured cableLossMeasured
	if json.Unmarshal([]byte(data), &pmMeasured) != nil {
		tsm.setError("Failed to parse PM offsets")
		return
	}

	if strings.EqualFold(mode, "PM") {
		tsm.measurePMReference(pmMeasured.Frequency)
	} else if strings.EqualFold(mode, "Cable Loss") {
		tsm.measureCableLoss(pmMeasured)
	} else {
		tsm.measurePathLoss(inputPort, outputPort)
	}
}

func (tsm *TSMInternalLoss) Stop() { tsm.stop = true }

func (tsm *TSMInternalLoss) GetStatusMonitor() chan RTStatus { return tsm.statusMonitor }
