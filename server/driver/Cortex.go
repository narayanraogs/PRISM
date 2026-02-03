package driver

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
)

//go:embed instructions/CortexCommands.csv
var cortexInstructions string

//go:embed instructions/CortexComponents.csv
var cortexComponents string

type cortex struct {
	connection instrument
	commands   map[string]utils.Command
	components map[string]utils.Component
}

func (device *cortex) loadLANDetails(name string) bool {
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
	device.connection.Configure(" ", "\n", false, false)
	return true
}

func (device *cortex) loadCommands() bool {
	inst := readCSV(cortexInstructions)
	if inst == nil {
		fmt.Println("Unable to read Instructions CSV")
		return false
	}
	comp := readComponents(cortexComponents)
	if comp == nil {
		fmt.Println("Unable to read Components CSV")
		return false
	}
	device.commands = inst
	device.components = comp
	return true
}

func (device *cortex) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", false, true)
	device.loadCommands()
}

func (device *cortex) setCarrierOn(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setIFCarrier")
	arguments = append(arguments, "1")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setCarrierOff(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setIFCarrier")
	arguments = append(arguments, "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setModulationOn(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setModulation")
	arguments = append(arguments, "1")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setModulationOff(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setModulation")
	arguments = append(arguments, "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setFrequencyDeviationTC(component string, deviation float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setPeakFrequencyDeviationTCU")
	arguments = append(arguments, fmt.Sprintf("%.2f", deviation))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setFrequencyDeviationTone(component string, deviation float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setPeakFrequencyDeviationTMS1")
	arguments = append(arguments, fmt.Sprintf("%.2f", deviation))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setModIndexTC(component string, modIndex float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setModIndexTCU")
	arguments = append(arguments, fmt.Sprintf("%.2f", modIndex))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setModIndexTone(component string, modIndex float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setModIndexTMS1")
	arguments = append(arguments, fmt.Sprintf("%.2f", modIndex))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setRangingToneFrequency(frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSubCarrierFrequency")
	arguments = append(arguments, fmt.Sprintf("%.2f", frequency))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["TMS-1"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setOnlyTC(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var argumentValue = 0x80000200

	mnemonics = append(mnemonics, "setModulationSignal")
	arguments = append(arguments, strconv.Itoa(argumentValue))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setOnlyRanging(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var argumentValue = 0x80000800

	mnemonics = append(mnemonics, "setModulationSignal")
	arguments = append(arguments, strconv.Itoa(argumentValue))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setTCAndRanging(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	var argumentValue = 0x80000A00

	mnemonics = append(mnemonics, "setModulationSignal")
	arguments = append(arguments, strconv.Itoa(argumentValue))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setSweepRate(component string, sweepRate float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepRate")
	arguments = append(arguments, fmt.Sprintf("%.2f", sweepRate))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setSweepStep(component string, sweepStep float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepStep")
	arguments = append(arguments, fmt.Sprintf("%.2f", sweepStep))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) triggerSweep(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepState")
	arguments = append(arguments, "6")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) sweepHold(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepState")
	arguments = append(arguments, "6")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) sweepContinuous(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepState")
	arguments = append(arguments, "4")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) enableDoppler(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepState")
	arguments = append(arguments, "256")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) startSweep(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepState")
	arguments = append(arguments, "4")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) stopSweep(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepState")
	arguments = append(arguments, "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setSweepRange(component string, sweepRange float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setSweepRange")
	arguments = append(arguments, fmt.Sprintf("%.2f", sweepRange))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setFrequency(component string, frequency float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setCarrierFrequency")
	arguments = append(arguments, fmt.Sprintf("%.2f", frequency))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setPower(component string, power float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setIFCarrierLevel")
	arguments = append(arguments, fmt.Sprintf("%.2f", power))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setChipRate(component string, chipRate float64) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setChipRate")
	arguments = append(arguments, fmt.Sprintf("%.2f", chipRate))
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) checkConnection() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setBitRateLowBW")
	arguments = append(arguments, "1000")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["TMS-1"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setDopplerCompensationEnable() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setDopplerCompensation")
	arguments = append(arguments, "1")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["TCU"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setDopplerCompensationDisable() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "setDopplerCompensation")
	arguments = append(arguments, "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["TCU"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getCommandPacket(cmds[0], comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) getDeviceTime() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getTimeStamp")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["Global"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getReadPacket(comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Read")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	readPacket, _ := base64.StdEncoding.DecodeString(retVal[0])
	tbr := device.getValue(readPacket, cmds[0].Command)
	response := getSuccessResponse()
	response.Result["Time"] = utils.CommandResult{
		ResultType: "Integer",
		Integer:    tbr,
	}
	return response
}

func (device *cortex) getDopplerCompensationStatus(component string) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getCurrentDopplerStatus")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components[component]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getReadPacket(comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Read")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	readPacket, _ := base64.StdEncoding.DecodeString(retVal[0])
	tbr := device.getValue(readPacket, cmds[0].Command)
	response := getSuccessResponse()
	response.Result["Doppler"] = utils.CommandResult{
		ResultType: "Integer",
		Integer:    tbr,
	}
	return response
}

func (device *cortex) setIdlePatternOn() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getIdlePattern")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["TCU"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getReadPacket(comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Read")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	readPacket, _ := base64.StdEncoding.DecodeString(retVal[0])
	tbr := device.getValue(readPacket, cmds[0].Command)

	if tbr != 0 {
		return getSuccessResponse()
	}

	packet = device.getToggleIdlePatternPacket()
	cmds[0].Packet = packet
	retVal = device.communicate(cmds, "Control")
	ack, ok = device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *cortex) setIdlePatternOff() utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getIdlePattern")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	comp, ok := device.components["TCU"]
	if !ok {
		return getErrorResponse("Component name not proper")
	}
	packet := device.getReadPacket(comp)
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Read")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	readPacket, _ := base64.StdEncoding.DecodeString(retVal[0])
	tbr := device.getValue(readPacket, cmds[0].Command)

	if tbr == 0 {
		return getSuccessResponse()
	}

	packet = device.getToggleIdlePatternPacket()
	cmds[0].Packet = packet
	retVal = device.communicate(cmds, "Control")
	ack, ok = device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

/*func (device *cortex) setDopplerCompensationTable(timeOffset int, frequencies []int, times []int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getIdlePattern")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	response := device.getDeviceTime()
	if !response.Success {
		return response
	}

	sequence := 0
	deviceTime := response.Result["Time"].Integer
	requiredTime := deviceTime + timeOffset
	completeTimes := times
	completeFrequencies := frequencies
	tableAddress := 1
	repeat := true
	for repeat {
		var tempTime = make([]int, 0)
		var tempFreq = make([]int, 0)
		if sequence > 0 {
			tableAddress = 2
		}
		if len(completeTimes) > 4000 {
			tempTime = append(tempTime, completeTimes[:4000]...)
			tempFreq = append(tempFreq, completeFrequencies[:4000]...)
			completeTimes = completeTimes[4000:]
			completeFrequencies = completeFrequencies[4000:]
		} else {
			tempTime = append(tempTime, completeTimes...)
			tempFreq = append(tempFreq, completeFrequencies...)
			repeat = false
		}
		packet := make([]byte, 0)
		packet = append(packet, utils.GetByteArrayForInt(1234567890)...)
		packet = append(packet, utils.GetByteArrayForInt(0)...)
		packet = append(packet, utils.GetByteArrayForInt(1)...)
		packet = append(packet, utils.GetByteArrayForInt(tableAddress)...)
		packet = append(packet, utils.GetByteArrayForInt(requiredTime)...)
		packet = append(packet, utils.GetByteArrayForInt(0)...)
		packet = append(packet, utils.GetByteArrayForInt(sequence)...)
		for i := 0; i < 9; i++ {
			packet = append(packet, utils.GetByteArrayForInt(0)...)
		}
		packet = append(packet, utils.GetByteArrayForInt(len(tempFreq))...)
		for i := 0; i < len(tempFreq); i++ {
			packet = append(packet, utils.GetByteArrayForInt(tempTime[i])...)
			packet = append(packet, utils.GetByteArrayForInt(tempFreq[i])...)
		}
		packet = append(packet, utils.GetByteArrayForInt(-1234567890)...)
		length := len(packet)
		lengthBytes := utils.GetByteArrayForInt(length)
		for i := 0; i < 4; i++ {
			packet[4+i] = lengthBytes[i]
		}
		cmds[0].Packet = packet
		retVal := device.communicate(cmds, "Doppler")
		if retVal == nil {
			return getErrorResponse("Unable to communicate with Doppler Port")
		}
		sequence = sequence + 1
	}
	return getSuccessResponse()
}*/

func (device *cortex) setDopplerCompensationTable(timeOffset int, frequencies []int, extendedFrequencies []int, times []int) utils.CommandResponse {
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)

	mnemonics = append(mnemonics, "getIdlePattern")
	arguments = append(arguments, "")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))

	response := device.getDeviceTime()
	if !response.Success {
		return response
	}

	sequence := 0
	deviceTime := response.Result["Time"].Integer
	requiredTime := deviceTime + timeOffset
	completeTimes := times
	completeFrequencies := frequencies
	completeExtendedFrequencies := extendedFrequencies
	tableAddress := 0x101
	repeat := true
	for repeat {
		var tempTime = make([]int, 0)
		var tempFreq = make([]int, 0)
		var tempExtFreq = make([]int, 0)
		if sequence > 0 {
			tableAddress = 0x102
		}
		if len(completeTimes) > 4000 {
			tempTime = append(tempTime, completeTimes[:4000]...)
			tempFreq = append(tempFreq, completeFrequencies[:4000]...)
			tempExtFreq = append(tempExtFreq, completeExtendedFrequencies[:4000]...)
			completeTimes = completeTimes[4000:]
			completeFrequencies = completeFrequencies[4000:]
			completeExtendedFrequencies = completeExtendedFrequencies[4000:]
		} else {
			tempTime = append(tempTime, completeTimes...)
			tempFreq = append(tempFreq, completeFrequencies...)
			tempExtFreq = append(tempExtFreq, completeExtendedFrequencies...)
			repeat = false
		}
		packet := make([]byte, 0)
		packet = append(packet, utils.GetByteArrayForInt(1234567890)...)
		packet = append(packet, utils.GetByteArrayForInt(0)...)
		packet = append(packet, utils.GetByteArrayForInt(1)...)
		packet = append(packet, utils.GetByteArrayForInt(tableAddress)...)
		packet = append(packet, utils.GetByteArrayForInt(requiredTime)...)
		packet = append(packet, utils.GetByteArrayForInt(0)...)
		packet = append(packet, utils.GetByteArrayForInt(sequence)...)
		packet = append(packet, utils.GetByteArrayForInt(0)...)
		if sequence == 0 {
			packet = append(packet, utils.GetByteArrayForInt(0x1040)...)
			packet = append(packet, utils.GetByteArrayForFloat(1)...)
			packet = append(packet, utils.GetByteArrayForFloat(0)...)
			packet = append(packet, utils.GetByteArrayForFloat(0)...)
			packet = append(packet, utils.GetByteArrayForInt(0x1010)...)
			packet = append(packet, utils.GetByteArrayForFloat(0)...)
			packet = append(packet, utils.GetByteArrayForFloat(utils.Config.TestRelated.ChipDopplerRate)...)
			packet = append(packet, utils.GetByteArrayForFloat(0)...)
		} else {
			for i := 0; i < 8; i++ {
				packet = append(packet, utils.GetByteArrayForInt(0)...)
			}
		}
		for i := 0; i < 25; i++ {
			packet = append(packet, utils.GetByteArrayForInt(0)...)
		}
		packet = append(packet, utils.GetByteArrayForInt(len(tempFreq))...)

		for i := 0; i < len(tempFreq); i++ {
			packet = append(packet, utils.GetByteArrayForInt(tempTime[i])...)
			packet = append(packet, utils.GetByteArrayForInt(tempFreq[i])...)
			packet = append(packet, utils.GetByteArrayForInt(tempExtFreq[i])...)
		}
		packet = append(packet, utils.GetByteArrayForInt(-1234567890)...)
		length := len(packet)
		lengthBytes := utils.GetByteArrayForInt(length)
		for i := 0; i < 4; i++ {
			packet[4+i] = lengthBytes[i]
		}
		cmds[0].Packet = packet
		retVal := device.communicate(cmds, "Doppler")
		if retVal == nil {
			return getErrorResponse("Unable to communicate with Doppler Port")
		}
		sequence = sequence + 1
	}
	return getSuccessResponse()
}

func (device *cortex) getValue(packet []byte, command string) int {
	tempArray := strings.Split(command, ";")
	offset, _ := strconv.Atoi(tempArray[1])
	offset = offset + 4
	offset = offset * 4
	value := utils.GetIntForByte(packet[offset:])
	return value
}
func (device *cortex) getResponseStatus(retVal []string) (string, bool) {
	if retVal == nil {
		return "Cannot Communicate with Cortex TTCP", false
	}
	packet, err := base64.StdEncoding.DecodeString(retVal[0])
	if err != nil {
		return "Unable to verify acknowledgement", false
	}
	return device.getAckAndError(packet)
}

func (device *cortex) getCommandPacket(cmd utils.Command, component utils.Component) []byte {
	temp := strings.Split(cmd.Command, ";")
	cmdInt, _ := strconv.Atoi(temp[1])
	argumentType := temp[2]
	var packet = make([]byte, 0)
	header := utils.GetByteArrayForInt(1234567890)
	trailer := utils.GetByteArrayForInt(-1234567890)
	length := utils.GetByteArrayForInt(32)
	flowID := utils.GetByteArrayForInt(1)
	componentCode := utils.GetByteArrayForInt(int(component.ComponentCode))
	noOfParams := utils.GetByteArrayForInt(1)
	command := utils.GetByteArrayForInt(cmdInt)
	argument := make([]byte, 0)
	if strings.EqualFold(argumentType, "int") {
		tempInt, _ := strconv.Atoi(cmd.ArgumentValue)
		argument = utils.GetByteArrayForInt(tempInt)
	}
	if strings.EqualFold(argumentType, "float") {
		tempFloat, _ := strconv.ParseFloat(cmd.ArgumentValue, 64)
		tempUInt := math.Float32bits(float32(tempFloat))
		argument = utils.GetByteArrayForInt(int(tempUInt))
	}

	packet = append(packet, header...)
	packet = append(packet, length...)
	packet = append(packet, flowID...)
	packet = append(packet, componentCode...)
	packet = append(packet, noOfParams...)
	packet = append(packet, command...)
	packet = append(packet, argument...)
	packet = append(packet, trailer...)
	return packet
}

func (device *cortex) getReadPacket(component utils.Component) []byte {
	var packet = make([]byte, 0)
	header := utils.GetByteArrayForInt(1234567890)
	trailer := utils.GetByteArrayForInt(-1234567890)
	length := utils.GetByteArrayForInt(20)
	flowID := utils.GetByteArrayForInt(0)
	componentCode := utils.GetByteArrayForInt(int(component.ComponentCode))

	packet = append(packet, header...)
	packet = append(packet, length...)
	packet = append(packet, flowID...)
	packet = append(packet, componentCode...)
	packet = append(packet, trailer...)
	return packet
}

func (device *cortex) getToggleIdlePatternPacket() []byte {
	var packet = make([]byte, 0)
	header := utils.GetByteArrayForInt(1234567890)
	trailer := utils.GetByteArrayForInt(-1234567890)
	length := utils.GetByteArrayForInt(28)
	flowID := utils.GetByteArrayForInt(1)
	componentCode := utils.GetByteArrayForInt(0x1010)
	commandCode := utils.GetByteArrayForInt(1)
	argumentValue := utils.GetByteArrayForInt(0xFFFFFFFD)

	packet = append(packet, header...)
	packet = append(packet, length...)
	packet = append(packet, flowID...)
	packet = append(packet, componentCode...)
	packet = append(packet, commandCode...)
	packet = append(packet, argumentValue...)
	packet = append(packet, trailer...)
	return packet
}

func (device *cortex) getAckAndError(response []byte) (string, bool) {
	if len(response) != 20 {
		return "", true
	}
	ack := int(response[15])
	switch ack {
	case 0:
		return "", true
	case 1:
		return "Invalid Syntax", false
	case 3:
		return "Component Not Mounted", false
	case 4:
		return "Not Remotely Configurable parameter", false
	case 5:
		return "Not Available", false
	case 7:
		return "Bad Command Size", false
	case 8:
		return "Non Initialized file", false
	case -1:
		return "Unidentified Message", false
	case -2:
		return "Connection Rejected", false
	default:
		return "Unknown Error, Contact PRISM Development team", false
	}
}

func (device *cortex) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *cortex) communicate(cmds []utils.Command, port string) []string {
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
