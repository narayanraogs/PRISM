package driver

import (
	_ "embed"
	"fmt"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strings"
)

//go:embed instructions/ML2488B.csv
var ml2488bInstructions string

type ml2488b struct {
	connection instrument
	connected  bool
	commands   map[string]utils.Command
}

func (device *ml2488b) loadLANDetails(name string) bool {
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

func (device *ml2488b) loadCommands() bool {
	inst := readCSV(ml2488bInstructions)
	if inst == nil {
		fmt.Println("Unable to read CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *ml2488b) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", true, true)
	device.loadCommands()
}

func (device *ml2488b) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *ml2488b) communicate(cmds []utils.Command, port string, connect bool) []string {
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

func (device *ml2488b) setChannelA(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)

	mnemonics = append(mnemonics, "setUnitdBm", "setResolutionChA", "setFrequencyChA")
	arguments = append(arguments, "", "2", freqString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) getPowerChannelA(connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getChannelAPower")
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

func (device *ml2488b) setChannelB(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	freqString := fmt.Sprintf("%.2f", frequency)

	mnemonics = append(mnemonics, "setUnitdBm", "setResolutionChB", "setFrequencyChB")
	arguments = append(arguments, "", "2", freqString)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) getPowerChannelB(connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getChannelBPower")
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

func (device *ml2488b) setChAAverageOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAverageStateChA")
	arguments = append(arguments, "Off")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) setChAAverageOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAverageStateChA")
	arguments = append(arguments, "On")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) setChBAverageOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAverageStateChB")
	arguments = append(arguments, "Off")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) setChBAverageOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAverageStateChB")
	arguments = append(arguments, "On")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) connect() utils.CommandResponse {
	ok := device.connection.Connect("Control")
	if !ok {
		return getErrorResponse("Cannot Communicate with PM")
	}
	device.connected = true
	return getSuccessResponse()
}

func (device *ml2488b) disConnect() utils.CommandResponse {
	device.connected = false
	return getSuccessResponse()
}

func (device *ml2488b) presetPM() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "reset", "setContMeasurementOn")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

//------------------------------Pulse Related Functions----------------------------------------------

func (device *ml2488b) presetPPM() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "reset", "setContMeasurementOn")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) setChannelFrequency(channel string, frequency float64) utils.CommandResponse {
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
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) getAveragePower(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPowerCW")

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

// returns peak and pulse average
func (device *ml2488b) getPeakPower(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "getPulsePower")

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
	ret := retVal[0]
	values := strings.Split(ret, ",")
	fValue := utils.GetFloatArray(values)
	response := getSuccessResponse()
	response.Result["PulseAveragePower"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[5],
	}
	response.Result["PulsePeakPower"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[7],
	}
	fmt.Println(response.Result["PulsePeakPower"].Value)
	return response

}

func (device *ml2488b) setPulseParameters(pulseWidth float64, pulsePeriod float64, _ float64, _ float64, _ float64, channel string, preset bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)
	mnemonics = append(mnemonics, "setSensorConfig", "setActiveChannel", "setPulseMeasurementMode",
		"setDisplayFormat", "setCaptureTime", "setOneScreen", "setInternalTrigger",
		"setTriggerDelay", "setAutoScale", "setGateReference1", "setGateReference2",
		"setRepeatGates", "setGateRepeatOffset", "setGateRepeatTime", "setAverageOn",
		"setLowerMarkerVal", "setUpperMarkerVal", "setSourceMarker",
		"setAllMarkerOff", "setActiveMarker", "setMarkerOn")

	arguments = utils.GetRepeatedArray("", len(mnemonics))
	arguments[0] = "A"
	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
		arguments[0] = "B"
	}

	captureTime := ((pulsePeriod * 5) + pulseWidth) / 1e6

	arguments[4] = fmt.Sprintf("%.6f", captureTime) + " s"
	offset := (-1 * pulseWidth) / 1e6
	arguments[7] = fmt.Sprintf("%.6f", offset) + " s"
	arguments[9] = "0 us"
	arguments[10] = fmt.Sprintf("%.2f", pulseWidth) + " us"
	arguments[13] = fmt.Sprintf("%.2f", pulsePeriod) + " us"
	arguments[15] = "10 %"
	arguments[16] = "90 %"
	arguments[20] = "ON"
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", true)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PPM")
	}
	return getSuccessResponse()
}

