package measurements

import (
	"fmt"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"slices"
	"time"
)

func init() {
	executeTest.Register("Spurious", "", newTxSpuriousMeasurement)
	results.Register("Spurious", results.NewDefaultProcessor([]string{"Results"}))
}

func newTxSpuriousMeasurement() executeTest.Tester {
	var test txSpuriousMeasurement
	return &test
}

type txSpuriousMeasurement struct {
	txBaseTest
	subCarrierFrequencies []float64
}

func (test *txSpuriousMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *txSpuriousMeasurement) DBValidate() error {
	readSubCarrierFunc := func() error {
		return test.readTxSubCarrierDetails(test.config.TxName)
	}
	return test.validateAndPrepare(readSubCarrierFunc)
}

func (test *txSpuriousMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txSpuriousMeasurement) measure(runner *StepRunner) error {
	start := time.Now()
	spurSpec := test.txSpec.Spurious
	var spuriousRows = make([][]string, 0)
	var header []string

	sa := test.ctx.Selected.SA

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)
	test.setupSAForDownlink(runner, false)

	var freqs = make([]float64, 0)
	var peaks = make([]float64, 0)

	runner.Run("Detecting Spurious", false, func() {
		runner.Exec(sa.SetMaxHold)
		runner.Exec(waitForSweeps(sa, 5))
		runner.Exec(getTraceDump(sa, 1001))
		resp := runner.Exec(getAllPeaksAbove(sa, 10, 1))
		if runner.execErr != nil {
			return
		}
		freqs = resp.Result["Frequencies"].Values
		peaks = resp.Result["Peaks"].Values
		expectedFrequencies := make([]float64, 0)
		expectedFrequencies = append(expectedFrequencies, float64(test.txSpec.Frequency))
		for _, f := range test.subCarrierFrequencies {
			expectedFrequencies = append(expectedFrequencies, float64(test.txSpec.Frequency)+f,
				float64(test.txSpec.Frequency)-f)
		}
		removeAtIndex := make([]int, 0)
		allowedFrequencyDeviation := test.txSpec.AllowedFrequencyDeviation
		for _, expectedFreq := range expectedFrequencies {
			for i := 0; i < len(freqs); i++ {
				if freqs[i]-expectedFreq <= allowedFrequencyDeviation {
					removeAtIndex = append(removeAtIndex, i)
				}
			}
		}

		for i := 0; i < len(freqs); i++ {
			if !slices.Contains(removeAtIndex, i) {
				resultData := txSpuriousResult{
					FrequencyKHz: freqs[i] / 1e3,
					LevelDBc:     peaks[i],
					Spec:         spurSpec,
				}
				if header == nil {
					header = resultData.ToHeader()
				}
				spuriousRows = append(spuriousRows, resultData.ToRow())
			}
		}
		test.success(fmt.Sprintf("%d Components found", len(spuriousRows)))
		if len(spuriousRows) == 0 {
			var res txSpuriousResult
			header = res.ToHeader()
			spuriousRows = append(spuriousRows, []string{"-", "-"})
		}
	})

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Spurious Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	test.saToNormalMode(runner)

	if !runner.Describe && runner.Err() == nil {
		time.Sleep(500 * time.Millisecond)

		test.saveResultsToCSV("", header, spuriousRows)
		test.addFinalTestInformation(start)
	}
	return runner.Err()
}
