package results

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"prismServer/database"
	"prismServer/reports"
	"prismServer/utils"
	"strconv"
	"strings"
)

func NewHRProcessor(reports []string) ResultProcessor {
	return &hrProcessor{
		RequiredReports: reports,
	}
}

type hrProcessor struct {
	RequiredReports []string
	configName      string
	testPhase       string
	batchSize       int
	mode            string
}

func (p *hrProcessor) Process(filenames []string) (map[string]reports.Result, []string, []reports.Images, error) {
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
			batch := d.getValue("BatchSize")
			p.batchSize, _ = strconv.Atoi(batch)
			break
		}
	}
	for i, filename := range filenames {
		if strings.Contains(filename, "Details") {
			continue
		}
		res, img, err := p.generateReport(filename)
		if err != nil {
			return results, p.RequiredReports, nil, err
		}
		results[p.RequiredReports[i]] = res
		images = append(images, img)

	}
	return results, p.RequiredReports, images, nil
}

func (p *hrProcessor) generateReport(filename string) (reports.Result, reports.Images, error) {
	var result reports.Result
	var image reports.Images
	var chart utils.Chart

	if strings.Contains(filename, "HRPulse") {
		p.mode = "HR"
	} else if strings.Contains(filename, "LRPulse") {
		p.mode = "LR"
	} else {
		p.mode = ""
	}

	result.Header = []string{"Parameter", "Specification", "Max", "Min", "Mean"}
	result.Data = make([][]reports.DataCell, 0)
	spec := make(map[string]database.PLParameterSpec)
	var ok bool
	columnMap := make(map[string]string)
	var selectedParametes []string
	var plotFields []string
	var pulse utils.PulseParameters

	if strings.EqualFold(p.mode, "HR") || strings.EqualFold(p.mode, "") {
		spec, ok = database.GetPayloadSpec(p.configName, p.testPhase, p.mode) //changed here
		if !ok {
			return result, image, fmt.Errorf("unable to get Payload spec")
		}

		columnMap = getColumnMap(false)
		selectedParametes = utils.GetSelectedVSAParams()
		plotFields = []string{"PeakPowerVSA", "AverageTxPowerVSA", "PulseWidth", "PulsePeriod", "Droop", "RepetitionRate"}

	} else {
		spec["No Of Pulses"] = database.PLParameterSpec{
			Name:             "No Of Pulses",
			Unit:             "",
			ExpectedValue:    16.0,
			AllowedDeviation: 1.0,
			Operation: func(val float64) (float64, string) {
				return val, fmt.Sprintf("%.0f", val)
			},
		}
		spec["Off Time"] = database.PLParameterSpec{
			Name:             "Off Time",
			Unit:             "\u00B5s",
			ExpectedValue:    3830,
			AllowedDeviation: 100,
			Operation: func(val float64) (float64, string) {
				val = val * 1e3
				return val, fmt.Sprintf("%.2f", val)
			},
		}

		columnMap["No Of Pulses"] = "No Of Pulses"
		columnMap["Off Time"] = "Off Time"
		selectedParametes = []string{"No Of Pulses", "Off Time"}
		plotFields = []string{"No Of Pulses", "Off Time"}
	}

	err := pulse.LoadFile(filename, p.batchSize)
	if err != nil {
		return result, image, err
	}

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

		maxValue, _ := pulse.GetMaxValue(columnMap[param])
		_, maxStr := calculator(maxValue)

		cellMax := reports.GetDataCell(maxStr)
		row = append(row, cellMax)
		minValue, _ := pulse.GetMinValue(columnMap[param])
		_, minStr := calculator(minValue)
		cellMin := reports.GetDataCell(minStr)
		row = append(row, cellMin)
		measured, _, _ = pulse.GetMeanSDValue(columnMap[param])
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

	var values = make([][]float64, 0)
	for _, param := range plotFields {
		pl := spec[param]
		op := pl.Operation

		column, ok := pulse.GetValues(columnMap[param])
		if !ok {
			fmt.Println("Cannot read CSV for ", param, columnMap[param])
			continue
		}
		computed := make([]float64, 0)
		for _, v := range column {
			c, _ := op(v)
			computed = append(computed, c)
		}
		values = append(values, computed)
	}
	chart.MakeChart("VSA Plot", plotFields, values)
	ok = chart.Save("plot.png")
	if ok {
		data, err := os.ReadFile("plot.png")
		if err != nil {
			fmt.Println("Error ", err.Error())
		}
		image = reports.Images{
			FileData: base64.StdEncoding.EncodeToString(data),
			Caption:  "Plotted from Acquired Values",
		}
	}

	return result, image, nil
}
