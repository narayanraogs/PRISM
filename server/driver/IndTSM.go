package driver

import (
	_ "embed"
	"fmt"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
)

//go:embed instructions/INDTSM.csv
var indTSMInstructions string

type indTSM struct {
	connection instrument
	commands   map[string]utils.Command
}

func (device *indTSM) loadLANDetails(name string) bool {
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
	device.connection.Configure("", "\r", true, false)

	return true
}

func (device *indTSM) loadCommands() bool {
	inst := readCSV(indTSMInstructions)
	if inst == nil {
		fmt.Println("Unable to read CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *indTSM) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure("", "\r", true, true)
	device.loadCommands()
}

func (device *indTSM) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *indTSM) communicate(cmds []utils.Command, port string) []string {
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

func (device *indTSM) getDriverPath() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getDriverStatus")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with TSM")
	}
	response := getSuccessResponse()
	response.Result["DriverPath"] = utils.CommandResult{
		ResultType: "String",
		String:     getDriverStatus(retVal[0]),
	}
	return response
}

func (device *indTSM) setDriverPath(driverNo int, onStatus string, offStatus string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)
	var arguments = make([]string, 0)

	var arg string
	if len(onStatus) > 0 {
		arg = arg + "A" + onStatus
	}
	if len(offStatus) > 0 {
		arg = arg + "B" + offStatus
	}
	mnemonics = append(mnemonics, "setDriverStatus")
	arguments = append(arguments, arg)
	replacements = append(replacements, strconv.Itoa(driverNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with TSM")
	}
	return getSuccessResponse()
}

func (device *indTSM) getError() utils.CommandResponse {
	var mnemonics = make([]string, 0)

	mnemonics = append(mnemonics, "getError")
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with TSM")
	}
	response := getSuccessResponse()
	response.Result["DriverPath"] = utils.CommandResult{
		ResultType: "String",
		String:     retVal[0],
	}
	return response
}

func (device *indTSM) setAttn(value float64, attnNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "setAttn")
	arguments = append(arguments, fmt.Sprintf("%.3f", value))
	replacements = append(replacements, strconv.Itoa(attnNo))
	var cmds = device.getCommands(mnemonics, arguments, replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with TSM")
	}
	return getSuccessResponse()
}

func (device *indTSM) getAttn(attnNo int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var replacements = make([]string, 0)

	mnemonics = append(mnemonics, "getAttn")
	replacements = append(replacements, strconv.Itoa(attnNo))
	var cmds = device.getCommands(mnemonics, make([]string, len(mnemonics)), replacements)

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with TSM")
	}
	response := getSuccessResponse()
	response.Result["Attn"] = utils.CommandResult{
		ResultType: "String",
		String:     retVal[0],
	}
	return response
}
