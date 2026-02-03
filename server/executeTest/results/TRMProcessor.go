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

func NewTRMProcessor(reports []string) ResultProcessor {
	return &trmProcessor{
		RequiredReports: reports,
	}
}

type trmProcessor struct {
	RequiredReports []string
	configName      string
	testPhase       string
	batchSize       int
	noOfTRMs        string
	timePerTRM      string
	mode            string
}

func (p *trmProcessor) Process(filenames []string) (map[string]reports.Result, []string, []reports.Images, error) {
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
			p.noOfTRMs = d.getValue("NoOfTRMs")
			p.timePerTRM = d.getValue("TimePerTRM")
			batch := d.getValue("BatchSize")
			p.batchSize, _ = strconv.Atoi(batch)
			break
		}
	}
	reportIndex := 0
	for _, filename := range filenames {
		if strings.Contains(filename, "Details") {
			continue
		}
		if reportIndex >= len(p.RequiredReports) {
			break
		}
		res, img, err := p.generateReport(filename)
		if err != nil {
			return results, p.RequiredReports, nil, err
		}
		results[p.RequiredReports[reportIndex]] = res
		images = append(images, img...)
		reportIndex++
	}
	return results, p.RequiredReports, images, nil
}

func (p *trmProcessor) generateReport(filename string) (reports.Result, []reports.Images, error) {
	var result reports.Result
	var images []reports.Images
	var chart utils.Chart

	eachTRMTime, _ := strconv.ParseFloat(p.timePerTRM, 64)
	noOfTRMs, _ := strconv.ParseInt(p.noOfTRMs, 10, 32)

	result.Header = []string{"Parameter", "Specification", "Max", "Min", "Mean"}
	result.Data = make([][]reports.DataCell, 0)

	_, ok := database.GetPayloadSpec(p.configName, p.testPhase, p.mode)
	if !ok {
		return result, images, fmt.Errorf("unable to get Payload spec")
	}

	var pulse utils.PulseParameters
	err := pulse.LoadFile(filename, p.batchSize)
	if err != nil {
		return result, images, err
	}
	averagePowers := pulse.GetMiddleValues("TimeStamp", "AveragePower", eachTRMTime)
	acqTime := pulse.GetMiddleValues("TimeStamp", "TimeStamp", eachTRMTime)
	allDelta, consequtiveDelta := getDeltaTime(acqTime)

	allDelta, consequtiveDelta, averagePowers = getCleanData(allDelta, consequtiveDelta, averagePowers, eachTRMTime)
	averagePowers = averagePowers[:noOfTRMs]
	consequtiveDelta = consequtiveDelta[:noOfTRMs]
	averagePowers = averagePowers[:noOfTRMs]

	result.Header = []string{"No Of TRMs", "TRMs Not Present (\u00B11)"}
	result.Data = make([][]reports.DataCell, 0)
	var row = make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell(fmt.Sprintf("%d", len(averagePowers))))
	var missing string
	for i, value := range averagePowers {
		if math.IsNaN(value) {
			missing = missing + fmt.Sprintf("%d, ", i+1)
		}
	}
	missing = strings.TrimRight(missing, ", ")
	row = append(row, reports.GetDataCell(missing))
	result.Data = append(result.Data, row)

	plot1Values := make([][]float64, 0)
	plot1Values = append(plot1Values, averagePowers, allDelta)
	plot1Titles := make([]string, 0)
	plot1Titles = append(plot1Titles, "AveragePower", "AcqusitionTime")
	chart.MakeChart("TRM Plot", plot1Titles, plot1Values)
	ok = chart.Save("plot.png")
	if ok {
		data, err := os.ReadFile("plot.png")
		if err != nil {
			fmt.Println("Error ", err.Error())
		}
		image1 := reports.Images{
			FileData: base64.StdEncoding.EncodeToString(data),
			Caption:  "Plotted from Acquired Values",
		}

		images = append(images, image1)
	}

	plot2Values := make([][]float64, 0)
	plot2Values = append(plot2Values, consequtiveDelta)
	plot2Titles := make([]string, 0)
	plot2Titles = append(plot2Titles, "DeltaAcquisitionTime")
	var chart2 utils.Chart
	chart2.MakeChart("TRM Plot", plot2Titles, plot2Values)
	ok = chart2.Save("plot2.png")
	if ok {
		data, err := os.ReadFile("plot2.png")
		if err != nil {
			fmt.Println("Error ", err.Error())
		}
		image2 := reports.Images{
			FileData: base64.StdEncoding.EncodeToString(data),
			Caption:  "Plotted from Acquired Values",
		}
		images = append(images, image2)
	}

	return result, images, nil
}

func getDeltaTime(acqTime []float64) ([]float64, []float64) {
	var allDelta []float64
	var consequtiveDelta []float64
	var ref = acqTime[0]
	var prev = acqTime[0]

	for _, acq := range acqTime {
		delta1 := acq - ref
		delta1 = delta1 / 1000
		allDelta = append(allDelta, delta1)
		delta2 := acq - prev
		delta2 = delta2 / 1000
		consequtiveDelta = append(consequtiveDelta, delta2)
		prev = acq
	}
	return allDelta, consequtiveDelta
}

