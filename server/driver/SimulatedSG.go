package driver

import (
	_ "embed"
	"prismServer/utils"
)

type simulatedSG struct {
	connection instrument
	commands   map[string]utils.Command
	rfOn       bool
	modOn      bool
	frequency  float64
	power      float64
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
	device.rfOn = false
	device.modOn = false
	device.frequency = 1e9
	device.power = -10
}

func (device *simulatedSG) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	return make([]utils.Command, 0)
}

func (device *simulatedSG) communicate(cmds []utils.Command, port string) []string {
	return make([]string, 0)
}

func (device *simulatedSG) setRFOn() utils.CommandResponse {
	device.rfOn = true
	return getSuccessResponse()
}

func (device *simulatedSG) setRFOff() utils.CommandResponse {
	device.rfOn = false
	return getSuccessResponse()
}

func (device *simulatedSG) setModOn() utils.CommandResponse {
	device.modOn = true
	return getSuccessResponse()
}

func (device *simulatedSG) setModOff() utils.CommandResponse {
	device.modOn = false
	return getSuccessResponse()
}

func (device *simulatedSG) getModulationStatus() utils.CommandResponse {
	response := getSuccessResponse()
	status := "0"
	if device.modOn {
		status = "1"
	}
	response.Result["Modulation"] = utils.CommandResult{
		ResultType: "String",
		String:     status,
	}
	return response
}

func (device *simulatedSG) setFrequency(value float64) utils.CommandResponse {
	device.frequency = value
	return getSuccessResponse()
}

func (device *simulatedSG) setPower(value float64) utils.CommandResponse {
	device.power = value
	return getSuccessResponse()
}
