package driver

import (
	"prismServer/utils"
)

type simulatedPM struct {
	connection instrument
	commands   map[string]utils.Command
}

func (device *simulatedPM) loadLANDetails(name string) bool {
	return true
}

func (device *simulatedPM) loadCommands() bool {
	device.commands = make(map[string]utils.Command)
	return true
}

func (device *simulatedPM) getCommandDatabase() map[string]utils.Command {
	return device.commands
}

func (device *simulatedPM) initializeDevice(name string) {
}

func (device *simulatedPM) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	return make([]utils.Command, 0)
}

func (device *simulatedPM) communicate(cmds []utils.Command, port string, connect bool) []string {
	return make([]string, 0)
}

func (device *simulatedPM) setChannelA(frequency float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) getPowerChannelA(connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	return response
}

func (device *simulatedPM) getFrequency(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Frequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      1e9,
	}
	return response
}

func (device *simulatedPM) setChannelB(frequency float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) getPowerChannelB(connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	return response
}

func (device *simulatedPM) setChAAverageOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) setChAAverageOn() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) setChBAverageOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) setChBAverageOn() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) presetPM() utils.CommandResponse {
	return getSuccessResponse()
}

//------------------------------Pulse Related Functions----------------------------------------------

func (device *simulatedPM) presetPPM() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) setChannelFrequency(channel string, frequency float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) getAveragePower(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	return response
}

func (device *simulatedPM) getPeakPower(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["PulseAveragePower"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	response.Result["PulsePeakPower"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	return response
}

func (device *simulatedPM) setPulseParameters(pulseWidth float64, pulsePeriod float64,
	triggerLevel float64, referenceLevel float64, yDiv float64, channel string, preset bool) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) getRiseTime(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["RiseTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      10e-9,
	}
	return response
}

func (device *simulatedPM) getFallTime(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["FallTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      10e-9,
	}
	return response
}

func (device *simulatedPM) getPulsePeriod(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["PulsePeriod"] = utils.CommandResult{
		ResultType: "Value",
		Value:      200e-6,
	}
	return response
}

func (device *simulatedPM) getPulseOffTime(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["PulseOffTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      150e-9,
	}
	return response
}

func (device *simulatedPM) getPulseWidth(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["PulseOnTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      50e-6,
	}
	return response
}

func (device *simulatedPM) getDutyCycle(channel string, connect bool) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["DutyCycle"] = utils.CommandResult{
		ResultType: "Value",
		Value:      25,
	}
	return response
}

func (device *simulatedPM) connect() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedPM) disConnect() utils.CommandResponse {
	return getSuccessResponse()
}
