package driver

import (
	_ "embed"
	"prismServer/utils"
)

type simulatedSG struct {
	connection instrument
	commands   map[string]utils.Command
}

func (device *simulatedSG) loadLANDetails(name string) bool {
	return true
}

func (device *simulatedSG) getCommandDatabase() map[string]utils.Command {
	return device.commands
}

func (device *simulatedSG) loadCommands() bool {
	return true
}

func (device *simulatedSG) initializeDevice(name string) {
	device.commands = make(map[string]utils.Command)
}

func (device *simulatedSG) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	return make([]utils.Command, 0)
}

func (device *simulatedSG) communicate(cmds []utils.Command, port string) []string {
	return make([]string, 0)
}

func (device *simulatedSG) setRFOn() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSG) setRFOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSG) setModOn() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSG) setModOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSG) getModulationStatus() utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Modulation"] = utils.CommandResult{
		ResultType: "String",
		String:     "OFF",
	}
	return response
}

func (device *simulatedSG) setFrequency(value float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSG) setPower(value float64) utils.CommandResponse {
	return getSuccessResponse()
}
