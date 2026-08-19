package measurements

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"prismServer/tc"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type rxBaseTest struct {
	baseTest
	rxRelated
	component             string
	zerodBmDifference     float64
	attnSetter            utils.AttenuationTable
	progAttnUsed          bool
	currentPower          float64
	dopplerTimes          []int
	dopplerFrequencies    []int
	dopplerExtFrequencies []int
	tm                    rxTM
}

func (test *rxBaseTest) validateAndPrepare(readExtraData func() error) error {
	if err := test.baseTest.DBValidate(); err != nil {
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

func (test *rxBaseTest) readFrequencyProfile() error {
	err := test.getFrequencyProfile(test.test.FrequencyProfileName)
	if err != nil {
		return err
	}
	if !test.frequencyProfile.DopplerFile.Valid {
		return nil
	}

	data, err := os.ReadFile(test.frequencyProfile.DopplerFile.String)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	test.dopplerTimes = make([]int, 0)
	test.dopplerFrequencies = make([]int, 0)
	test.dopplerExtFrequencies = make([]int, 0)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		temp := strings.Split(line, ",")
		if len(temp) != 3 {
			return fmt.Errorf("no of columns in doppler profile is not 3")
		}
		t, errTime := strconv.ParseInt(strings.TrimSpace(temp[0]), 10, 64)
		f, errFreq := strconv.ParseInt(strings.TrimSpace(temp[1]), 10, 64)
		e, errExtFreq := strconv.ParseInt(strings.TrimSpace(temp[2]), 10, 64)
		if errTime != nil {
			return errTime
		}
		if errFreq != nil {
			return errFreq
		}
		if errExtFreq != nil {
			return errExtFreq
		}
		test.dopplerTimes = append(test.dopplerTimes, int(t))
		test.dopplerFrequencies = append(test.dopplerFrequencies, int(f))
		test.dopplerExtFrequencies = append(test.dopplerExtFrequencies, int(e))
	}
	return nil
}

func (test *rxBaseTest) prepareSteps(descriptions []string) {
	test.ctx.Progress.MeasurementSteps = make([]string, 0)
	test.ctx.Progress.MeasurementSteps = append(test.ctx.Progress.MeasurementSteps, descriptions...)
	test.ctx.Progress.MeasurementStatus = utils.GetRepeatedArray("Queued", len(descriptions))
	test.ctx.Progress.MeasurementValues = utils.GetRepeatedArray("", len(descriptions))

	test.report.AddTestInformation("Total Uplink Loss", fmt.Sprintf("%.2f", test.scLoss))
	test.report.AddTestInformation("Uplink Loss with Pad",
		fmt.Sprintf("%.2f", test.scLoss+test.attnSetter.GetFixedPadValue()))
	test.report.AddTestInformation("GTx to SA Loss", fmt.Sprintf("%.2f", test.saLoss))
	test.report.BatchSize = -1
}

func (test *rxBaseTest) setTSMPathForSA(runner *StepRunner) {
	if test.configChanged {
		runner.Run("Setting TSM Path for SA", true, func() {
			tsm := test.ctx.Selected.TSM
			time.Sleep(1 * time.Second)
			runner.Exec(setTSMPath(tsm, test.tsm.UplinkToSA.String))
		})
	}
}

func (test *rxBaseTest) Rollback() error {
	if !test.rollbackRequired {
		return nil
	}
	tsm := test.ctx.Selected.TSM
	gtx := test.ctx.Selected.GTx
	setTSMPath(tsm, test.tsm.TerminateUplink.String)
	gtx.SetIdlePatternOff()
	setGTxModulationOff(gtx, test.component)
	setGTxCarrierOff(gtx, test.component)
	return test.baseTest.rollback()
}

func (test *rxBaseTest) removeRFLink(runner *StepRunner) {
	runner.Run("Removing RF Link", true, func() {
		test.removeRFLinkWithoutRun(runner)
	})
}

func (test *rxBaseTest) removeRFLinkWithoutRun(runner *StepRunner) {
	tsm := test.ctx.Selected.TSM
	gtx := test.ctx.Selected.GTx
	runner.Exec(setTSMPath(tsm, test.tsm.TerminateUplink.String))
	runner.Exec(gtx.SetIdlePatternOff)
	runner.Exec(setGTxModulationOff(gtx, test.component))
	runner.Exec(setGTxCarrierOff(gtx, test.component))
	response := runner.Exec(getGTxDopplerCompensation(gtx, test.component))
	if !response.Success {
		return
	}
	if response.Result["Doppler"].Integer != 0 {
		test.setDopplerCompensation(runner, 5, []int{0, 0}, []int{0, 0}, []int{0, 5000})
		runner.Exec(setGTxDopplerCompensationDisable(gtx))
		runner.Exec(setGTxStopSweep(gtx, test.component))
	}
	test.setIntermediateFrequency(runner, 0)
}

func (test *rxBaseTest) enableDopplerCompensation(runner *StepRunner) {
	runner.Run("Enabling Doppler Compenstation", true, func() {
		test.setDopplerCompensation(runner, 30, test.dopplerFrequencies, test.dopplerExtFrequencies, test.dopplerTimes)
	})
}

func (test *rxBaseTest) setDopplerCompensation(runner *StepRunner, timeOffset int, freqs []int, extFreqs []int, times []int) {
	gtx := test.ctx.Selected.GTx
	runner.Exec(setGTxDopplerCompensationTable(gtx, timeOffset, freqs, extFreqs, times))
	runner.Exec(setGTxEnableDoppler(gtx, test.component))
	runner.Exec(setGTxDopplerCompensationEnable(gtx))
}

func (test *rxBaseTest) setupSAForUplink(runner *StepRunner) {
	sa := test.ctx.Selected.SA

	runner.Run("Presetting SA", true, func() {
		runner.Exec(sa.SystemPreset)
	})

	runner.Run("Settting SA Profile", true, func() {
		runner.Exec(setSpectrum(sa, test.uplinkProfile.CenterFrequency, test.uplinkProfile.Span,
			float64(test.uplinkProfile.RBW), float64(test.uplinkProfile.VBW)))
	})

}

func (test *rxBaseTest) computeZerodBmDifference(runner *StepRunner) {
	sa := test.ctx.Selected.SA
	gtx := test.ctx.Selected.GTx

	runner.Run("Computing the difference in zero dBm Power", false, func() {
		runner.Exec(test.setAttenuationValue(runner, 0))
		runner.Exec(setGTxModulationOff(gtx, test.component))
		runner.Exec(setGTxCarrierOff(gtx, test.component))
		runner.Exec(setGTxOnlyTC(gtx, test.component))
		runner.Exec(setGTxPower(gtx, test.component, 0.0))
		runner.Exec(setGTxCarrierOn(gtx, test.component))
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
		test.success(fmt.Sprintf("Measured: %.2f, Difference: %.2f", powerMeasured, test.zerodBmDifference))
		test.currentPower = powerMeasured
	})
}

func (test *rxBaseTest) setAttenuationValue(runner *StepRunner, attn float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		tsm := test.ctx.Selected.TSM
		gtx := test.ctx.Selected.GTx
		//attn should be positive and power should be negative
		if test.progAttnUsed && attn < 0 {
			attn = attn * -1
		}
		if !test.progAttnUsed && attn > 0 {
			attn = attn * -1
		}
		set, fixed := test.attnSetter.GetValueToBeSet(attn)
		if fixed && test.progAttnUsed {
			runner.Exec(setTSMPath(tsm, test.tsm.IncludePad.String))
		}
		if test.progAttnUsed && !fixed {
			runner.Exec(setTSMPath(tsm, test.tsm.ExcludePad.String))
		}
		if test.progAttnUsed {
			tsm.SetAttn(int(test.tsm.AttnNumber.Int64), set)
		} else {
			runner.Exec(setGTxPower(gtx, test.component, attn))
		}
		response := utils.CommandResponse{
			Success: true,
			Result:  make(map[string]utils.CommandResult),
		}
		response.Result["Fixed Pad"] = utils.CommandResult{
			Bool: fixed,
		}
		response.Result["Attn Set"] = utils.CommandResult{
			Value: attn,
		}
		return response
	}
}

