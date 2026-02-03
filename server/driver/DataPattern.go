package driver

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
)

//go:embed instructions/DataPatternCommands.csv
var dpInstructions string

type dp struct {
	connection instrument
	commands   map[string]utils.Command
	components map[string]utils.Component
}

func (device *dp) loadLANDetails(name string) bool {
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

func (device *dp) loadCommands() bool {
	inst := readCSV(dpInstructions)
	if inst == nil {
		fmt.Println("Unable to read Instructions CSV")
		return false
	}
	device.commands = inst
	return true
}

func (device *dp) initializeDevice(name string) {
	device.loadLANDetails(name)
	device.connection.Configure(" ", "\n", false, true)
	device.loadCommands()
}

func (device *dp) setCarrierOn(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setPRBS", "setFrame", "setCW")
	arguments = append(arguments, "0", "0", "1")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) setCarrierOff(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setPRBS", "setFrame", "setCW")
	arguments = append(arguments, "0", "0", "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) setModulationOn(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setPRBS", "setCW", "setFrame")
	arguments = append(arguments, "0", "0", "1")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) setModulationOff(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setPRBS", "setCW", "setFrame")
	arguments = append(arguments, "0", "0", "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) setFrequency(component string, frequency float64) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	freqInt := int(frequency)
	freqStr := strconv.Itoa(freqInt)
	mnemonics = append(mnemonics, "setFrequency")
	arguments = append(arguments, freqStr)
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) switchOnIdlePattern(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setCW", "setFrame", "setPRBS")
	arguments = append(arguments, "0", "0", "1")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) switchOffIdlePattern(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "setCW", "setFrame", "setPRBS")
	arguments = append(arguments, "0", "0", "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}
func (device *dp) getResponseStatus(retVal []string) (string, bool) {
	if retVal == nil {
		return "Cannot Communicate with DataPattern", false
	}
	packet, err := base64.StdEncoding.DecodeString(retVal[0])
	if err != nil {
		return "Unable to verify acknowledgement", false
	}
	return device.getAckAndError(packet)
}

func (device *dp) getCommandPacket(cmd utils.Command) []byte {
	temp := strings.Split(cmd.Command, ";")
	length, _ := strconv.Atoi(temp[1])
	actualLength := length
	var packet = make([]byte, 0)
	packetID, _ := strconv.ParseInt(temp[0], 0, 64)
	byte1 := byte(packetID & 0xFF)
	packetID = packetID >> 8
	byte2 := byte(packetID & 0xFF)
	len1 := byte(length & 0xFF)
	length = length >> 8
	len2 := byte(length & 0xFF)
	packet = append(packet, 0xAA, 0x55, byte1, byte2, 0x00, 0x00, len1, len2)
	data, _ := strconv.Atoi(cmd.ArgumentValue)
	for i := 0; i < actualLength; i++ {
		value := byte(data & 0xFF)
		packet = append(packet, value)
		data = data >> 8
	}
	checksum := computeChecksum(packet)
	packet = append(packet, checksum...)
	fmt.Printf("% 02X\n", packet)
	return packet
}
func computeChecksum(packet []byte) []byte {
	var checksum = make([]byte, 2)
	for i := 2; i < len(packet); i++ {
		checksum[0] = checksum[0] ^ packet[i]
	}
	return checksum
}

func (device *dp) readTable(component string) utils.CommandResponse {
	component = ""
	var mnemonics = make([]string, 0)
	var arguments = make([]string, 0)
	mnemonics = append(mnemonics, "getConfiguration")
	arguments = append(arguments, "0")
	var cmds = device.getCommands(mnemonics, arguments, make([]string, len(mnemonics)))
	packet := device.getCommandPacket(cmds[0])
	cmds[0].Packet = packet

	retVal := device.communicate(cmds, "Control")
	ack, ok := device.getResponseStatus(retVal)
	if !ok {
		return getErrorResponse(ack)
	}
	return getSuccessResponse()
}

func (device *dp) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	for i, mnemonic := range mnemonics {
		cmd := device.commands[mnemonic]
		cmd.ArgumentValue = arguments[i]
		cmd.Command = strings.ReplaceAll(cmd.Command, "#", replace[i])
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (device *dp) communicate(cmds []utils.Command, port string) []string {
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

func (device *dp) getAckAndError(response []byte) (string, bool) {
	return "", true
}

func (device *dp) checkConnection() utils.CommandResponse {
	return device.readTable("")
}

//----------------------Functions not implemented for DataPattern---------------------

func (device *dp) setFrequencyDeviationTC(component string, deviation float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}
func (device *dp) setFrequencyDeviationTone(component string, deviation float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}
func (device *dp) setModIndexTC(component string, modIndex float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setModIndexTone(component string, modIndex float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}
func (device *dp) setRangingToneFrequency(frequency float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setOnlyTC(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setOnlyRanging(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setTCAndRanging(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setSweepRate(component string, sweepRate float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setSweepStep(component string, sweepStep float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) triggerSweep(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) sweepHold(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) sweepContinuous(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) enableDoppler(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) startSweep(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}
func (device *dp) stopSweep(component string) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setSweepRange(component string, sweepRange float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setPower(component string, power float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}

func (device *dp) setChipRate(component string, chipRate float64) utils.CommandResponse {
	fmt.Println("Not implemented for DataPattern")
	return getSuccessResponse()
}
func (device *dp) setDopplerCompensationEnable() utils.CommandResponse {
	fmt.Println("To be implemented")
	return getSuccessResponse()
}

func (device *dp) setDopplerCompensationDisable() utils.CommandResponse {
	fmt.Println("To be implemented")
	return getSuccessResponse()
}

func (device *dp) getDeviceTime() utils.CommandResponse {
	fmt.Println("To be implemented")
	return getSuccessResponse()
}

func (device *dp) getDopplerCompensationStatus(component string) utils.CommandResponse {
	fmt.Println("To be implemented")
	return getSuccessResponse()
}

func (device *dp) setIdlePatternOn() utils.CommandResponse {
	fmt.Println("IdlePattern On/Off is not supported by TTCP")
	return getSuccessResponse()
}

func (device *dp) setIdlePatternOff() utils.CommandResponse {
	fmt.Println("IdlePattern On/Off is not supported by TTCP")
	return getSuccessResponse()
}

func (device *dp) setDopplerCompensationTable(timeOffset int, frequencies []int, extendedFrequencies []int, times []int) utils.CommandResponse {
	fmt.Println("Doppler to be implented, not supported as of now")
	return getSuccessResponse()
}

func (device *dp) getValue(packet []byte, command string) int {
	fmt.Println("Not implemented for DataPattern")
	return -1
}

func (device *dp) getReadPacket(component utils.Component) []byte {
	fmt.Println("Not implemented for DataPattern")
	return nil
}

func (device *dp) getToggleIdlePatternPacket() []byte {
	fmt.Println("Not implemented for DataPattern")
	return nil
}
