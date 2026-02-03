package measurements

import (
	"fmt"
	"math"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"time"
)

func init() {
	executeTest.Register("Frequency", "", newTxFrequencyMeasurement)
	results.Register("Frequency", results.NewDefaultProcessor([]string{"Results"}))
}

func newTxFrequencyMeasurement() executeTest.Tester {
	var test txFrequencyMeasurement
	return &test
}

type txFrequencyMeasurement struct {
	txBaseTest
}

func (test *txFrequencyMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *txFrequencyMeasurement) DBValidate() error {
	return test.validateAndPrepare(nil)
}

func (test *txFrequencyMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txFrequencyMeasurement) measure(runner *StepRunner) error {
	start := time.Now()

	sa := test.ctx.Selected.SA

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)
	test.setupSAForDownlink(runner, false)

	var freq float64
	runner.Run("Reading Frequency", false, func() {
		resp := runner.Exec(getFrequencyInCounterMode(sa, 1))
		if runner.execErr != nil {
			return
		}
		freq = resp.Result["Frequency"].Value / 1e6
		test.success(fmt.Sprintf("%.6f MHz", freq))
	})

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Downlink Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	test.saToNormalMode(runner)

	if !runner.Describe && runner.Err() == nil {

		deviation := freq - (float64(test.txSpec.Frequency) / 1e6)
		deviationStatus := "Success"
		if math.Abs(deviation) > test.txSpec.AllowedFrequencyDeviation {
			deviationStatus = "Error"
		}
		fmt.Println(float64(test.txSpec.Frequency))

		ppm := deviation / (float64(test.txSpec.Frequency) / 1e6) * 1e6

		resultData := txFrequencyResult{
			SpecificationMHz:    float64(test.txSpec.Frequency) / 1e6,
			MeasuredMHz:         freq,
			AllowedDeviationKHz: test.txSpec.AllowedFrequencyDeviation / 1e3,
			DeviationKHz:        resultValue{Value: deviation * 1e3, Status: deviationStatus},
			DeviationPPM:        resultValue{Value: ppm, Status: deviationStatus},
		}

		rows := [][]string{resultData.ToRow()}
		header := resultData.ToHeader()
		test.saveResultsToCSV("", header, rows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