func (test *rxBaseTest) setPowerLevel(runner *StepRunner, powerLevel float64, toBeMeasured bool) float64 {

	var powerAtRx float64
	runner.Run(fmt.Sprintf("Setting the Power Level: %.2f", powerLevel), false, func() {
		powerAtRx = test.setPowerLevelWithoutRun(runner, powerLevel, toBeMeasured)
		if runner.Err() == nil {
			test.success(fmt.Sprintf("%.2f dBm", powerAtRx))
		}
	})
	return powerAtRx
}

func (test *rxBaseTest) setPowerLevelWithoutRun(runner *StepRunner, powerLevel float64, toBeMeasured bool) float64 {
	sa := test.ctx.Selected.SA
	gtx := test.ctx.Selected.GTx
	var powerAtRx float64
	attn := powerLevel + test.scLoss - test.zerodBmDifference
	respAttn := runner.Exec(test.setAttenuationValue(runner, attn))
	runner.Exec(setGTxCarrierOn(gtx, test.component))
	if toBeMeasured {
		resp := runner.Exec(sa.SetReferenceNominal)
		if runner.execErr != nil {
			return 0.0
		}
		saPower := resp.Result["ReferenceLevel"].Value - 10
		fixed := respAttn.Result["Fixed Pad"].Bool
		powerAtRx = saPower - test.scLoss + test.saLoss
		if fixed {
			powerAtRx = powerAtRx - test.attnSetter.GetFixedPadValue()
		}
		if powerAtRx > test.rxSpec.MaxPower ||
			math.Abs(powerAtRx-powerLevel) > utils.Config.TestRelated.RxPowerLevelTolerance {
			test.failure("Power to be set is outside receiver spec/tolerance")
			err := fmt.Errorf("power to be set is outside receiver spec/tolerance")
			runner.execErr = err
			runner.chainErr = err
			return 0.0
		}
	} else {
		powerAtRx = powerLevel
	}
	return powerAtRx
}

