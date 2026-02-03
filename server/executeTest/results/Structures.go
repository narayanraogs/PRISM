package results

import (
	"prismServer/reports"
)

type ResultProcessor interface {
	Process(filenames []string) (map[string]reports.Result, []string, []reports.Images, error)
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
