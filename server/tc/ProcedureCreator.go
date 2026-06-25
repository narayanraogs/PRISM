package tc

import (
	"fmt"
	"strings"
)

func getLineNumber() func() string {
	lineNo := 0
	return func() string {
		lineNo = lineNo + 10
		lineNoStr := fmt.Sprintf("01.%04d", lineNo)
		return fmt.Sprintf("%-10s", lineNoStr)
	}
}

func getCommandLine(lineNo string, command string, args string) string {
	return fmt.Sprintf("%-10s %-19s %s\n", lineNo, command, args)
}
func getContinuationLine(args string) string {
	return fmt.Sprintf("%-30s %s\n", " ", args)
}

func getSingleBatch(lineNo string, setCmd string, resetCmd string, noOfCommands int) string {
	var builder strings.Builder
	var lines = make([]string, 0)
	for i := 0; i < noOfCommands; i++ {
		if i%2 == 0 {
			if i == 0 {
				lines = append(lines, strings.ReplaceAll(getCommandLine(lineNo, "SEND", setCmd), "\n", ""))
			} else {
				lines = append(lines, strings.ReplaceAll(getContinuationLine(setCmd), "\n", ""))
			}
		} else {
			lines = append(lines, strings.ReplaceAll(getContinuationLine(resetCmd), "\n", ""))
		}
	}
	builder.WriteString(strings.Join(lines, ";\n"))
	builder.WriteString("\n")
	return builder.String()
}

func CreateProcedure(rxName string, setCmd string, resetCmd string, noOfCommands int) func() string {
	var creator = func() string {
		isSingleCommand := noOfCommands == 1
		maxBatchSize := 10
		lineNumber := getLineNumber()
		var builder strings.Builder
		builder.WriteString(getCommandLine(lineNumber(), "FLASH_DISPLAY", fmt.Sprintf("Procedure for %s with %d commands", rxName, noOfCommands)))

		if isSingleCommand {
			builder.WriteString(getCommandLine(lineNumber(), "SET TC", "TC_MODE dryrun"))
		}

		for noOfCommands > 0 {
			batch := maxBatchSize
			if batch > noOfCommands {
				batch = noOfCommands
			}
			noOfCommands = noOfCommands - batch
			cmds := getSingleBatch(lineNumber(), setCmd, resetCmd, batch)
			builder.WriteString(cmds)
		}

		if isSingleCommand {
			builder.WriteString(getCommandLine(lineNumber(), "SET TC", "TC_MODE nml"))
		}

		builder.WriteString(getCommandLine(lineNumber(), "END", ""))
		return builder.String()
	}
	return creator
}
