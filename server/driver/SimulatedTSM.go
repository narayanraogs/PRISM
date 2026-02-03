package driver

import (
	_ "embed"
	"prismServer/utils"
)

type simulatedTSM struct {
	connection instrument
}

func (device *simulatedTSM) loadLANDetails(name string) bool {
	return true
}

func (device *simulatedTSM) loadCommands() bool {
	return true
}

func (device *simulatedTSM) initializeDevice(name string) {
}

func (device *simulatedTSM) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	return make([]utils.Command, 0)
}

func (device *simulatedTSM) communicate(cmds []utils.Command, port string) []string {
	return make([]string, 0)
}

func (device *simulatedTSM) getDriverPath() utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["DriverPath"] = utils.CommandResult{
		ResultType: "String",
		String:     "D1A1234567890!D2A1234567890",
	}
	return response
}

func (device *simulatedTSM) setDriverPath(driverNo int, onStatus string, offStatus string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedTSM) getError() utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["DriverPath"] = utils.CommandResult{
		ResultType: "String",
		String:     "",
	}
	return response
}

func (device *simulatedTSM) setAttn(value float64, attnNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedTSM) getAttn(attnNo int) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Attn"] = utils.CommandResult{
		ResultType: "String",
		String:     "0",
	}
	return response
}
