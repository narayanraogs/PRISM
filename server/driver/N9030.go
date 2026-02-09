package driver

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
)

//go:embed instructions/N9030.csv
var n9030Instructions string

type n9030 struct {
	connection instrument
	commands   map[string]utils.Command
}

func (device *n9030) loadLANDetails(name string) bool {
	d, ok := database.GetDeviceDetails(name)
	if !ok {
		logger.Log.Error("Unable to connect to " + name)
		return false
	}
	device.connection.IPAddress = d.IPAddress
	device.connection.PortNo = int(d.ControlPort)
	device.connection.AlternatePortNo = int(d.AlternateControlPort.Int64)
	device.connection.ReadPortNo = int(d.ReadPort.Int64)
	device.connection.DopplerPortNo = int(d.DopplerPort.Int64)
	device.connection.Timeout = int(d.TimeoutInMillisecs)
	device.connection.Configure(" ", "\n", true, true)
	return true
}

func (device *n9030) loadCommands() bool {
	inst := readCSV(n9030Instructions)
	if inst == nil {
		fmt.Println("Unable to read CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *n9030) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", true, false)
	device.loadCommands()
}

func (device *n9030) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *n9030) getCommandDatabase() map[string]utils.Command {
	return device.commands
}

func (device *n9030) communicate(cmds []utils.Command, port string) []string {
	ok := device.connection.Connect(port)
	if !ok {
		fmt.Println("Connection timeout")
		return nil
	}
	values, err := device.connection.Communicate(cmds)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return values
}

func (device *n9030) setSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	centerFreqString := fmt.Sprintf("%.2f", centerFrequency)
	spanString := fmt.Sprintf("%.2f", span)
	rbwString := fmt.Sprintf("%.2f", rbw)
	vbwString := fmt.Sprintf("%.2f", vbw)

	mnemonics = append(mnemonics, "setCenterFrequency", "setSpan", "setRBW", "setVBW")
	arguments = append(arguments, centerFreqString, spanString, rbwString, vbwString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, cmds[0].Port)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setOBWSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	centerFreqString := fmt.Sprintf("%.2f", centerFrequency)
	spanString := fmt.Sprintf("%.2f", span)
	rbwString := fmt.Sprintf("%.2f", rbw)
	vbwString := fmt.Sprintf("%.2f", vbw)

	mnemonics = append(mnemonics, "setCenterFrequency", "setOBWSpan", "setOBWRBW", "setOBWVBW")
	arguments = append(arguments, centerFreqString, spanString, rbwString, vbwString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getSpectrum() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getCenterFrequency", "getSpan", "getRBW", "getVBW", "getReferenceLevel")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValues := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["CenterFrequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[0],
	}
	ret.Result["Span"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[1],
	}
	ret.Result["RBW"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[2],
	}
	ret.Result["VBW"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[3],
	}
	ret.Result["ReferenceLevel"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[4],
	}
	return ret
}

func (device *n9030) setMarkerNormal(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "setMarkerMode")
	arguments = append(arguments, "POS")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMarkerDelta(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "setMarkerMode")
	arguments = append(arguments, "DELT")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMarkerMaxPeak(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "setMarkerMax")
	arguments = append(arguments, "")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMarkerMinPeak(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "setMarkerMin")
	arguments = append(arguments, "")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getMarkerValue(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getMarkerY", "getMarkerX")
	replacements = append(replacements, strconv.Itoa(markerNo), strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValues := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["MarkerY"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[0],
	}
	ret.Result["MarkerX"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[1],
	}
	return ret
}

func (device *n9030) singleSweep() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setSingleSweep")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) sweepOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setSingleSweep")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) continuousSweep() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setContinuousSweep")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) restartSweep() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepRestart")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setFrequencyCounterMode(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerFrequencyCounterMode", "setMarkerFrequencyCounterResolution", "setCounterGate")
	arguments = append(arguments, "ON", "1", "0.0001")
	replacements = append(replacements, strconv.Itoa(markerNo), strconv.Itoa(markerNo), strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getFrequencyInCounterMode(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "getMarkerFrequency")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)

	retVal := device.communicate(cmds, "Control")

	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValues := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["Frequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[0],
	}
	return ret
}

func (device *n9030) setNormalMode(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerFrequencyCounterMode", "setTraceMode")
	arguments = append(arguments, "OFF", "WRIT")
	replacements = append(replacements, strconv.Itoa(markerNo), "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMaxHold() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setTraceMode")
	arguments = append(arguments, "MAXH")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setClearWrite() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setTraceMode")
	arguments = append(arguments, "WRIT")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMaxHoldOBW() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setTraceModeOBW")
	arguments = append(arguments, "MAXH")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getSweepTime() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getSweepTime")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValues := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["SweepTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[0],
	}
	return ret
}

func (device *n9030) getSweepTimeOBW() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getSweepTimeOBW")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValues := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["SweepTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[0],
	}
	return ret
}

func (device *n9030) setOccupiedBandwidth(bwPercent float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setOccupiedBWMode", "setOccupiedBW")
	arguments = append(arguments, "", strconv.FormatFloat(bwPercent, 'G', 2, 64))
	replacements = append(replacements, "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getOccupiedBW() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getOccupiedBW")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	temp := strings.Split(retVal[0], ",")
	fValues := utils.GetFloatArray(temp)
	ret := getSuccessResponse()
	ret.Result["Bandwidth"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[0],
	}
	ret.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValues[1],
	}
	return ret
}

func (device *n9030) setClearScreen() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setClearScreen")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getOperationStatus() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getOperationWait")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getNoOfRowsToSkipInTrace() utils.CommandResponse {
	var ret = getSuccessResponse()
	ret.Result["NoOfRows"] = utils.CommandResult{
		ResultType: "Integer",
		Integer:    45,
	}
	return ret
}

