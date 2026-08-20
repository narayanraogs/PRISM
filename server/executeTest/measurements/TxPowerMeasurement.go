package measurements

import (
	"fmt"
	"math"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"strings"
	"time"
)

func init() {
	executeTest.Register("Power", "Non-Coherent", newTxPowerMeasurement)
	executeTest.Register("Power", "Coherent", newTxPowerMeasurement)
	results.Register("Power", results.NewDefaultProcessor([]string{"Results"}))
}

func newTxPowerMeasurement() executeTest.Tester {
	var test txPowerMeasurement
	return &test
}

type txPowerMeasurement struct {
	txBaseTest
}

func (test *txPowerMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *txPowerMeasurement) DBValidate() error {
	readPowerProfileFunc := func() error {
		return test.readDownlinkPowerProfile(test.test.DownlinkPowerProfileName)
	}
	return test.validateAndPrepare(readPowerProfileFunc)
}

func (test *txPowerMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM", "PM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txPowerMeasurement) measure(runner *StepRunner) error {
	start := time.Now()

	sa := test.ctx.Selected.SA
	tsm := test.ctx.Selected.TSM
	pm := test.ctx.Selected.PM

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)
	test.setupSAForDownlink(runner, false)

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Downlink Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	runner.Run("Setting TSM Path for PM", true, func() {
		runner.Exec(setTSMPath(tsm, test.tsm.DownlinkToPM.String))
	})

	channel := test.downlinkPowerProfile.PMChannel
	var power float64

	frequency := float64(test.txSpec.Frequency)
	runner.Run("Reading power from PM", false, func() {
		if strings.EqualFold(channel, "A") {
			runner.Exec(setFrequencyChA(pm, frequency))
			if runner.execErr != nil {
				return
			}
			resp := runner.Exec(readPowerChannelA(pm))
			if runner.execErr != nil {
				return
			}
			power = resp.Result["Power"].Value
		} else {
			runner.Exec(setFrequencyChB(pm, frequency))
			if runner.execErr != nil {
				return
			}
			resp := runner.Exec(readPowerChannelB(pm))
			if runner.execErr != nil {
				return
			}
			power = resp.Result["Power"].Value
		}
		test.success(fmt.Sprintf("%.2f dBm", power))
	})

	computedPower := power + test.pmLoss
	deviation := computedPower - test.txSpec.Power
	deviationStatus := "Success"
	if math.Abs(deviation) > test.txSpec.AllowedPowerDevaition {
		deviationStatus = "Error"
	}

	runner.Run("Setting TSM Path for SA", true, func() {
		runner.Exec(setTSMPath(tsm, test.tsm.DownlinkToSA.String))
	})

	bwPercent := test.downlinkPowerProfile.OccupiedBW
	freq := test.txSpec.Frequency
	span := test.downlinkProfile.Span
	rbw := test.downlinkProfile.RBW
	vbw := test.downlinkProfile.VBW

	var obwPower float64
	var occupiedBW float64

	runner.Run("Measuring Power in Occupied Bandwidth Mode", false, func() {
		runner.Exec(setOccupiedBW(sa, float64(bwPercent)))
		runner.Exec(setOBWSpectrum(sa, float64(freq), span, float64(rbw), float64(vbw)))
		runner.Exec(waitForSweeps(sa, 1))
		resp := runner.Exec(getOccupiedBW(sa))
		obwPower = resp.Result["Power"].Value
		occupiedBW = resp.Result["Bandwidth"].Value / 1000
		test.success(fmt.Sprintf("%.2f dBm", obwPower))
		test.report.AddTestInformation("Occupied Bandwidth", fmt.Sprintf("%0.2f kHz", occupiedBW))
	})

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Downlink Spectrum for %s in OBW Mode", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	test.saToNormalMode(runner)

	if !runner.Describe && runner.Err() == nil {

		resultData := txPowerResult{
			SpecifiedDBm:  test.txSpec.Power,
			MeasuredDBm:   resultValue{Value: computedPower, Status: deviationStatus},
			DeviationDB:   resultValue{Value: deviation, Status: deviationStatus},
			PMReadingDBm:  power,
			SAOBWPowerDBm: obwPower,
		}

		rows := [][]string{resultData.ToRow()}
		header := resultData.ToHeader()
		test.saveResultsToCSV("", header, rows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
