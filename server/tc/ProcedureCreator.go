package tc

import (
	"fmt"
	"strings"
)

func getLineNumber() func() string {
	lineNo := 0
	return func() string {
		lineNo = lineNo + 10
		return fmt.Sprintf("01.%04d \t", lineNo)
	}
}

func getSingleBatch(setCmd string, resetCmd string, noOfCommands int) string {
	var builder strings.Builder
	var cmds = make([]string, 0, noOfCommands)
	for i := 0; i < noOfCommands; i++ {
		if i%2 == 0 {
			cmds = append(cmds, "\t"+setCmd)
		} else {
			cmds = append(cmds, "\t"+resetCmd)
		}
	}
	builder.WriteString(strings.Join(cmds, ";\n"))
	return builder.String()
}

func CreateProcedure(rxName string, setCmd string, resetCmd string, noOfCommands int) func() string {
	var creator = func() string {
		maxBatchSize := 10
		lineNumber := getLineNumber()
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("%s FLASH_DISPLAY Auto generated procedure for %s, with %d commands\n",
			lineNumber(), rxName, noOfCommands))
		for noOfCommands > 0 {
			batch := maxBatchSize
			if batch > noOfCommands {
				batch = noOfCommands
			}
			noOfCommands = noOfCommands - batch
			cmds := getSingleBatch(setCmd, resetCmd, batch)
			builder.WriteString(fmt.Sprintf("%s SEND %s\n", lineNumber(), cmds))
		}
		builder.WriteString(fmt.Sprintf("%s END\n", lineNumber()))
		return builder.String()
	}
	return creator
}
