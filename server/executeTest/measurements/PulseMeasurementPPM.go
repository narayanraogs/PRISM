package measurements

import (
	"fmt"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"strings"
	"time"

	"prismServer/utils"
)

func init() {
	executeTest.Register("PulseMeasurement", "PPM", newPulseMeasurementPPM)
	results.Register("PulseMeasurement", results.NewPPMProcessor([]string{"Results"}))
}

func newPulseMeasurementPPM() executeTest.Tester {
	var test pulseMeasurementPPM
	return &test
}

type pulseMeasurementPPM struct {
	pulseBaseTest

	pulsePeakPowerStr string
	pulseAvgPowerStr  string
	pulseWidthStr     string
	pulsePeriodStr    string
	pulseOffTimeStr   string
	riseTimeStr       string
	fallTimeStr       string
	dutyCycleStr      string
}

func (test *pulseMeasurementPPM) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.pulseBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *pulseMeasurementPPM) DBValidate() error {
	return test.validateAndPrepare(true, nil)
}

func (test *pulseMeasurementPPM) getInstruments() {
	test.ctx.Progress.Instruments = []string{"PPM", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *pulseMeasurementPPM) measure(runner *StepRunner) error {
	start := time.Now()
	ppm := test.ctx.Selected.PPM

	if test.rollbackToBeRead {
		test.readRollback(runner) // in base test, sa and pm are there as rollback, added in pulseBaseTest
	}

	test.setTSMPathForPPM(runner)

	runner.Run("Presetting PPM and configuring Pulse Mode", true, func() {
		runner.Exec(ppm.PresetPPM)
		runner.Exec(setChannelFrequency(ppm, test.pulseParameters.PPMChannel, test.pulseSpec.CenterFrequency))
		runner.Exec(setPulseParameters(ppm, test.pulseSpec.PulseWidth, test.pulseSpec.PulsePeriod,
			test.pulseParameters.PPMTriggerLevel, test.pulseParameters.PPMReferenceLevel,
			test.pulseParameters.PPMYDivision, test.pulseParameters.PPMChannel, true))
	})

	runner.Run("Waiting for MAP ON", true, func() {
		test.ctx.AskForConfirmation("Press Continue after Giving Map On", 0)
	})

	var pulsePeakPower float64
	var pulseAvgPower float64
	var pulseWidth float64
	var pulsePeriod float64
	var pulseOffTime float64
	var riseTime float64
	var fallTime float64
	var dutyCycle float64

	runner.Run("Reading Pulse Parameters", true, func() {
		time.Sleep(2000 * time.Millisecond)
		Resp := runner.Exec(getPeakPower(ppm, test.pulseParameters.PPMChannel))
		pulsePeakPower = Resp.Result["PulsePeakPower"].Value
		pulseAvgPower = Resp.Result["PulseAveragePower"].Value
		time.Sleep(1000 * time.Millisecond)
		pulseWidth = runner.Exec(getPulseWidth(ppm, test.pulseParameters.PPMChannel)).Result["PulseOnTime"].Value
		time.Sleep(1000 * time.Millisecond)
		pulsePeriod = runner.Exec(getPulsePeriod(ppm, test.pulseParameters.PPMChannel)).Result["PulsePeriod"].Value
		time.Sleep(1000 * time.Millisecond)
		pulseOffTime = runner.Exec(getPulseOffTime(ppm, test.pulseParameters.PPMChannel)).Result["PulseOffTime"].Value
		time.Sleep(1000 * time.Millisecond)
		riseTime = runner.Exec(getRiseTime(ppm, test.pulseParameters.PPMChannel)).Result["RiseTime"].Value
		time.Sleep(1000 * time.Millisecond)
		fallTime = runner.Exec(getFallTime(ppm, test.pulseParameters.PPMChannel)).Result["FallTime"].Value
		time.Sleep(1000 * time.Millisecond)
		dutyCycle = runner.Exec(getDutyCycle(ppm, test.pulseParameters.PPMChannel)).Result["DutyCycle"].Value

		test.pulsePeakPowerStr = fmt.Sprintf("%.2f", pulsePeakPower)
		test.pulseAvgPowerStr = fmt.Sprintf("%.2f", pulseAvgPower)
		test.pulseWidthStr = fmt.Sprintf("%.6f", pulseWidth)
		test.pulsePeriodStr = fmt.Sprintf("%.6f", pulsePeriod)
		test.pulseOffTimeStr = fmt.Sprintf("%.6f", pulseOffTime)
		test.riseTimeStr = fmt.Sprintf("%.9f", riseTime)
		test.fallTimeStr = fmt.Sprintf("%.9f", fallTime)
		test.dutyCycleStr = fmt.Sprintf("%.2f", dutyCycle)

	})

	if !runner.Describe && runner.Err() == nil {
		detailsHeader := []string{"Parameter", "Value"}
		tp, _ := database.GetSelectedTestPhase()

		detailRows := make([][]string, 0)
		detailRows = append(detailRows, []string{"Config", test.configName})

		detailRows = append(detailRows, []string{"TestPhase", tp})

		var line1 = make([]string, 0)
		var line2 = make([]string, 0)

		line1 = append(line1, "PeakPower", "AveragePower", "PulseWidth", "PulsePeriod")
		line1 = append(line1, "PulseSeparation", "RiseTime", "FallTime", "DutyCycle")
		line2 = append(line2, test.pulsePeakPowerStr, test.pulseAvgPowerStr, test.pulseWidthStr, test.pulsePeriodStr,
			test.pulseOffTimeStr, test.riseTimeStr, test.fallTimeStr, test.dutyCycleStr)

		var fileContent string
		fileContent = fileContent + strings.Join(line1, ",")
		fileContent = fileContent + "\n"
		fileContent = fileContent + strings.Join(line2, ",")
		fileContent = fileContent + "\n"

		csvDir := utils.GetCSVResultDirectory()
		csvDir = filepath.Join(csvDir, test.testName)
		_ = os.MkdirAll(csvDir, 0755)
		fileName := test.testName
		if strings.TrimSpace(test.testCategory) != "" {
			fileName += "-" + test.testCategory
		}
		fileName += "-" + test.configName

		fileName = utils.GetOldTimeStampedFileName(fileName, test.reportTime) + ".csv"
		fullPath := filepath.Join(csvDir, fileName)
		_ = os.WriteFile(fullPath, []byte(fileContent), 0666)

		test.filenames = append(test.filenames, fullPath)
		test.saveResultsToCSV("Details", detailsHeader, detailRows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