func (test *rxBaseTest) uplinkWithoutModulation(runner *StepRunner, raiseError bool) (bool, float64) {
	var lock bool
	var agc float64

	runner.Run("Uplinking without modulation and getting lock status", false, func() {
		lock, agc = test.uplinkWithoutModulationWithoutRun(runner, raiseError)
		if runner.Err() == nil {
			test.success(fmt.Sprintf("AGC: %.2f", agc))
		}
	})
	return lock, agc
}

func (test *rxBaseTest) uplinkWithoutModulationWithoutRun(runner *StepRunner, raiseError bool) (bool, float64) {
	gtx := test.ctx.Selected.GTx
	tsm := test.ctx.Selected.TSM
	var lock bool
	var agc float64
	runner.Exec(setGTxIdleOff(gtx))
	runner.Exec(setGTxModulationOff(gtx, test.component))
	runner.Exec(setGTxCarrierOn(gtx, test.component))
	runner.Exec(setTSMPath(tsm, test.tsm.UplinkToSC.String))
	test.sweep(runner)
	if runner.execErr != nil {
		return lock, agc
	}
	if !raiseError {
		time.Sleep(10 * time.Second)
	}
	lock, agc = test.tm.getLockAndAGCValue()
	if !lock && raiseError {
		test.failure("Receiver did not Lock")
		err := fmt.Errorf("receiver did not lock")
		runner.execErr = err
		runner.chainErr = err
		return lock, agc
	}
	return lock, agc
}

func (test *rxBaseTest) sweep(runner *StepRunner) {
	gtx := test.ctx.Selected.GTx
	if !strings.EqualFold(test.rxSpec.ModulationScheme, "PM") {
		return
	}
	runner.Exec(setGTxStopSweep(gtx, test.component))
	runner.Exec(setGTxStartSweep(gtx, test.component))
	sleepTime := test.rxSpec.SweepRange.Float64 * 4 / test.rxSpec.SweepRate.Float64
	runner.Wait(int(math.Ceil(sleepTime)))
	runner.Exec(setGTxStopSweep(gtx, test.component))
}

func (test *rxBaseTest) executeCommands(runner *StepRunner, noOfCommands int) {
	procdureName := fmt.Sprintf("%s-%dcmds.tst", test.rxSpec.RxName, noOfCommands)
	proc := tc.CreateProcedure(test.rxSpec.RxName, test.tmtc.TestCommandSet, test.tmtc.TestCommandReset, noOfCommands)
	data, err := tc.RunProcedureFromProvider(runner.Ctx, procdureName, proc, test.rxSpec.RxName)
	if err != nil {
		prompt := fmt.Sprintf("Unable to auto execute procedure, Send %d commands manually and press continue", noOfCommands)
		test.ctx.AskForConfirmation(prompt, 0)
	} else {
		p := <-data
		if !p.Success {
			prompt := fmt.Sprintf("Unable to confirm execution of %d commands, Please send %d commands manually and press continue", noOfCommands, noOfCommands)
			test.ctx.AskForConfirmation(prompt, 0)
		}
	}
}

func (test *rxBaseTest) checkCommandExecution(runner *StepRunner, noOfCommands int, raiseError bool) (int, bool) {
	var noOfCommandsExecuted int
	var success bool
	runner.Run(fmt.Sprintf("Executing %d commands", noOfCommands), true, func() {
		noOfCommandsExecuted, success = test.checkCommandExecutionWithoutRun(runner, noOfCommands, raiseError)
	})
	return noOfCommandsExecuted, success
}