func (device *ml2488b) getRiseTime(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setLowerMarkerVal", "setUpperMarkerVal", "getRiseTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = append(arguments, "10 %", "90 %", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	var temp = make([]string, 0)
	for _, val := range retVal {
		temp = append(temp, strings.Split(val, ",")[1])
	}
	fValue := utils.GetFloatArray(temp)
	response := getSuccessResponse()
	response.Result["RiseTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	fmt.Println("RiseTime...", response.Result["RiseTime"].Value)
	return response
}

func (device *ml2488b) getFallTime(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setLowerMarkerVal", "setUpperMarkerVal", "getFallTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = append(arguments, "10 %", "90 %", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	var temp = make([]string, 0)
	for _, val := range retVal {
		temp = append(temp, strings.Split(val, ",")[1])
	}
	fValue := utils.GetFloatArray(temp)
	response := getSuccessResponse()
	response.Result["FallTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	fmt.Println("FallTime...", response.Result["FallTime"].Value)
	return response
}

func (device *ml2488b) getPulsePeriod(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setLowerMarkerVal", "setUpperMarkerVal", "getPulsePeriod")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = append(arguments, "50 %", "50 %", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	var temp = make([]string, 0)
	for _, val := range retVal {
		temp = append(temp, strings.Split(val, ",")[1])
	}
	fValue := utils.GetFloatArray(temp)
	response := getSuccessResponse()
	response.Result["PulsePeriod"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0] * 1e6,
	}
	return response
}

func (device *ml2488b) getPulseOffTime(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setLowerMarkerVal", "setUpperMarkerVal", "getPulseOffTime")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = append(arguments, "50 %", "50 %", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	var temp = make([]string, 0)
	for _, val := range retVal {
		temp = append(temp, strings.Split(val, ",")[1])
	}
	fValue := utils.GetFloatArray(temp)
	response := getSuccessResponse()
	response.Result["PulseOffTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0] * 1e6,
	}
	return response
}

func (device *ml2488b) getPulseWidth(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setLowerMarkerVal", "setUpperMarkerVal", "getPulseWidth")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = append(arguments, "50 %", "50 %", "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	var temp = make([]string, 0)
	for _, val := range retVal {
		temp = append(temp, strings.Split(val, ",")[1])
	}
	fValue := utils.GetFloatArray(temp)
	response := getSuccessResponse()
	response.Result["PulseOnTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0] * 1e6,
	}
	return response
}

func (device *ml2488b) getDutyCycle(channel string, connect bool) utils.CommandResponse {
	onTime := device.getPulseWidth(channel, connect).Result["PulseOnTime"].Value
	totalTime := device.getPulsePeriod(channel, connect).Result["PulsePeriod"].Value
	dutyCycle := (onTime / totalTime) * 100
	response := getSuccessResponse()
	response.Result["DutyCycle"] = utils.CommandResult{
		ResultType: "Value",
		Value:      dutyCycle,
	}
	fmt.Println("DutyCycle...", response.Result["DutyCycle"].Value)
	return response
}

func (device *ml2488b) getFrequency(channel string, connect bool) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "getFrequency")

	replacements = utils.GetRepeatedArray("1", len(mnemonics))
	if channel == "B" {
		replacements = make([]string, 0)
		replacements = utils.GetRepeatedArray("2", len(mnemonics))
	}
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, replacements)
	retVal := device.communicate(cmds, "Control", connect)
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with PM")
	}
	var temp = make([]string, 0)
	for _, val := range retVal {
		temp = append(temp, strings.Split(val, ",")[1])
	}
	fValue := utils.GetFloatArray(temp)
	response := getSuccessResponse()
	response.Result["Frequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      fValue[0],
	}
	return response
}
