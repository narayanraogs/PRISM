package measurements

import (
	"fmt"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"time"
)

func init() {
	executeTest.Register("FrequencyDeviationMeasurement", "", newTxFrequencyDeviationMeasurement)
	results.Register("FrequencyDeviationMeasurement", results.NewDefaultProcessor([]string{"Results"}))
}

func newTxFrequencyDeviationMeasurement() executeTest.Tester {
	var test txFrequencyDeviationMeasurement
	return &test
}

type txFrequencyDeviationMeasurement struct {
	txBaseTest
}

func (test *txFrequencyDeviationMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.getInstruments()
}

func (test *txFrequencyDeviationMeasurement) DBValidate() error {
	readTxSubCarrFunc := func() error {
		return test.readTxSubCarrierDetails(test.config.TxName)
	}
	return test.validateAndPrepare(readTxSubCarrFunc)
}

func (test *txFrequencyDeviationMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txFrequencyDeviationMeasurement) measure(runner *StepRunner) error {
	start := time.Now()
	sa := test.ctx.Selected.SA

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)
	test.setupSAForDownlink(runner, true)
	var freq1 float64
	var freq2 float64

	runner.Run("Measuring frequency deviation", false, func() {
		runner.Exec(sa.SetMaxHold)
		respTrace := runner.Exec(getTraceDump(sa, 1001))
		runner.Exec(peakSearch(sa, true, 1))
		respRows := runner.Exec(sa.GetNoOfRowsToSkipInTrace)
		if runner.execErr != nil {
			return
		}
		fileData := respTrace.Result["TraceDump"].String
		skipRows := respRows.Result["NoOfRows"].Value
		freq1, freq2 = utils.MeasureFreqDeviation(fileData, int(skipRows))
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

	if freq1 > freq2 {
		tempFreq := freq2
		freq2 = freq1
		freq1 = tempFreq
	}

	if !runner.Describe && runner.Err() == nil {
		txFreq := float64(test.txSpec.Frequency)
		freqDevFromPeak := test.txSpecSubCarrier[test.txSpec.TxName].PeakFrequencyDeviation.Float64 / 2
		freq1Spec := txFreq - freqDevFromPeak
		freq2Spec := txFreq + freqDevFromPeak
		freq1Dev := freq1 - freq1Spec
		freq2Dev := freq2 - freq2Spec

		resultData := txFrequencyDeviationResult{
			Freq1Spec: freq1Spec,
			Freq2spec: freq2Spec,
			Freq1Meas: resultValue{Value: freq1, Status: "Success"},
			Freq2Meas: resultValue{Value: freq2, Status: "Success"},
			Freq1Dev:  freq1Dev,
			Freq2Dev:  freq2Dev,
		}

		rows := [][]string{resultData.ToRow()}
		header := resultData.ToHeader()
		test.saveResultsToCSV("", header, rows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
