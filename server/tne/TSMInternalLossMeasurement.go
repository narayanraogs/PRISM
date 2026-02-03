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

type TSMInternalLoss struct {
	table              [][]string
	pm                 driver.PM
	sg                 driver.SG
	tsm                driver.TSM
	statusMonitor      chan MeasurementStatus
	pmChannel          string
	stop               bool
	currentMeasurement map[string]float64
}

func (tsm *TSMInternalLoss) Initialize(deviceProfile string, pmChannel string) bool {
	tsm.statusMonitor = make(chan MeasurementStatus, 20)
	tsm.pmChannel = pmChannel
	tsm.currentMeasurement = make(map[string]float64)
	tsm.stop = false
	pmName, ok := database.GetPMFromDeviceProfile(deviceProfile)
	if !ok {
		return false
	}
	sgName, ok := database.GetSGFromDeviceProfile(deviceProfile)
	if !ok {
		return false
	}
	tsmName, ok := database.GetTSMFromDeviceProfile(deviceProfile)
	if !ok {
		return false
	}
	ok = tsm.pm.LoadDevice(pmName)
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
func (tsm *TSMInternalLoss) CreateNewTable() bool {
	tsmConfigs, ok := database.GetTSMConfigurations()
	if !ok {
		return false
	}
	tsm.table = make([][]string, 0)
	var frequencyMap = make(map[string][]float64)
	for _, t := range tsmConfigs {
		var freqs = make([]float64, 0)
		cfgs, ok := database.GetConfigNamesForTSMConfig(t)
		if !ok {
			return false
		}
		for _, cfg := range cfgs {
			config, ok := database.GetConfigurationDetails(cfg)
			if !ok {
				return false
			}

			freq, ok := tsm.getFrequenciesForConfig(config)
			if slices.Index(freqs, freq) == -1 {
				freqs = append(freqs, freq)
			}
		}
		frequencyMap[t] = freqs
		tsm.addToMeasurementTable(t, freqs)
	}
	ok = tsm.createNewDBEntry()
	if ok {
		fmt.Println("TSM Internal Loss Table created")
	} else {
		fmt.Println("Error in TSM Internal Loss Table creation")
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
	outputPorts := make([]string, 0)
	paths := make([]string, 0)
	if !strings.EqualFold(strings.TrimSpace(tsmDetails.UplinkToSC.String), "") {
		outputPorts = append(outputPorts, tsmDetails.OutputPortName.String)
		paths = append(paths, tsmDetails.UplinkToSC.String)
		if tsmDetails.IncludePad.Valid {
			outputPorts = append(outputPorts, tsmDetails.OutputPortName.String+"-WithPad")
			paths = append(paths, tsmDetails.UplinkToSC.String+"!"+tsmDetails.IncludePad.String+";"+tsmDetails.ExcludePad.String)
		}
		if tsmDetails.UplinkToSA.Valid {
			outputPorts = append(outputPorts, tsmDetails.SAPortName.String)
			paths = append(paths, tsmDetails.UplinkToSA.String)
		}
		if tsmDetails.UplinkToPM.Valid {
			outputPorts = append(outputPorts, tsmDetails.PMPortName.String)
			paths = append(paths, tsmDetails.UplinkToPM.String)
		}
	}

	if !strings.EqualFold(strings.TrimSpace(tsmDetails.DownlinkToPM.String), "") {
		outputPorts = append(outputPorts, tsmDetails.PMPortName.String)
		paths = append(paths, tsmDetails.DownlinkToPM.String)
		outputPorts = append(outputPorts, tsmDetails.OutputPortName.String)
		paths = append(paths, tsmDetails.DownlinkToDC.String)
		if tsmDetails.DownlinkToSA.Valid {
			outputPorts = append(outputPorts, tsmDetails.SAPortName.String)
			paths = append(paths, tsmDetails.DownlinkToSA.String)
		}
	}
	for i := 0; i < len(outputPorts); i++ {
		for _, freq := range frequencies {
			freqStr := fmt.Sprintf("%.2f", freq)
			if !tsm.isDuplicate(inputPort, outputPorts[i], paths[i], freqStr) {
				row := make([]string, 0)
				row = append(row, inputPort, outputPorts[i], paths[i], freqStr)
				tsm.table = append(tsm.table, row)
			}
		}
	}

	return true
}

func (tsm *TSMInternalLoss) isDuplicate(inputPort string, outputPort string, path string, frequency string) bool {
	for _, row := range tsm.table {
		if strings.EqualFold(row[0], inputPort) && strings.EqualFold(row[1], outputPort) &&
			strings.EqualFold(row[2], path) && strings.EqualFold(row[3], frequency) {
			return true
		}
	}
	return false
}

func (tsm *TSMInternalLoss) createNewDBEntry() bool {
	var consolidatedMap = make(map[string]*cableLossMeasured)
	var frequencyList = make([]string, 0)
	for _, row := range tsm.table {
		if slices.Index(frequencyList, row[3]) == -1 {
			frequencyList = append(frequencyList, row[3])
		}
		key := row[0] + ":" + row[1] + ":" + row[2]
		val, ok := consolidatedMap[key]
		if !ok {
			consolidatedMap[key] = &cableLossMeasured{
				Frequency: make([]string, 0),
				Measured:  make([]string, 0),
			}
			val = consolidatedMap[key]
		}
		val.Frequency = append(val.Frequency, row[3])
		val.Measured = append(val.Measured, "")
	}
	var dbEntry = make([][]string, 0)
	allCables := cableLossMeasured{
		Frequency: frequencyList,
		Measured:  make([]string, len(frequencyList)),
	}
	jsonData, err := json.MarshalIndent(allCables, "", " ")
	if err != nil {
		return false
	}
	pmRow := []string{"", "", "PM-Measurement", string(jsonData)}
	cableRow := []string{"", "", "Cable-Measurement", string(jsonData)}
	dbEntry = append(dbEntry, pmRow, cableRow)
	for key, value := range consolidatedMap {
		row := strings.Split(key, ":")
		jsonData, err := json.MarshalIndent(&value, "", " ")
		if err != nil {
			return false
		}
		row = append(row, string(jsonData))
		dbEntry = append(dbEntry, row)
	}

	return resultsDB.CreateNewTSMInternalLoss(dbEntry)
}

func (tsm *TSMInternalLoss) measureForFrequencies(frequencies []string, offset map[string]float64) bool {
	response := tsm.pm.SetChAAverageOff()
	if !response.Success {
		tsm.setError("Unable to communicate with PM")
		return false
	}
	response = tsm.pm.SetChBAverageOff()
	if !response.Success {
		tsm.setError("Unable to communicate with PM")
		return false
	}
	response = tsm.sg.SetModOff()
	if !response.Success {
		tsm.setError("Unable to communicate with SG")
		return false
	}
	defer func() {
		tsm.pm.SetChAAverageOn()
		tsm.pm.SetChBAverageOn()
		tsm.sg.SetRFOff()
		fmt.Println("Restored")
	}()

	for _, freq := range frequencies {
		if tsm.stop {
			tsm.setError("User Aborted")
			return false
		}

		var measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Measuring Loss for " + freq + " Hz",
			CurrentStatus: make([][]string, 0),
		}
		tsm.statusMonitor <- measure

		f, _ := strconv.ParseFloat(freq, 64)
		response = tsm.sg.SetFrequency(f)
		if !response.Success {
			tsm.setError("Unable to communicate with SG")
			return false
		}
		power := offset[freq]
		power = power * -1
		response = tsm.sg.SetPower(power)
		if !response.Success {
			tsm.setError("Unable to communicate with SG")
			return false
		}
		response = tsm.sg.SetRFOn()
		if !response.Success {
			tsm.setError("Unable to communicate with SG")
			return false
		}
		if strings.EqualFold(tsm.pmChannel, "A") {
			response = tsm.pm.SetChannelA(f)
			if !response.Success {
				tsm.setError("Unable to communicate with PM")
				return false
			}
			response = tsm.pm.GetPowerChannelA(true)
			if !response.Success {
				tsm.setError("Unable to communicate with PM")
				return false
			}
			if response.Result["ChannelBPower"].Value < -60 {
				tsm.setError("Power read is less than -60dBm. Check PM Connection")
				return false
			}
			tsm.currentMeasurement[freq] = response.Result["Power"].Value
		} else {
			response = tsm.pm.SetChannelB(f)
			if !response.Success {
				tsm.setError("Unable to communicate with PM")
				return false
			}
			response = tsm.pm.GetPowerChannelB(true)
			if !response.Success {
				tsm.setError("Unable to communicate with PM")
				return false
			}
			if response.Result["ChannelBPower"].Value < -60 {
				tsm.setError("Power read is less than -60dBm. Check PM Connection")
				return false
			}
			tsm.currentMeasurement[freq] = response.Result["Power"].Value
		}
		response = tsm.sg.SetRFOff()
		if !response.Success {
			tsm.setError("Unable to communicate with SG")
			return false
		}
	}

	return true
}

func (tsm *TSMInternalLoss) measurePMReference(frequencies []string) {
	offset := make(map[string]float64)
	for _, freq := range frequencies {
		offset[freq] = 0
	}
	ok := tsm.measureForFrequencies(frequencies, offset)
	if !ok {
		return
	}
	var measurement cableLossMeasured
	for freq := range tsm.currentMeasurement {
		measurement.Frequency = append(measurement.Frequency, freq)
		measurement.Measured = append(measurement.Measured, fmt.Sprintf("%.2f", tsm.currentMeasurement[freq]))
	}

	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}
	resultsDB.UpdateTSMInternalLossPMOffset(string(jsonData))

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMInternalLoss) measureCableLoss(pmOffset cableLossMeasured) {
	offset := make(map[string]float64)
	for i, freq := range pmOffset.Frequency {
		offset[freq], _ = strconv.ParseFloat(pmOffset.Measured[i], 64)
	}
	ok := tsm.measureForFrequencies(pmOffset.Frequency, offset)
	if !ok {
		return
	}
	var measurement cableLossMeasured
	for freq := range tsm.currentMeasurement {
		measurement.Frequency = append(measurement.Frequency, freq)
		measurement.Measured = append(measurement.Measured, fmt.Sprintf("%.2f", tsm.currentMeasurement[freq]))
	}

	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}
	resultsDB.UpdateTSMInternalLossCableLoss(string(jsonData))

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMInternalLoss) measurePathLoss(inputPort string, outputPort string) {
	path, measured, ok := resultsDB.GetTSMInternalMeasuredLoss(inputPort, outputPort)
	if !ok {
		tsm.setError("Unable to get Path Mnemonic")
		return
	}
	cblLossString, ok := resultsDB.GetTSMInternalLossCableLoss()
	if !ok {
		tsm.setError("Unable to get Cable Loss")
		return
	}
	var ref cableLossMeasured
	err := json.Unmarshal([]byte(measured), &ref)
	if err != nil {
		tsm.setError("Unable to get Frequencies for Measurement")
		return
	}
	var cblLoss cableLossMeasured
	err = json.Unmarshal([]byte(cblLossString), &cblLoss)
	if err != nil {
		tsm.setError("Unable to get Frequencies for Measurement")
		return
	}
	paths := strings.Split(path, ";")
	if !strings.EqualFold(paths[0], "") {
		response := tsm.tsm.SetDriverStatus(paths[0])
		if !response.Success {
			tsm.setError("Unable to Communicate with TSM")
			return
		}
		time.Sleep(1 * time.Second)
	}
	response := tsm.tsm.SetAttn(1, 0)
	if !response.Success {
		tsm.setError("Unable to Communicate with TSM")
		return
	}
	time.Sleep(1 * time.Second)
	response = tsm.tsm.SetAttn(2, 0)
	if !response.Success {
		tsm.setError("Unable to Communicate with TSM")
		return
	}

	offset := make(map[string]float64)
	for _, freq := range ref.Frequency {
		offset[freq] = 0
	}
	ok = tsm.measureForFrequencies(ref.Frequency, offset)
	if !ok {
		return
	}

	if len(paths) > 1 {
		response := tsm.tsm.SetDriverStatus(paths[1])
		if !response.Success {
			tsm.setError("Unable to Communicate with TSM")
			return
		}
	}

	var measurement cableLossMeasured
	for freq := range tsm.currentMeasurement {
		measurement.Frequency = append(measurement.Frequency, freq)
		index := slices.Index(cblLoss.Frequency, freq)
		loss := 0.0
		if index != -1 {
			loss, _ = strconv.ParseFloat(cblLoss.Measured[index], 64)
		}
		m := tsm.currentMeasurement[freq] - loss
		measurement.Measured = append(measurement.Measured, fmt.Sprintf("%.2f", m))
	}

	jsonData, err := json.MarshalIndent(measurement, "", " ")
	if err != nil {
		return
	}
	resultsDB.UpdateTSMInternalLoss(inputPort, outputPort, string(jsonData))

	var measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Measurement Completed",
		CurrentStatus: make([][]string, 0),
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMInternalLoss) MeasureForConfig(mode string, inputPort string, outputPort string) {
	if strings.EqualFold(mode, "PM") {
		pmOffset, ok := resultsDB.GetTSMInternalLossPMOffset()
		if !ok {
			tsm.setError("Unable to read from Database")
			return
		}
		var pmMeasured cableLossMeasured
		err := json.Unmarshal([]byte(pmOffset), &pmMeasured)
		if err != nil {
			tsm.setError("Unable to read from Database")
			return
		}
		tsm.measurePMReference(pmMeasured.Frequency)
		return
	}
	if strings.EqualFold(mode, "Cable Loss") {
		pmOffset, ok := resultsDB.GetTSMInternalLossPMOffset()
		if !ok {
			tsm.setError("Unable to read from Database")
			return
		}
		var pmMeasured cableLossMeasured
		err := json.Unmarshal([]byte(pmOffset), &pmMeasured)
		if err != nil {
			tsm.setError("Unable to read from Database")
			return
		}
		tsm.measureCableLoss(pmMeasured)
		return
	}
	tsm.measurePathLoss(inputPort, outputPort)
}

func (tsm *TSMInternalLoss) setError(message string) {
	var measure = MeasurementStatus{
		Completed:     true,
		Success:       false,
		Message:       message,
		CurrentStatus: make([][]string, 0),
	}
	tsm.statusMonitor <- measure
	close(tsm.statusMonitor)
}

func (tsm *TSMInternalLoss) Stop() {
	tsm.stop = true
}

func (tsm *TSMInternalLoss) GetStatusMonitor() chan MeasurementStatus {
	return tsm.statusMonitor
}
