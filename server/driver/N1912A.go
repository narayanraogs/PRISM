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

//go:embed instructions/N1912A.csv
var n1912aInstructions string

type n1912a struct {
	connected  bool
	connection instrument
	commands   map[string]utils.Command
}

func (device *n1912a) loadLANDetails(name string) bool {
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

func (device *n1912a) loadCommands() bool {
	inst := readCSV(n1912aInstructions)
	if inst == nil {
		fmt.Println("Unable to read CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *n1912a) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", true, true)
	device.loadCommands()
}

func (device *n1912a) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *n1912a) communicate(cmds []utils.Command, port string, connect bool) []string {
	if connect {
		ok := device.connection.Connect(port)
		if !ok {
			fmt.Println("Connection timeout")
			return nil
		}
	} else {
		if !device.connected {
			ok := device.connection.Connect(port)
			if !ok {
				fmt.Println("Connection timeout")
				return nil
			}
			device.connected = true
		}
	}
	values, err := device.connection.Communicate(cmds)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return values
}

func (device *n1912a) setChannelA(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)

	mnemonics = append(mnemonics, "setUnitdBm", "setChAResolution", "setChAFrequency")
	arguments = append(arguments, "", "3", freqString)

	cmds := device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	time.Sleep(100 * time.Millisecond)
	return getSuccessResponse()
}

func (device *n1912a) getPowerChannelA(connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "getChAFetch")

	cmds := device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))
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

func (device *n1912a) setChannelB(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)

	mnemonics = append(mnemonics, "setUnitdBm", "setChBResolution", "setChBFrequency")
	arguments = append(arguments, "", "3", freqString)

	cmds := device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	time.Sleep(100 * time.Millisecond)
	return getSuccessResponse()
}

func (device *n1912a) getPowerChannelB(connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, "getChBFetch")

	cmds := device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))
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

func (device *n1912a) setChAAverageOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setChAAverageState")
	arguments = append(arguments, "Off")

	cmds := device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *n1912a) setChAAverageOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChAAverageState")
	arguments = append(arguments, "On")

	cmds := device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *n1912a) setChBAverageOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChBAverageState")
	arguments = append(arguments, "Off")

	cmds := device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *n1912a) setChBAverageOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChBAverageState")
	arguments = append(arguments, "On")

	cmds := device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *n1912a) connect() utils.CommandResponse {
	ok := device.connection.Connect("Control")
	if !ok {
		return getErrorResponse("Cannot Communicate with PM")
	}
	device.connected = true
	return getSuccessResponse()
}

func (device *n1912a) disConnect() utils.CommandResponse {
	device.connected = false
	return getSuccessResponse()
}

func (device *n1912a) presetPM() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "reset")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *n1912a) presetPPM() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "reset")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *n1912a) setChannelFrequency(channel string, frequency float64) utils.CommandResponse {
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

func (device *n1912a) getAveragePower(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getPeakPower(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPulseAveragePower", "getPulsePeakPower")

	replacements = []string{"3", "1"}
	if channel == "B" {
		replacements = []string{"4", "2"}
	}
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
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

func (device *n1912a) setPulseParameters(pulseWidth float64, pulsePeriod float64,
	triggerLevel float64, referenceLevel float64, yDiv float64, channel string, preset bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	var index = 0
	if preset {
		mnemonics = append(mnemonics, "setPulseMeasurementMode")
		index = 1
	}
	fmt.Println("YDiv is ", yDiv)
	mnemonics = append(mnemonics, "setTraceScreen", "setVBWOff",
		"setGate1Start", "setGate1Length", "setXStart", "setXScalePerDivision",
		"setYMax", "setYScalePerDivision", "setTriggerSource", "setTriggerLevel",
		//	"selectWindow", "setFullScreenMode")
		"selectWindow")
	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = utils.GetRepeatedArray("", len(mnemonics))
	arguments[0+index] = "TRACE"
	gateStart := (0.1 * pulseWidth) / 1e6
	arguments[2+index] = fmt.Sprintf("%.6f", gateStart)
	gateLength := (0.8 * pulseWidth) / 1e6
	arguments[3+index] = fmt.Sprintf("%.6f", gateLength)
	offset := (-1 * pulseWidth) / 1e6
	arguments[4+index] = fmt.Sprintf("%.6f", offset)
	xScalePerDiv := (pulsePeriod * 0.5) / 1e6
	arguments[5+index] = fmt.Sprintf("%.6f", xScalePerDiv)
	arguments[6+index] = fmt.Sprintf("%.2f", referenceLevel)
	arguments[7+index] = fmt.Sprintf("%.2f", yDiv)
	arguments[9+index] = fmt.Sprintf("%.2f", triggerLevel)

	var cmds = device.getCommands(mnemonics, arguments, replacements)
	for _, cmd := range cmds {
		fmt.Printf("%+v\n", cmd)
	}
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with N1912A")
	}
	return getSuccessResponse()
}

func (device *n1912a) getRiseTime(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getFallTime(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getPulsePeriod(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getPulseOffTime(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getPulseWidth(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getDutyCycle(channel string, connect bool) utils.CommandResponse {
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

func (device *n1912a) getFrequency(channel string, connect bool) utils.CommandResponse {
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
