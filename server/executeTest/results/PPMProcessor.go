package results

import (
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/reports"
	"prismServer/utils"
	"strings"
)

func NewPPMProcessor(reports []string) ResultProcessor {
	return &ppmProcessor{
		RequiredReports: reports,
	}
}

type ppmProcessor struct {
	RequiredReports []string
	configName      string
	testPhase       string
	pulsePeakPower  string
	pulseAvgPower   string
	pulseWidth      string
	pulsePeriod     string
	pulseOffTime    string
	riseTime        string
	fallTime        string
	dutyCycle       string
	mode            string
}

func (p *ppmProcessor) Process(filenames []string) (map[string]reports.Result, []string, []reports.Images, error) {
	var results = make(map[string]reports.Result)
	var images = make([]reports.Images, 0)
	for _, filename := range filenames {
		if strings.Contains(filename, "Details") {
			var d details
			err := d.load(filename)
			if err != nil {
				return nil, nil, nil, err
			}
			p.configName = d.getValue("Config")
			p.testPhase = d.getValue("TestPhase")

			break
		}
	}
	for i, filename := range filenames {
		if strings.Contains(filename, "Details") {
			continue
		}
		res, _, err := p.generateReport(filename)
		if err != nil {
			return results, p.RequiredReports, nil, err
		}
		results[p.RequiredReports[i]] = res

	}
	return results, p.RequiredReports, images, nil
}

func (p *ppmProcessor) generateReport(filename string) (reports.Result, reports.Images, error) {
	var result reports.Result
	var image reports.Images

	result.Header = []string{"Parameter", "Specification", "Measured"}
	result.Data = make([][]reports.DataCell, 0)

	spec, ok := database.GetPayloadSpec(p.configName, p.testPhase, p.mode)
	if !ok {
		return result, image, fmt.Errorf("unable to get Payload spec")
	}

	var pulse utils.PulseParameters
	err := pulse.LoadFile(filename, -1) //todo: make separate fileName
	if err != nil {
		return result, image, err
	}
	columnMap := getColumnMap(true)
	selectedParametes := utils.GetSelectedPPMParams()

	for _, param := range selectedParametes {
		var row = make([]reports.DataCell, 0)
		pl := spec[param]
		unit := pl.Unit
		name := param + " [" + unit + "]"
		row = append(row, reports.GetDataCell(name))

		expected := pl.ExpectedValue
		allowed := pl.AllowedDeviation
		ll := expected - allowed
		ul := expected + allowed

		var specToReport string
		if allowed == 0 {
			specToReport = fmt.Sprintf("%0.2f", expected)
		} else {
			specToReport = fmt.Sprintf("%0.2f to %0.2f", ll, ul)
		}

		row = append(row, reports.GetDataCell(specToReport))
		var calculator = pl.Operation
		var measured float64
		var measuredStr string

		measured, _ = pulse.GetFirstValue(columnMap[param])
		measured, measuredStr = calculator(measured)

		cell := reports.GetDataCell(measuredStr)
		deviation := measured - expected
		if math.Abs(deviation) < allowed {
			cell.SetSuccess()
		} else {
			cell.SetError()
		}
		row = append(row, cell)

		result.Data = append(result.Data, row)
	}

	return result, image, nil
}