func (test *rxBaseTest) checkCommandExecutionWithoutRun(runner *StepRunner, noOfCommands int, raiseError bool) (int, bool) {
	var noOfCommandsExecuted int
	var success bool
	commandCountInitial := test.tm.getCommandCounter()
	test.executeCommands(runner, noOfCommands)
	if runner.Err() != nil {
		return noOfCommandsExecuted, success
	}
	commandCounterFinal := test.tm.getCommandCounter()
	commandExecuted := commandCounterFinal - commandCountInitial
	if commandExecuted >= noOfCommands {
		noOfCommandsExecuted = noOfCommands
		success = true
		return noOfCommandsExecuted, success
	}
	prompt := fmt.Sprintf("Unable to confirm that %d commands are executed. Please enter no of commands executed", noOfCommands)
	noOfCommandsStr := test.ctx.AskForInput(prompt, "", 0)
	noOfCommandsInt, _ := strconv.Atoi(noOfCommandsStr)
	noOfCommandsExecuted = noOfCommandsInt
	success = noOfCommandsExecuted >= noOfCommands
	if !success && raiseError {
		test.failure("Unable to execute commands")
		runner.execErr = fmt.Errorf("unable to execute commands")
		runner.chainErr = fmt.Errorf("unable to execute commands")
	}
	return noOfCommandsExecuted, success
}

func (test *rxBaseTest) uplinkWithModulation(runner *StepRunner, raiseError bool, firstTime bool) (bool, float64) {
	var lock bool
	var agc float64

	runner.Run("Uplinking with modulation and getting lock status", false, func() {
		lock, agc = test.uplinkWithModulationWithoutRun(runner, raiseError, firstTime)
		if runner.Err() == nil {
			test.success(fmt.Sprintf("AGC: %.2f", agc))
		}
	})
	return lock, agc
}

