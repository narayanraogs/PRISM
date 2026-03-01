package driver

import (
	_ "embed"
	"fmt"
	"prismServer/utils"
	"strings"
)

type simulatedTSM struct {
	connection   instrument
	driverStates map[int]string
	attnStates   map[int]float64
}

func (device *simulatedTSM) loadLANDetails(name string) bool {
	return true
}

func (device *simulatedTSM) loadCommands() bool {
	return true
}

func (device *simulatedTSM) initializeDevice(name string) {
	device.driverStates = make(map[int]string)
	device.attnStates = make(map[int]float64)
	for i := 1; i <= utils.Config.TSM.NoOfDrivers; i++ {
		device.driverStates[i] = "A1234567890"
	}
}

func (device *simulatedTSM) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	return make([]utils.Command, 0)
}

func (device *simulatedTSM) communicate(cmds []utils.Command, port string) []string {
	return make([]string, 0)
}

func (device *simulatedTSM) getDriverPath() utils.CommandResponse {
	response := getSuccessResponse()
	var tbr = make([]string, 0)
	for i := 1; i <= utils.Config.TSM.NoOfDrivers; i++ {
		tbr = append(tbr, fmt.Sprintf("D%d%s", i, device.driverStates[i]))
	}
	response.Result["DriverPath"] = utils.CommandResult{
		ResultType: "String",
		String:     strings.Join(tbr, "!"),
	}
	return response
}

func (device *simulatedTSM) setDriverPath(driverNo int, onStatus string, offStatus string) utils.CommandResponse {
	currentState := device.driverStates[driverNo]
	if len(currentState) == 0 {
		currentState = "A1234567890"
	}
	//todo: Check
	chars := []rune(currentState)
	// Apply "A" (On) values
	for _, posStr := range strings.Split(onStatus, "") {
		var pos int
		if _, err := fmt.Sscanf(posStr, "%d", &pos); err == nil && pos >= 0 && pos < len(chars) {
			chars[pos] = '1'
		}
	}
	// Apply "B" (Off) values
	for _, posStr := range strings.Split(offStatus, "") {
		var pos int
		if _, err := fmt.Sscanf(posStr, "%d", &pos); err == nil && pos >= 0 && pos < len(chars) {
			chars[pos] = '0'
		}
	}

	device.driverStates[driverNo] = string(chars)
	return getSuccessResponse()
}

func (device *simulatedTSM) getError() utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Error"] = utils.CommandResult{
		ResultType: "String",
		String:     "No Error",
	}
	return response
}

func (device *simulatedTSM) setAttn(value float64, attnNo int) utils.CommandResponse {
	device.attnStates[attnNo] = value
	return getSuccessResponse()
}

func (device *simulatedTSM) getAttn(attnNo int) utils.CommandResponse {
	response := getSuccessResponse()
	val := device.attnStates[attnNo]
	response.Result["Attn"] = utils.CommandResult{
		ResultType: "String",
		String:     fmt.Sprintf("%.3f", val),
	}
	return response
}
