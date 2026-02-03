package measurements

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/tc"
	"prismServer/utils"
	"strconv"
	"strings"
)

type tpBaseTest struct {
	baseTest
	rxBaseTest
	rxRelated
	txRelated
	tpRelated
	component          string
	zerodBmDifference  float64
	attnSetter         utils.AttenuationTable
	progAttnUsed       bool
	currentPower       float64
	dopplerTimes       []int
	dopplerFrequencies []int
	tm                 rxTM
	profile            database.SpectrumProfile
}

func (test *tpBaseTest) validateAndPrepare(readExtraData func() error) error {
	if err := test.baseTest.DBValidate(); err != nil {
		return err
	}

	if err := test.readTxDetails(test.configName, test.config.TxName, test.test.DLProfileName); err != nil {
		return err
	}

	if err := test.readRxDetails(test.configName, test.config.RxName, test.test); err != nil {
		return err
	}

	if !test.config.CortexIFM.Valid {
		return fmt.Errorf("component is not filled in config table")
	}
	test.component = "IFM-" + strconv.FormatInt(test.config.CortexIFM.Int64, 10)
	if !test.config.IntermediateFrequency.Valid {
		return fmt.Errorf("intermediate frequency is not filled in config table")
	}
	var attnFile string
	if strings.EqualFold(test.config.ProgrammableAttnUsed.String, "yes") {
		attnFile = "/.resources/tsm-" + test.config.RxName.String + ".csv"
		test.progAttnUsed = true
	} else {
		attnFile = "/.resources/gtx-" + test.config.RxName.String + ".csv"
		test.progAttnUsed = false
	}
	attnFile = filepath.Join(utils.Config.BaseFolder, attnFile)

	data, err := os.ReadFile(attnFile)
	if err != nil {
		return fmt.Errorf("attenuation characterization is not done for %s", test.config.RxName.String)
	}
	test.attnSetter = utils.GetAttenuationTable(string(data))

	if readExtraData != nil {
		if err := readExtraData(); err != nil {
			return err
		}
	}

	descriptions := test.describe(context.Background())
	test.prepareSteps(descriptions)
	test.tm.initialize(test.tmtc, test.rxSpec, test.ctx)

	return nil
}

func (test *tpBaseTest) prepareSteps(descriptions []string) {
	test.ctx.Progress.MeasurementSteps = make([]string, 0)
	test.ctx.Progress.MeasurementSteps = append(test.ctx.Progress.MeasurementSteps, descriptions...)
	test.ctx.Progress.MeasurementStatus = utils.GetRepeatedArray("Queued", len(descriptions))
	test.ctx.Progress.MeasurementValues = utils.GetRepeatedArray("", len(descriptions))

	test.report.AddTestInformation("Total Uplink Loss", fmt.Sprintf("%.2f", test.scLoss))
	test.report.AddTestInformation("Uplink Loss with Pad",
		fmt.Sprintf("%.2f", test.scLoss+test.attnSetter.GetFixedPadValue()))
	test.report.AddTestInformation("GTx to SA Loss", fmt.Sprintf("%.2f", test.rxRelated.saLoss))
	test.report.AddTestInformation("DL Loss to SA", fmt.Sprintf("%.2f", test.txRelated.saLoss))
	test.report.AddTestInformation("DL Loss to PM", fmt.Sprintf("%.2f", test.txRelated.pmLoss))
	test.report.BatchSize = -1
}

func (test *tpBaseTest) measureModulationTone(runner *StepRunner) string {
	var measured string
	runner.Run("Measuring Uplink Tone Modulation", false, func() {
		var measurementRequired bool
		var measuredValue string
		var confirmationRequired bool
		switch strings.ToUpper(test.rxSpec.ModulationScheme) {
		case "PM":
			measurementRequired = true
			measuredValue, confirmationRequired = test.rxBaseTest.measureModIndex(runner)
			measured = measuredValue
		case "FM":
			measurementRequired = true
			measuredValue, confirmationRequired = test.measureFrequencyDeviation(runner)
			measured = measuredValue
		default:
			measurementRequired = false
			measured = "NA"
		}
		if !measurementRequired {
			test.success("NA")
			measured = "NA"
			return
		}
		if !confirmationRequired {
			test.success(measuredValue)
			return
		}
		prompt := fmt.Sprintf("Measured value is %s, Press Continue to proceed", measuredValue)
		test.ctx.AskForConfirmation(prompt, 30)
	})
	return measured
}

func (test *tpBaseTest) setupSAForDownlinkForDifferentProfiles(runner *StepRunner) error {
	sa := test.ctx.Selected.SA

	runner.Run("Presetting SA", true, func() {
		runner.Exec(sa.SystemPreset)
	})
	runner.Run("Settting SA Profile", true, func() {
		runner.Exec(setSpectrum(sa, test.profile.CenterFrequency, test.profile.Span,
			float64(test.profile.RBW), float64(test.profile.VBW)))
	})
	runner.Run("Settting Reference Level", true, func() {
		runner.Exec(sa.SetReferenceNominal)
		if runner.execErr != nil {
			return
		}
	})
	return nil
}

func (test *tpBaseTest) measureModIndexTone(runner *StepRunner) (string, bool) {
	gtx := test.ctx.Selected.GTx
	sa := test.ctx.Selected.SA

	runner.Exec(setGTxModIndexTone(gtx, test.component, test.tpRangingSpec.UplinkToneMIOnlyRanging))
	runner.Exec(setGTxModulationOn(gtx, test.component))
	test.executeCommands(runner, 1)
	resp := runner.Exec(getModIndex(sa, test.tpRangingSpec.ToneFrequency))
	if runner.Err() != nil {
		return "0 rad", true
	}
	mi := (resp.Result["modIndexForLeft"].Value + resp.Result["modIndexForRight"].Value) / 2
	diff := math.Abs(test.tpRangingSpec.ToneFrequency - mi)
	tolerance := 0.3 * test.tpRangingSpec.ToneFrequency
	runner.Exec(sa.SetNormalMode)
	runner.Exec(setGTxModulationOff(gtx, test.component))
	confirmationRequired := diff > tolerance
	return fmt.Sprintf("%.2f rad", mi), confirmationRequired
}

func (test *tpBaseTest) executeCommands(runner *StepRunner, noOfCommands int) {
	procdureName := fmt.Sprintf("%s-%dcmds.tst", test.rxSpec.RxName, noOfCommands)
	proc := tc.CreateProcedure(test.rxSpec.RxName, test.tmtc.TestCommandSet, test.tmtc.TestCommandReset, noOfCommands)
	data, err := tc.RunProcedureFromProvider(runner.Ctx, procdureName, proc, test.rxSpec.RxName)
	if err != nil {
		prompt := fmt.Sprintf("Unable to auto execute procedure, Send %d commands manually and press continue", noOfCommands)
		test.ctx.AskForConfirmation(prompt, 0)
	} else {
		p := <-data
		if !p.Success {
			test.failure(p.Err.Error())
			runner.execErr = p.Err
			runner.chainErr = p.Err
		}
	}
}
