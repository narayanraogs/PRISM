package results

import (
	"os"
	"prismServer/reports"
	"strings"
)

func NewDefaultProcessor(reports []string) ResultProcessor {
	return &defaultProcessor{
		RequiredReports: reports,
	}
}

type defaultProcessor struct {
	RequiredReports []string
}

func (p *defaultProcessor) Process(filenames []string) (map[string]reports.Result, []string, []reports.Images, error) {
	var results = make(map[string]reports.Result)
	for i, filename := range filenames {
		res, err := p.generateReport(filename)
		if err != nil {
			return results, p.RequiredReports, make([]reports.Images, 0), err
		}
		results[p.RequiredReports[i]] = res
	}
	return results, p.RequiredReports, make([]reports.Images, 0), nil
}

func (p *defaultProcessor) generateReport(filename string) (reports.Result, error) {
	var result reports.Result
	data, err := os.ReadFile(filename)
	if err != nil {
		return result, err
	}
	fileData := string(data)
	lines := strings.Split(fileData, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		row := make([]reports.DataCell, 0)
		fields := getFields(line)
		if i == 0 {
			result.Header = make([]string, 0)
			result.Header = append(result.Header, fields...)
			continue
		}
		for _, field := range fields {
			temp := strings.Split(field, ";")
			cell := reports.GetDataCell(temp[0])
			if len(temp) > 1 {
				if strings.EqualFold("Success", temp[1]) {
					cell.Success = true
				}
				if strings.EqualFold("Error", temp[1]) {
					cell.Error = true
				}
			}
			row = append(row, cell)
		}
		result.Data = append(result.Data, row)
	}
	return result, nil
}
