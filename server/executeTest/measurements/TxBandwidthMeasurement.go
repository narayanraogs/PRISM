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
	executeTest.Register("Bandwidth", "", newTxBandwidthMeasurement)
	results.Register("Bandwidth", results.NewDefaultProcessor([]string{"Results"}))
}

func newTxBandwidthMeasurement() executeTest.Tester {
	var test txBandwidthMeasurement
	return &test
}

type txBandwidthMeasurement struct {
	txBaseTest
}

func (test *txBandwidthMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *txBandwidthMeasurement) DBValidate() error {
	readPowerProfileFunc := func() error {
		return test.readDownlinkPowerProfile(test.test.DownlinkPowerProfileName)
	}
	return test.validateAndPrepare(readPowerProfileFunc)
}

func (test *txBandwidthMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txBandwidthMeasurement) measure(runner *StepRunner) error {
	start := time.Now()

	sa := test.ctx.Selected.SA

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)
	test.setupSAForDownlink(runner, true)

	bwPercent := test.downlinkPowerProfile.OccupiedBW
	freq := test.txSpec.Frequency
	span := test.downlinkProfile.Span
	rbw := float64(test.downlinkProfile.RBW)
	vbw := float64(test.downlinkProfile.VBW)

	var occupiedBWkHz float64

	runner.Run("Measuring Bandwidth using OBW method", false, func() {
		runner.Exec(setOccupiedBW(sa, float64(bwPercent)))
		runner.Exec(setOBWSpectrum(sa, float64(freq), span, rbw, vbw))
		runner.Exec(waitForSweeps(sa, 1))
		resp := runner.Exec(getOccupiedBW(sa))
		if runner.execErr != nil {
			return
		}
		occupiedBWkHz = resp.Result["Bandwidth"].Value / 1000.0
		test.success(fmt.Sprintf("%.4f kHz", occupiedBWkHz))
	})

	runner.Run("Getting Spectrum Dump", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Bandwidth Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{
			Caption:  caption,
			FileData: resp.Result["SpectrumDump"].String,
		})
	})

	test.saToNormalMode(runner)
	runner.Run("Setting SA Continuous Sweep", true, func() {
		_ = runner.Exec(sa.ContinuousSweep)
	})

	if !runner.Describe && runner.Err() == nil {

		resultData := txBandwidthResult{
			CentreFrequencyMHz: float64(test.txSpec.Frequency / 1e6),
			MeasuredBW:         resultValue{Value: occupiedBWkHz, Status: "Success"},
		}

		rows := [][]string{resultData.ToRow()}
		header := resultData.ToHeader()
		test.saveResultsToCSV("", header, rows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
