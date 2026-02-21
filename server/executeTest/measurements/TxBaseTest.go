package measurements

import (
	"context"
	"fmt"
	"prismServer/utils"
	"time"
)

type txBaseTest struct {
	baseTest
	txRelated
}

func (test *txBaseTest) validateAndPrepare(readExtraData func() error) error {
	// First, call the validation from the true base class
	if err := test.baseTest.DBValidate(); err != nil {
		return err
	}

	// Next, perform the common read operation that you identified
	if err := test.readTxDetails(test.configName, test.config.TxName, test.test.DLProfileName); err != nil {
		return err
	}

	// Then, run the extra, measurement-specific read function if one was provided
	if readExtraData != nil {
		if err := readExtraData(); err != nil {
			return err
		}
	}

	// Finally, handle all the UI boilerplate
	descriptions := test.describe(context.Background())
	test.prepareSteps(descriptions)
	return nil
}

func (test *txBaseTest) Rollback() error {
	return test.baseTest.rollback()
}

func (test *txBaseTest) prepareSteps(descriptions []string) {
	test.ctx.Progress.MeasurementSteps = make([]string, 0)
	test.ctx.Progress.MeasurementSteps = append(test.ctx.Progress.MeasurementSteps, descriptions...)
	test.ctx.Progress.MeasurementStatus = utils.GetRepeatedArray("Queued", len(descriptions))
	test.ctx.Progress.MeasurementValues = utils.GetRepeatedArray("", len(descriptions))

	test.report.AddTestInformation("DL Loss to SA", fmt.Sprintf("%.2f", test.saLoss))
	test.report.AddTestInformation("DL Loss to PM", fmt.Sprintf("%.2f", test.pmLoss))
	test.report.AddTestInformation("Instrument Used", "SA")
	test.report.BatchSize = -1
}

func (test *txBaseTest) setTSMPathForSA(runner *StepRunner) {
	if test.configChanged {
		runner.Run("Setting TSM Path for SA", true, func() {
			tsm := test.ctx.Selected.TSM
			time.Sleep(1 * time.Second)
			runner.Exec(setTSMPath(tsm, test.tsm.DownlinkToSA.String))
		})
	}
}

func (test *txBaseTest) setupSAForDownlink(runner *StepRunner, measureCarrier bool) float64 {
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

func (test *txBaseTest) saToNormalMode(runner *StepRunner) {
	runner.Run("Setting SA Back to Normal Mode", true, func() {
		sa := test.ctx.Selected.SA
		runner.Exec(sa.SetNormalMode)
	})
}