func getCleanData(allDelta []float64, consequtiveDelta []float64, average []float64, expected float64) ([]float64, []float64, []float64) {
	var cleanedAverage = make([]float64, 0)
	var cleanedAllDelta = make([]float64, 0)
	var cleanedConsequtive = make([]float64, 0)

	var tolerance = expected * 0.5
	fmt.Println(expected, tolerance)

	cleanedAverage = append(cleanedAverage, average[0])
	cleanedAllDelta = append(cleanedAllDelta, allDelta[0])
	cleanedConsequtive = append(cleanedConsequtive, consequtiveDelta[1])

	for i := 1; i < len(consequtiveDelta); i++ {
		expectedDelta := expected
		for math.Abs(consequtiveDelta[i]-expectedDelta) > tolerance {
			cleanedAverage = append(cleanedAverage, math.NaN())
			cleanedAllDelta = append(cleanedAllDelta, math.NaN())
			cleanedConsequtive = append(cleanedConsequtive, math.NaN())
			expectedDelta = expectedDelta + expected
		}
		cleanedAverage = append(cleanedAverage, average[i])
		cleanedAllDelta = append(cleanedAllDelta, allDelta[i])
		cleanedConsequtive = append(cleanedConsequtive, consequtiveDelta[i])
	}
	return cleanedAllDelta, cleanedConsequtive, cleanedAverage
}

/*var timestamps []float64
	var avgPowers []float64

	for _, line := range lines[1:] {
		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}
		ts, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		ap, _ := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
		timestamps = append(timestamps, ts)
		avgPowers = append(avgPowers, ap)
	}

	if len(timestamps) == 0 {
		return result, nil, fmt.Errorf("no valid records found in CSV")
	}

	noOfTRMs := 60
	eachTRMTime := 1.0

	allDelta, consecDelta := getDeltaTime(timestamps)
	allDelta, consecDelta, avgPowers = getCleanData(allDelta, consecDelta, avgPowers, eachTRMTime)

	if len(avgPowers) > noOfTRMs {
		avgPowers = avgPowers[:noOfTRMs]
	}

	row := make([]reports.DataCell, 0, 2)
	row = append(row, reports.GetDataCell(fmt.Sprintf("%d", len(avgPowers))))

	var missing string
	for i, v := range avgPowers {
		if math.IsNaN(v) {
			missing += fmt.Sprintf("%d, ", i+1)
		}
	}
	missing = strings.TrimRight(missing, ", ")
	row = append(row, reports.GetDataCell(missing))
	result.Data = append(result.Data, row)

	plot1Vals := [][]float64{avgPowers, allDelta}
	plot1Titles := []string{"AveragePower", "AcquisitionTime"}

	var chart1 utils.Chart
	chart1.MakeChart("TRM Plot 1", plot1Titles, plot1Vals)
	if ok := chart1.Save("trm_plot1.png"); ok {
		if data, err := os.ReadFile("trm_plot1.png"); err == nil {
			imgs = append(imgs, reports.Images{
				FileData: base64.StdEncoding.EncodeToString(data),
				Caption:  "Average Power vs Time",
			})
		}
	}

	plot2Vals := [][]float64{consecDelta}
	plot2Titles := []string{"DeltaAcquisitionTime"}

	var chart2 utils.Chart
	chart2.MakeChart("TRM Plot 2", plot2Titles, plot2Vals)
	if ok := chart2.Save("trm_plot2.png"); ok {
		if data, err := os.ReadFile("trm_plot2.png"); err == nil {
			imgs = append(imgs, reports.Images{
				FileData: base64.StdEncoding.EncodeToString(data),
				Caption:  "Delta Acquisition Time per TRM",
			})
		}
	}

	return result, imgs, nil
}

func (p *trmProcessor) reportKey(i int) string {
	if i >= 0 && i < len(p.RequiredReports) {
		return p.RequiredReports[i]
	}
	return fmt.Sprintf("TRM-%d", i+1)
}

func getDeltaTime(acqTime []float64) ([]float64, []float64) {
	allDelta := make([]float64, 0, len(acqTime))
	consecutive := make([]float64, 0, len(acqTime))
	if len(acqTime) == 0 {
		return allDelta, consecutive
	}
	ref := acqTime[0]
	prev := acqTime[0]
	for _, t := range acqTime {
		d1 := (t - ref) / 1000.0
		d2 := (t - prev) / 1000.0
		allDelta = append(allDelta, d1)
		consecutive = append(consecutive, d2)
		prev = t
	}
	return allDelta, consecutive
}

func getCleanData(allDelta, consecutive, average []float64, expected float64) ([]float64, []float64, []float64) {
	cleanAvg := make([]float64, 0, len(average))
	cleanAll := make([]float64, 0, len(allDelta))
	cleanCon := make([]float64, 0, len(consecutive))
	if len(consecutive) == 0 {
		return cleanAll, cleanCon, cleanAvg
	}

	tolerance := expected * 0.5
	cleanAvg = append(cleanAvg, average[0])
	cleanAll = append(cleanAll, allDelta[0])
	if len(consecutive) > 1 {
		cleanCon = append(cleanCon, consecutive[1])
	} else {
		cleanCon = append(cleanCon, math.NaN())
	}

	for i := 1; i < len(consecutive); i++ {
		exp := expected
		for math.Abs(consecutive[i]-exp) > tolerance {
			cleanAvg = append(cleanAvg, math.NaN())
			cleanAll = append(cleanAll, math.NaN())
			cleanCon = append(cleanCon, math.NaN())
			exp += expected
		}
		cleanAvg = append(cleanAvg, average[i])
		cleanAll = append(cleanAll, allDelta[i])
		cleanCon = append(cleanCon, consecutive[i])
	}
	return cleanAll, cleanCon, cleanAvg
}
*/
