package measurements

import (
	"context"
	"fmt"
	"prismServer/utils"
	"time"
)

type pulseBaseTest struct {
	baseTest
	pulseRelated
}

func (test *pulseBaseTest) validateAndPrepare(downlink bool, readExtraData func() error) error {
	if err := test.baseTest.DBValidate(); err != nil {
		return err
	}
	if downlink {
		if err := test.readPulseDetails(
			test.test.ConfigName,
			test.test.TestType,
			test.test.TestCategory,
			test.test.DLProfileName,
			test.test.PulseProfileName.String,
		); err != nil {
			return err
		}
	} else {
		if err := test.readUplinkDetails(
			test.test.ConfigName,
			test.test.TestType,
			test.test.TestCategory,
			test.test.ULProfileName,
			test.test.FrequencyProfileName,
			test.test.PowerProfileName,
		); err != nil {
			return err
		}
	}

	if readExtraData != nil {
		if err := readExtraData(); err != nil {
			return err
		}
	}

	descriptions := test.describe(context.Background())
	test.prepareSteps(descriptions)
	return nil
}

func (test *pulseBaseTest) Rollback() error {
	return test.baseTest.rollback()
}

func (test *pulseBaseTest) prepareSteps(descriptions []string) {
	test.ctx.Progress.MeasurementSteps = make([]string, 0)
	test.ctx.Progress.MeasurementSteps = append(test.ctx.Progress.MeasurementSteps, descriptions...)
	test.ctx.Progress.MeasurementStatus = utils.GetRepeatedArray("Queued", len(descriptions))
	test.ctx.Progress.MeasurementValues = utils.GetRepeatedArray("", len(descriptions))

	test.report.AddTestInformation("DL Loss to VSA", fmt.Sprintf("%.2f", test.saLoss))
	test.report.AddTestInformation("DL Loss to PM", fmt.Sprintf("%.2f", test.pmLoss))
	test.report.AddTestInformation("Instrument Used", "PPM")
	test.report.BatchSize = -1
}

func (test *pulseBaseTest) setTSMPathForPPM(runner *StepRunner) {
	if test.configChanged {
		runner.Run("Setting TSM Path for PPM", true, func() {
			tsm := test.ctx.Selected.TSM
			time.Sleep(1 * time.Second)
			runner.Exec(setTSMPath(tsm, test.tsm.DownlinkToPM.String))
		})
	}
}

func (test *pulseBaseTest) setTSMPathForVSA(runner *StepRunner) {
	if test.configChanged {
		runner.Run("Setting TSM Path for VSA", true, func() {
			tsm := test.ctx.Selected.TSM
			time.Sleep(1 * time.Second)
			runner.Exec(setTSMPath(tsm, test.tsm.DownlinkToSA.String))
		})
	}
}

func (test *pulseBaseTest) readRollback(runner *StepRunner) {
	tsm := test.ctx.Selected.TSM
	runner.Run("Reading TSM State", true, func() {
		time.Sleep(1 * time.Second)
		tsmResp := runner.Exec(tsm.GetDriverPath)
		test.rollbackMap["TSM"] = tsmResp
	})
}

// add a function to check pm mode or ppm mode

func (test *pulseBaseTest) presetPPM(runner *StepRunner) {
	ppm := test.ctx.Selected.PPM
	runner.Exec(ppm.PresetPPM)
}

func (test *pulseBaseTest) setupSAForDownlink(runner *StepRunner, measureCarrier bool) float64 {
	sa := test.ctx.Selected.SA
	var carrierPower float64

	runner.Run("Presetting SA", true, func() {
		runner.Exec(sa.SystemPreset)
	})

	runner.Run("Settting SA Profile", true, func() {
		runner.Exec(setSpectrum(sa, test.downlinkProfile.CenterFrequency, test.downlinkProfile.Span,
			float64(test.downlinkProfile.RBW), float64(test.downlinkProfile.VBW)))
	})

	runner.Run("Settting Reference Level", !measureCarrier, func() {
		resp := runner.Exec(sa.SetReferenceNominal)
		if runner.execErr != nil {
			return
		}
		if measureCarrier {
			carrierPower = resp.Result["ReferenceLevel"].Value - 10
			test.success(fmt.Sprintf("%.2f dB", carrierPower))
		}
	})
	return carrierPower
}
