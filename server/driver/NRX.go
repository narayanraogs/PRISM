package driver

import (
	_ "embed"
	"fmt"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strings"
	"time"
)

//go:embed instructions/NRX.csv
var nrxInstructions string

type nrx struct {
	connection instrument
	connected  bool
	commands   map[string]utils.Command
}

func (device *nrx) loadLANDetails(name string) bool {
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

func (device *nrx) loadCommands() bool {
	inst := readCSV(nrxInstructions)
	if inst == nil {
		logger.Log.Error("Unable to read CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *nrx) getCommandDatabase() map[string]utils.Command {
	return device.commands
}

func (device *nrx) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", true, true)
	device.loadCommands()
}

func (device *nrx) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *nrx) communicate(cmds []utils.Command, port string, connect bool) []string {
	if connect {
		ok := device.connection.Connect(port)
		if !ok {
			logger.Log.Error("Connection timeout")
			return nil
		}
	} else {
		if !device.connected {
			ok := device.connection.Connect(port)
			if !ok {
				logger.Log.Error("Connection timeout")
				return nil
			}
			device.connected = true
		}
	}
	values, err := device.connection.Communicate(cmds)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}
	return values
}

func (device *nrx) setChannelA(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)

	mnemonics = append(mnemonics, "setUnitdBm", "setChAResolution", "setChAFrequency")
	arguments = append(arguments, "", "2", freqString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) getPowerChannelA(connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setChASweep", "getChAFetch")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) setChannelB(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)

	mnemonics = append(mnemonics, "setUnitdBm", "setChBResolution", "setChBFrequency")
	arguments = append(arguments, "", "2", freqString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) getPowerChannelB(connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "setChBSweep", "getChBFetch")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) setChAAverageOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChAAverageState")
	arguments = append(arguments, "Off")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) setChAAverageOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChAAverageState")
	arguments = append(arguments, "On")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) setChBAverageOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChBAverageState")
	arguments = append(arguments, "Off")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) setChBAverageOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChBAverageState")
	arguments = append(arguments, "On")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) presetPM() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "reset")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

//------------------------------Pulse Related Functions----------------------------------------------

func (device *nrx) presetPPM() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "reset")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *nrx) setChannelFrequency(channel string, frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)
	if channel == "A" {
		mnemonics = append(mnemonics, "setFrequencyChA")
	} else {
		mnemonics = append(mnemonics, "setFrequencyChB")
	}
	arguments = append(arguments, freqString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *nrx) getAveragePower(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getAveragePower")

	replacements = append(replacements, "1")
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = append(replacements, "2")
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getPeakPower(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPulseAveragePower", "getPulsePeakPower")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)

	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	time.Sleep(1000 * time.Millisecond)
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["PulseAveragePower"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	response.Result["PulsePeakPower"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[1],
	}
	return response
}

func (device *nrx) setPulseParameters(pulseWidth float64, pulsePeriod float64,
	triggerLevel float64, referenceLevel float64, yDiv float64, channel string, preset bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "setOneScreen", "setPulseMeasurementMode",
		"setPulseMeasurementON", "setTriggerMode", "setTriggerSource", "setTriggerLevel",
		"setStartTime", "setTraceLength", "setPowerReference", "setPowerDiv",
		"displayRiseTime", "displayFallTime", "displayPulsePeriod", "displayPulseOnTime",
		"displayPulseOffTime", "displayDutyCycle", "displayAveragePower",
		"displayPulsePeakPower", "displayPulseAveragePower")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = utils.GetRepeatedArray("", len(mnemonics))
	arguments[5] = fmt.Sprintf("%.2f", triggerLevel) + " dBm"
	offset := (-1 * pulseWidth) / 1e6
	arguments[6] = fmt.Sprintf("%.6f", offset) + " s"
	length := ((pulsePeriod * 5) + pulseWidth) / 1e6
	arguments[7] = fmt.Sprintf("%.6f", length) + " s"
	arguments[8] = fmt.Sprintf("%.2f", referenceLevel) + " dBm"
	arguments[9] = fmt.Sprintf("%.2f", yDiv)
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *nrx) getRiseTime(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getRiseTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["RiseTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getFallTime(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getFallTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["FallTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getPulsePeriod(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPulsePeriod")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["PulsePeriod"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getPulseOffTime(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPulseOffTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["PulseOffTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getPulseWidth(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPulseOnTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["PulseOnTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getDutyCycle(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getDutyCycle")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["DutyCycle"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

func (device *nrx) getFrequency(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getFrequency")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	fValue := utils.GetFloatArray(retVal)
	response := getSuccessResponse()
	response.Result["Frequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}

// ----------------Stability Related Functions-----------------------------------

func (device *nrx) connect() utils.CommandResponse {
	ok := device.connection.Connect("Control")
	if !ok {
		return getErrorResponse("Cannot Communicate with PM")
	}
	device.connected = true
	return getSuccessResponse()
}

func (device *nrx) disConnect() utils.CommandResponse {
	device.connected = false
	return getSuccessResponse()
}