func (test *rxBaseTest) uplinkWithModulationWithoutRun(runner *StepRunner, raiseError bool, firstTime bool) (bool, float64) {
	var lock bool
	var agc float64
	gtx := test.ctx.Selected.GTx
	tsm := test.ctx.Selected.TSM

	switch strings.ToLower(test.rxSpec.ModulationScheme) {
	case "cdma":
		runner.Exec(setGTxModulationOn(gtx, test.component))
	default:
		runner.Exec(setGTxModulationOff(gtx, test.component))
	}
	runner.Exec(setGTxIdleOff(gtx))
	runner.Exec(setGTxCarrierOn(gtx, test.component))
	if strings.Contains(test.testCategory, "Doppler") {
		var repeat = true
		for repeat {
			response := runner.Exec(getGTxDopplerCompensation(gtx, test.component))
			if !response.Success {
				return false, 0.0
			}
			if response.Result["Doppler"].Integer == 2 {
				repeat = false
			} else {
				runner.Wait(1)
			}
		}
	}
	runner.Exec(setTSMPath(tsm, test.tsm.UplinkToSC.String))

	if firstTime {
		test.sweep(runner)
	}
	if runner.execErr != nil {
		return lock, agc
	}
	time.Sleep(1 * time.Second)
	masterFrameTime := utils.Config.TestRelated.MasterFrameTimeSecs
	deadline := time.Now().Add(time.Duration(masterFrameTime) * time.Second)
	for {
		lock, agc = test.tm.getLockAndAGCValue()
		if lock {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		select {
		case <-runner.Ctx.Done():
			runner.execErr = fmt.Errorf("user aborted")
			runner.chainErr = runner.execErr
			return lock, agc
		case <-time.After(1 * time.Second):
		}
	}
	if !lock && raiseError {
		test.failure("Receiver did not Lock")
		err := fmt.Errorf("receiver did not lock")
		runner.execErr = err
		runner.chainErr = err
		return lock, agc
	}
	return lock, agc

}

func (test *rxBaseTest) checkForBSLock(runner *StepRunner, raiseError bool) bool {
	var bsLock bool
	runner.Run("Executing command and Checking for BS Lock Stats", false, func() {
		bsLock = test.checkForBSLockWithoutRun(runner, raiseError)
		if runner.Err() == nil && bsLock {
			test.success(fmt.Sprintf("LOCK"))
		}
	})
	return bsLock
}

func (test *rxBaseTest) checkForBSLockWithoutRun(runner *StepRunner, raiseError bool) bool {
	var bsLock bool

	runner.Exec(setGTxModulationOn(test.ctx.Selected.GTx, test.component))
	test.executeCommands(runner, 1)
	if runner.Err() != nil {
		return bsLock
	}
	time.Sleep(1 * time.Second)
	masterFrameTime := utils.Config.TestRelated.MasterFrameTimeSecs
	deadline := time.Now().Add(time.Duration(masterFrameTime) * time.Second)
	for {
		bsLock = test.tm.checkRxBitSyncLock()
		if bsLock {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		select {
		case <-runner.Ctx.Done():
			runner.execErr = fmt.Errorf("user aborted")
			runner.chainErr = runner.execErr
			return bsLock
		case <-time.After(1 * time.Second):
		}
	}
	if !bsLock && raiseError {
		test.failure("Bit Sync did not Lock")
		err := fmt.Errorf("bit sync did not lock")
		runner.execErr = err
		runner.chainErr = err
	}
	return bsLock
}

func (test *rxBaseTest) measureModulation(runner *StepRunner) string {
	var measured string
	runner.Run("Measuring Uplink Modulation", false, func() {
		var measurementRequired bool
		var measuredValue string
		var confirmationRequired bool
		switch strings.ToUpper(test.rxSpec.ModulationScheme) {
		case "PM":
			measurementRequired = true
			measuredValue, confirmationRequired = test.measureModIndex(runner)
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
		confirm := test.ctx.AskForConfirmation(prompt, 30)
		if confirm {
			test.success(measuredValue)
		}
	})
	return measured
}

func (test *rxBaseTest) measureModIndex(runner *StepRunner) (string, bool) {
	gtx := test.ctx.Selected.GTx
	sa := test.ctx.Selected.SA

	runner.Exec(setGTxModIndexTC(gtx, test.component, test.rxSpec.TCModIndex.Float64))
	runner.Exec(setGTxModulationOn(gtx, test.component))
	test.executeCommands(runner, 1)
	resp := runner.Exec(getModIndex(sa, test.rxSpec.TCSubCarrierFrequency))
	if runner.Err() != nil {
		return "0 rad", true
	}
	mi := (resp.Result["modIndexForLeft"].Value + resp.Result["modIndexForRight"].Value) / 2
	diff := math.Abs(test.rxSpec.TCModIndex.Float64 - mi)
	tolerance := 0.3 * test.rxSpec.TCModIndex.Float64
	runner.Exec(sa.SetNormalMode)
	runner.Exec(setGTxModulationOff(gtx, test.component))
	confirmationRequired := diff > tolerance
	return fmt.Sprintf("%.2f rad", mi), confirmationRequired
}

func (test *rxBaseTest) measureFrequencyDeviation(runner *StepRunner) (string, bool) {
	gtx := test.ctx.Selected.GTx
	sa := test.ctx.Selected.SA

	runner.Exec(setGTxFrequencyDeviationTC(gtx, test.component, test.rxSpec.FrequencyDeviationFM.Float64))
	runner.Exec(setGTxModulationOn(gtx, test.component))
	test.executeCommands(runner, 1)
	resp := runner.Exec(getSAFrequencyDeviation(sa, float64(test.rxSpec.Frequency)))
	if runner.Err() != nil {
		return "0 kHz", true
	}
	tolerance := 8 * test.rxSpec.TCSubCarrierFrequency
	freqDev := resp.Result["FrequencyDeviation"].Value
	diff := math.Abs(freqDev - test.rxSpec.FrequencyDeviationFM.Float64*2)
	runner.Exec(sa.SetNormalMode)
	runner.Exec(setGTxModulationOff(gtx, test.component))
	confirmationRequired := diff > tolerance
	return fmt.Sprintf("%.2f kHz", freqDev/1000), confirmationRequired
}

func (test *rxBaseTest) setIntermediateFrequency(runner *StepRunner, offset float64) {
	gtx := test.ctx.Selected.GTx
	rfFreq := float64(test.rxSpec.Frequency)
	freq := float64(test.config.IntermediateFrequency.Int64)
	runner.Exec(setGTxIntermediateFrequency(gtx, test.component, freq+offset))
	if strings.EqualFold(test.rxSpec.ModulationScheme, "CDMA") {
		offsetChipRate := (offset / rfFreq) * test.rxSpec.CodeRateInMcps.Float64
		chipRateNew := test.rxSpec.CodeRateInMcps.Float64 + offsetChipRate
		runner.Exec(setGTxIdleOff(gtx))
		runner.Exec(setGTxChipRateDSSS(gtx, chipRateNew))
	}
}
