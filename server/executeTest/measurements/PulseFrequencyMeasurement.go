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
	executeTest.Register("PulseFrequency", "", newPulseFrequencyMeasurement)
	results.Register("PulseFrequency", results.NewDefaultProcessor([]string{"Results"}))
}

func newPulseFrequencyMeasurement() executeTest.Tester {
	var test pulseFrequencyMeasurement
	return &test
}

type pulseFrequencyMeasurement struct {
	pulseBaseTest
}

func (test *pulseFrequencyMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.pulseBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *pulseFrequencyMeasurement) DBValidate() error {
	return test.validateAndPrepare(true, nil)
}

func (test *pulseFrequencyMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"VSA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *pulseFrequencyMeasurement) measure(runner *StepRunner) error {
	start := time.Now()
	vsa := test.ctx.Selected.VSA
	sa := vsa.GetAssociatedSA()

	expectedFreq := test.pulseSpec.CenterFrequency
	span := test.downlinkProfile.Span
	rbw := float64(test.downlinkProfile.RBW)
	vbw := float64(test.downlinkProfile.VBW)
	var freqMeasured float64

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	runner.Run("Setting SA Mode", true, func() {
		runner.Exec(vsa.StartSA)
	})

	test.setTSMPathForVSA(runner)

	runner.Run("Setting Spectrum Parameters", true, func() {
		runner.Exec(sa.SetAlignmentOff)
		runner.Exec(sa.SystemPreset)
		runner.Exec(setSpectrum(sa, expectedFreq, span, rbw, vbw))
	})

	runner.Run("Reading Frequency", false, func() {
		runner.Exec(waitForSweeps(sa, 2))
		resp := runner.Exec(getFrequencyInCounterMode(sa, 1))
		if runner.execErr != nil {
			return
		}
		freqMeasured = (resp.Result["Frequency"].Value) / 1e6
		test.success(fmt.Sprintf("%.6f MHz", freqMeasured))
	})

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Downlink Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	//test.saToNormalMode(runner)

	if !runner.Describe && runner.Err() == nil {
		deviation := (expectedFreq / 1e6) - freqMeasured
		deviationStatus := "Success"
		if math.Abs(deviation) > test.pulseSpec.FrequencyShiftTolerance.Float64 {
			deviationStatus = "Error"
		}

		ppm := (deviation / freqMeasured) * 1e6

		resultData := pulseFrequencyResult{
			SpecificationMHz:    expectedFreq / 1e6,
			MeasuredMHz:         freqMeasured,
			AllowedDeviationKHz: test.pulseSpec.FrequencyShiftTolerance.Float64 / 1e3,
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
