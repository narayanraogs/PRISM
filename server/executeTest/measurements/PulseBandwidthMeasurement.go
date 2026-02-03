package measurements

import (
	"fmt"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"time"

	"prismServer/utils"
)

func init() {
	executeTest.Register("PulseBandwidth", "SA", newPulseBWMeasurement)
	results.Register("PulseBandwidth", results.NewDefaultProcessor([]string{"Results"}))
}

func newPulseBWMeasurement() executeTest.Tester {
	var test pulseBWMeasurement
	return &test
}

type pulseBWMeasurement struct {
	pulseBaseTest
	peakFreq   float64
	measuredBW float64
}

func (test *pulseBWMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.pulseBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *pulseBWMeasurement) DBValidate() error {
	return test.validateAndPrepare(true, nil)
}

func (test *pulseBWMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"VSA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *pulseBWMeasurement) measure(runner *StepRunner) error {
	start := time.Now()
	vsa := test.ctx.Selected.VSA
	sa := vsa.GetAssociatedSA()

	freq := test.pulseSpec.CenterFrequency
	span := test.downlinkProfile.Span
	rbw := float64(test.downlinkProfile.RBW)
	vbw := float64(test.downlinkProfile.VBW)
	chirpBW := test.pulseSpec.ChirpBandwidth.Float64

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
		runner.Exec(setSpectrum(sa, freq, span, rbw, vbw))
	})

	runner.Run("Setting Nominal Reference Level", true, func() {
		runner.Exec(sa.SetReferenceNominal)
	})

	runner.Run("Getting Peak Frequency", true, func() {
		resp := runner.Exec(sa.GetMaxMarkerValue)
		runner.Exec(sa.SetAllMarkerOff)
		test.peakFreq = resp.Result["MarkerX"].Value

	})

	runner.Run("Getting Bandwidth", true, func() {
		resp := runner.Exec(getPulseBandwidth(vsa, test.peakFreq, test.pulseSpec.CenterFrequency))
		test.measuredBW = resp.Result["Bandwidth"].Value
	})

	runner.Run("Fetching Spectrum", true, func() {
		runner.Exec(waitForSweeps(sa, 2))
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("VSA Downlink Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	if !runner.Describe && runner.Err() == nil {

		resultData := pulseBandwidthResult{
			CentreFrequencyMHz: freq / 1e6,
			SpecBW:             chirpBW / 1e3,
			MeasuredBW:         resultValue{Value: test.measuredBW / 1000, Status: "Success"},
			Deviation:          resultValue{Value: (test.measuredBW - test.pulseSpec.ChirpBandwidth.Float64) / 1000, Status: "Success"},
		}

		rows := [][]string{resultData.ToRow()}
		header := resultData.ToHeader()
		test.saveResultsToCSV("", header, rows)

		test.addFinalTestInformation(start)
	}
	return runner.Err()
}
