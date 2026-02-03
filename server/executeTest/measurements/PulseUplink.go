package measurements

import (
	"fmt"
	"math"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/utils"
	"time"
)

func init() {
	executeTest.Register("PulseUplink", "", newPulseFrequencyMeasurement)
	results.Register("PulseUplink", results.NewDefaultProcessor([]string{"Results"}))
}

func newPulseUplinkMeasurement() executeTest.Tester {
	var test pulseUplinkMeasurement
	return &test
}

type pulseUplinkMeasurement struct {
	pulseBaseTest
	zerodBmDifference float64
}

func (test *pulseUplinkMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.pulseBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *pulseUplinkMeasurement) DBValidate() error {
	return test.validateAndPrepare(false, nil)
}

func (test *pulseUplinkMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"VSA", "TSM", "SG"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *pulseUplinkMeasurement) measure(runner *StepRunner) error {
	start := time.Now()
	vsa := test.ctx.Selected.VSA
	sa := vsa.GetAssociatedSA()
	sg := test.ctx.Selected.SG
	tsm := test.ctx.Selected.TSM

	results := make([]pulseUplink, 0)

	startFrequency := test.uplinkFrequency.MaxFrequency.Float64 * -1
	stopFrequency := test.uplinkFrequency.MaxFrequency.Float64
	stepSize := test.uplinkFrequency.StepSize.Float64

	freq := test.pulseSpec.CenterFrequency
	span := test.downlinkProfile.Span
	rbw := float64(test.downlinkProfile.RBW)
	vbw := float64(test.downlinkProfile.VBW)

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	runner.Run("Setting SA Mode", true, func() {
		runner.Exec(vsa.StartSA)
	})

	runner.Run("Setting TSM Path for Pulse Uplink", true, func() {
		runner.Exec(setTSMPath(tsm, test.tsm.UplinkToSA.String))
	})

	runner.Run("Setting Spectrum Parameters", true, func() {
		runner.Exec(sa.SetAlignmentOff)
		runner.Exec(sa.SystemPreset)
		runner.Exec(setSpectrum(sa, freq, span, rbw, vbw))
	})

	runner.Run("Setting Power and Frequency in SA", true, func() {
		runner.Exec(setSGFrequency(sg, freq))
		runner.Exec(sg.SetModOff)
		runner.Exec(setSGPower(sg, 0))
	})

	runner.Run("Measuring zero dBm differenceSA", true, func() {
		runner.Exec(sg.SetRFOn)
		resp := runner.Exec(sa.SetReferenceNominal)
		if runner.execErr != nil {
			return
		}
		powerMeasured := resp.Result["ReferenceLevel"].Value - 10
		zerodBMLoss := test.saLoss * -1
		test.zerodBmDifference = powerMeasured - zerodBMLoss
		if math.Abs(test.zerodBmDifference) > 3 {
			runner.SetError(fmt.Errorf("zero dBm > 3 dB: %.2f", test.zerodBmDifference))
			return
		}
		if math.Abs(test.zerodBmDifference) > utils.Config.TestRelated.ZerodBmTolerence {
			prompt := fmt.Sprintf("Zero dBm greater than allowed Tolerance, Difference = %.2f, Press Continue to Continue", test.zerodBmDifference)
			test.ctx.AskForConfirmation(prompt, 30)
		}
	})

	for _, power := range test.uplinkPowers {
		for frequency := startFrequency; frequency <= stopFrequency; frequency = frequency + stepSize {

			runner.Run(fmt.Sprintf("Setting power at %.2 dBm and frequency at %.2f GHz", power, frequency/1e9), false, func() {
				var result pulseUplink
				result.SetFrequency = frequency
				result.ExpectedPower = power
				powerSet := power + test.scLoss - test.zerodBmDifference
				runner.Exec(setSGPower(sg, powerSet))
				runner.Exec(setSGFrequency(sg, freq+frequency))
				runner.Exec(setSpectrum(sa, freq+frequency, span, rbw, vbw))
				resp := runner.Exec(sa.SetReferenceNominal)
				if runner.execErr != nil {
					return
				}
				powerMeasured := resp.Result["ReferenceLevel"].Value - 10
				result.MeasuredPower = powerMeasured
				powerAtSC := powerMeasured - test.scLoss + test.saLoss
				runner.Exec(setTSMPath(tsm, test.tsm.UplinkToSC.String))
				prompt := "Press continue after issuing Map ON"
				test.ctx.AskForConfirmation(prompt, 0)
				prompt = "Press continue after issuing Map OFF"
				test.ctx.AskForConfirmation(prompt, 0)

				runner.Exec(setTSMPath(tsm, test.tsm.TerminateUplink.String))
				results = append(results, result)
				test.success(fmt.Sprintf("%.2f dBm", powerAtSC))
			})

		}
	}

	runner.Run("SettingRF Off", true, func() {
		runner.Exec(sg.SetRFOff)
	})

	if !runner.Describe && runner.Err() == nil {
		var rows = make([][]string, 0)
		header := results[0].ToHeader()
		for _, result := range results {
			rows = append(rows, result.ToRow())
		}
		test.saveResultsToCSV("", header, rows)
		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