func (device *n9030) setAutoAlignOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAutoAlign")
	arguments = append(arguments, "OFF")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setAutoAlignOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAutoAlign")
	arguments = append(arguments, "LIGH")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setAllMarkerOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setAllMarkerOff")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) systemPreset() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setSystemPreset")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setYScale(scale float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAmplitudeScale")
	arguments = append(arguments, strconv.FormatFloat(scale, 'G', 2, 64))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setPeakThresholdAndExcursion(threshold float64, excursion float64, markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacement = make([]string, 0)
	var thStr = strconv.FormatFloat(threshold, 'G', 2, 64)
	var exStr = strconv.FormatFloat(excursion, 'G', 2, 64)
	var markerNoStr = strconv.Itoa(markerNo)

	mnemonics = append(mnemonics, "setMarkerPeakThresholdAutoState", "setMarkerPeakThresholdState",
		"setMarkerPeakThreshold", "setMarkerPeakExcursionAutoState", "setMarkerPeakExcursionState", "setMarkerPeakExcursion")
	arguments = append(arguments, "0", "1", thStr, "0", "1", exStr)
	replacement = append(replacement, markerNoStr, markerNoStr, markerNoStr, markerNoStr, markerNoStr, markerNoStr)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) nextPeakLeft(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerLeftPeak")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) nextPeakRight(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerRightPeak")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) shiftMarker(offset float64, markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerX")
	arguments = append(arguments, fmt.Sprintf("%.2fHz", offset))
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMarkerNextPeak(markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerNextPeak")
	arguments = append(arguments, "")
	replacements = append(replacements, strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setReferenceLevel(level float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setReferenceLevel")
	arguments = append(arguments, fmt.Sprintf("%.2f", level))
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getPeakFrequencyDeviationFM(firstMarker string, secondMarker string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerMode", "setMarkerMax", "getMarkerX", "setMarkerMax",
		"setMarkerNextPeak", "getMarkerX")
	arguments = append(arguments, "POS", "", "", "", "", "")
	replacements = append(replacements, firstMarker, firstMarker, firstMarker, secondMarker, secondMarker, secondMarker)
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")

	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValue := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["Frequency1"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	ret.Result["Frequency2"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[1],
	}
	return ret
}

func (device *n9030) setPhaseNoiseMode() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setPhaseNoiseMode", "setPhaseNoiseLogMode", "setDecadeTable")
	arguments = append(arguments, "", "", "")
	replacements = append(replacements, "", "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setSAMode() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setSAMode")
	arguments = append(arguments, "")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setAutoTune() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setAutoTune")
	arguments = append(arguments, "")
	replacements = append(replacements, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setMarkerValuePhaseNoise(value float64, markerNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setMarkerPhaseNoise", "setMarkerXPhaseNoise")
	arguments = append(arguments, "POS", fmt.Sprintf("%.2f", value))
	replacements = append(replacements, strconv.Itoa(markerNo), strconv.Itoa(markerNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) setPhaseNoiseMeasurement(startOffset float64, stopOffset float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setSignalTrackingOff", "setStartOffsetPhaseNoise", "setStopOffsetPhaseNoise")
	arguments = append(arguments, "0", fmt.Sprintf("%.2f", startOffset),
		fmt.Sprintf("%.2f", stopOffset))
	replacements = append(replacements, "", "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	return getSuccessResponse()
}

func (device *n9030) getPhaseNoiseMarkerY(marker int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "getMarkerYPhaseNoise")
	arguments = append(arguments, "")
	replacements = append(replacements, strconv.Itoa(marker))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}
	fValue := utils.GetFloatArray(retVal)
	ret := getSuccessResponse()
	ret.Result["MarkerY"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return ret
}

func (device *n9030) getSpectrumDump() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setClearScreen", "setFullScreen", "setMonochromeBackground", "setScreenDump", "getScreenDump")
	arguments = append(arguments, "", "", "", "\"D:\\temp.png\"", "\"D:\\temp.png\"")
	replacements = append(replacements, "", "", "", "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}

	data, _ := base64.StdEncoding.DecodeString(retVal[0])
	sep := []byte{0x50, 0x4E, 0x47}
	var index = bytes.Index(data, sep)
	if index == -1 {
		return getErrorResponse("Sprectum Dump is Not PNG")
	}
	index = index - 1
	data = data[index:]

	crop := utils.CropImage(0, 45, 0, 45, data)
	if crop == nil {
		return getErrorResponse("Unable to Crop Image")
	}

	filename := utils.GetTimeStampedFileName("screenshot")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".png"
	_ = os.WriteFile(filename, crop, os.ModePerm)

	var encodedImage = base64.StdEncoding.EncodeToString(crop)
	ret := getSuccessResponse()
	ret.Result["SpectrumDump"] = utils.CommandResult{
		ResultType: "Image",
		String:     encodedImage,
	}
	return ret
}

func (device *n9030) getTraceDump(points int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setNoOfTracePoints", "setSaveTrace", "getScreenDump")
	arguments = append(arguments, strconv.Itoa(points), "TRACE1, 'D:\\VALUES.csv'", "'D:\\VALUES.csv'")
	replacements = append(replacements, "", "", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SA")
	}

	data, _ := base64.StdEncoding.DecodeString(retVal[0])
	filename := utils.GetTimeStampedFileName("trace")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".csv"
	_ = os.WriteFile(filename, data, os.ModePerm)

	fileData := string(data)
	ret := getSuccessResponse()
	ret.Result["TraceDump"] = utils.CommandResult{
		ResultType: "String",
		String:     fileData,
	}
	return ret
}
