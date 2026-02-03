package driver

import (
	"fmt"
	"prismServer/utils"
	"slices"
	"strconv"
	"strings"
)

func readComponents(fileData string) map[string]utils.Component {
	lines := strings.Split(fileData, "\n")
	slNoIndex := -1
	componentNameIndex := -1
	componentTypeIndex := -1
	componentCodeIndex := -1

	var components = make(map[string]utils.Component)
	for i, line := range lines {
		if i == 0 {
			var values = getFields(line)
			slNoIndex = slices.Index(values, "SlNo")
			componentNameIndex = slices.Index(values, "ComponentName")
			componentTypeIndex = slices.Index(values, "ComponentType")
			componentCodeIndex = slices.Index(values, "ComponentCode")
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		var comp utils.Component
		var slNoInt int
		var err error

		var values = getFields(line)
		if slNoIndex != -1 {
			tempSlNo := strings.TrimSpace(values[slNoIndex])
			slNoInt, err = strconv.Atoi(tempSlNo)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}
			comp.SlNo = slNoInt
		}
		if componentNameIndex != -1 {
			comp.ComponentName = strings.TrimSpace(values[componentNameIndex])
		}
		if componentTypeIndex != -1 {
			comp.ComponentType = strings.TrimSpace(values[componentTypeIndex])
		}
		if componentCodeIndex != -1 {
			tempCode := strings.TrimSpace(values[componentCodeIndex])
			compCodeInt, err := strconv.ParseInt(tempCode, 16, 64)
			if err != nil {
				fmt.Println(tempCode, err.Error())
				return nil
			}
			comp.ComponentCode = compCodeInt
		}
		components[comp.ComponentName] = comp
	}
	return components
}

func readCSV(fileData string) map[string]utils.Command {
	lines := strings.Split(fileData, "\n")
	slNoIndex := -1
	mnemonicIndex := -1
	commandIndex := -1
	delayIndex := -1
	argumentIndex := -1
	readIndex := -1
	readBinaryIndex := -1
	dataTypeIndex := -1
	componentIndex := -1
	portIndex := -1

	var instructions = make(map[string]utils.Command)
	for i, line := range lines {
		if i == 0 {
			var values = getFields(line)
			slNoIndex = slices.Index(values, "SlNo")
			mnemonicIndex = slices.Index(values, "Mnemonic")
			commandIndex = slices.Index(values, "Command")
			delayIndex = slices.Index(values, "Delay")
			argumentIndex = slices.Index(values, "Argument")
			readIndex = slices.Index(values, "Read")
			readBinaryIndex = slices.Index(values, "ReadBinary")
			dataTypeIndex = slices.Index(values, "DataType")
			componentIndex = slices.Index(values, "Component")
			portIndex = slices.Index(values, "Port")
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		var inst utils.Command
		var slNoInt int
		var err error

		var values = getFields(line)
		if slNoIndex != -1 {
			tempSlNo := strings.TrimSpace(values[slNoIndex])
			slNoInt, err = strconv.Atoi(tempSlNo)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}
			inst.SlNo = slNoInt
		}
		if mnemonicIndex != -1 {
			inst.Mnemonic = strings.TrimSpace(values[mnemonicIndex])
		}
		if commandIndex != -1 {
			inst.Command = strings.TrimSpace(values[commandIndex])
		}
		if delayIndex != -1 {
			tempDelay := strings.TrimSpace(values[delayIndex])
			delay, err := strconv.ParseFloat(tempDelay, 64)
			if err != nil {
				fmt.Println(slNoInt, err.Error())
				return nil
			}
			inst.Delay = delay
		}
		if argumentIndex != -1 {
			tempArg := strings.TrimSpace(values[argumentIndex])
			inst.Argument = strings.EqualFold(tempArg, "true")
		}
		if readIndex != -1 {
			tempArg := strings.TrimSpace(values[readIndex])
			inst.Read = strings.EqualFold(tempArg, "true")
		}
		if readBinaryIndex != -1 {
			tempBinary := strings.TrimSpace(values[readBinaryIndex])
			inst.ReadBinary = strings.EqualFold(tempBinary, "true")
		}
		if dataTypeIndex != -1 {
			inst.DataType = strings.TrimSpace(values[dataTypeIndex])
		}
		if componentIndex != -1 {
			inst.Component = strings.TrimSpace(values[componentIndex])
		}
		if portIndex != -1 {
			inst.Port = strings.TrimSpace(values[portIndex])
		}
		instructions[inst.Mnemonic] = inst
	}
	return instructions
}

func getFields(line string) []string {
	var values = make([]string, 0)
	var temp = ""
	var quote bool = false
	for _, char := range line {
		if char == '"' {
			quote = !quote
			continue
		}
		if char == ',' && !quote {
			values = append(values, temp)
			temp = ""
			continue
		}
		temp = temp + string(char)
	}
	values = append(values, temp)
	return values
}
