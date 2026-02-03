package results

import (
	"fmt"
	"prismServer/reports"
)

var Registry = make(map[string]ResultProcessor)

func Register(name string, processor ResultProcessor) {
	Registry[name] = processor
}

func GenerateResults(testName string, filenames []string) (map[string]reports.Result, []string, []reports.Images, error) {
	processor, ok := Registry[testName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("no result processor found for test %s", testName)
	}
	return processor.Process(filenames)
}
