package driver

import (
	_ "embed"
	"fmt"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strings"
)

//go:embed instructions/E82x7D.csv
var e82x7dInstructions string

type e82x7d struct {
	connection instrument
	commands   map[string]utils.Command
}

func (device *e82x7d) loadLANDetails(name string) bool {
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
	device.connection.Configure(" ", "\n", true, false)
	return true
}

func (device *e82x7d) loadCommands() bool {
	inst := readCSV(e82x7dInstructions)
	if inst == nil {
		fmt.Println("Unable to read CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *e82x7d) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", true, true)
	device.loadCommands()
}

func (device *e82x7d) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *e82x7d) communicate(cmds []utils.Command, port string) []string {
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

func (device *e82x7d) setRFOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setAlcOn", "setRFOn")
	arguments = append(arguments, "", "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	return getSuccessResponse()
}

func (device *e82x7d) setRFOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setRFOff")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	return getSuccessResponse()
}

func (device *e82x7d) setModOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setModOn")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	return getSuccessResponse()
}

func (device *e82x7d) setModOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setModOff")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	return getSuccessResponse()
}

func (device *e82x7d) getModulationStatus() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getModState")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	response := getSuccessResponse()
	response.Result["Modulation"] = utils.CommandResult{
		ResultType: "String",
		String:     retVal[0],
	}
	return response
}

func (device *e82x7d) setFrequency(value float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setFrequency")
	arguments = append(arguments, fmt.Sprintf("%.2f", value))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	return getSuccessResponse()
}

func (device *e82x7d) setPower(value float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setPower")
	arguments = append(arguments, fmt.Sprintf("%.2f", value))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	retVal := device.communicate(cmds, "Control")
	if retVal == nil {
		return getErrorResponse("Cannot Communicate with SG")
	}
	return getSuccessResponse()
}
