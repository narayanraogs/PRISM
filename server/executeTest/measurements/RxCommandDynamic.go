package measurements

import (
	"fmt"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"strings"
	"time"
)

func init() {
	executeTest.Register("CommandDynamic", "", newRxCommandDynamic)
	results.Register("CommandDynamic", results.NewDefaultProcessor([]string{"Results"}))
}

func newRxCommandDynamic() executeTest.Tester {
	var test rxCommandDynamic
	return &test
}

type rxCommandDynamic struct {
	rxBaseTest
}

func (test *rxCommandDynamic) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rxBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *rxCommandDynamic) DBValidate() error {
	err := test.validateAndPrepare(test.readFrequencyProfile)
	if err != nil {
		return err
	}
	if strings.Contains(test.testCategory, "Doppler") {
		test.report.AddTestInformation("Doppler Enabled", "true")
		test.report.AddTestInformation("Total Entrues", fmt.Sprintf("%d", len(test.dopplerFrequencies)))
		test.report.AddTestInformation("First Frequency", fmt.Sprintf("%d", test.dopplerFrequencies[0]))
		test.report.AddTestInformation("Last Frequency", fmt.Sprintf("%d", test.dopplerFrequencies[len(test.dopplerFrequencies)-1]))
	}
	return nil
}

func (test *rxCommandDynamic) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "GTX", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *rxCommandDynamic) measure(runner *StepRunner) error {
	sa := test.ctx.Selected.SA
	var rows = make([]rxCommandDynamicResult, 0)
	start := time.Now()

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}
	test.setTSMPathForSA(runner)
	test.removeRFLink(runner)

	test.setupSAForUplink(runner)

	test.computeZerodBmDifference(runner)
	modMeasured := test.measureModulation(runner)
	test.report.AddTestInformation("Modulation Measured", modMeasured)

	if strings.Contains(test.testCategory, "Doppler") {
		test.enableDopplerCompensation(runner)
	}

	measurePowerLevel := true
	if strings.EqualFold(test.rxSpec.ModulationScheme, "CDMA") {
		measurePowerLevel = false
	}
	if strings.Contains(test.testCategory, "Doppler") {
		measurePowerLevel = false
	}

	for i, powerLevel := range test.powerLevels {
		var result rxCommandDynamicResult
		result.receiverPower = powerLevel
		if !strings.EqualFold(test.rxSpec.ModulationScheme, "CDMA") && !strings.EqualFold(test.rxSpec.ModulationScheme, "PM") {
			test.removeRFLink(runner)
		}
		actualPower := test.setPowerLevel(runner, powerLevel, measurePowerLevel)
		result.actualPower = actualPower
		lock, agc := test.uplinkWithModulation(runner, true, i == 0)
		if runner.Err() == nil {
			if !lock {
				result.lockStatus = "UNLOCK"
			} else {
				result.lockStatus = "LOCK"
			}
			result.agcValue = agc
		}

		test.checkForBSLock(runner, true)
		if i == 0 {
			runner.Run("Fetching Spectrum", true, func() {
				resp := runner.Exec(sa.GetSpectrumDump)
				if runner.execErr != nil {
					return
				}
				caption := fmt.Sprintf("Uplink Spectrum for %s", test.configName)
				test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
			})
		}
		if i != len(test.powerLevels)-1 || strings.EqualFold(test.testCategory, "Establish") {
			noOfCmds, ok := test.checkCommandExecution(runner, test.noOfCommandsNominal, true)
			if ok {
				result.cmdsSent = noOfCmds
				result.cmdsExecuted = noOfCmds
			}
		} else {
			noOfCmds, ok := test.checkCommandExecution(runner, test.noOfCommandsAtThreshold, true)
			if ok {
				result.cmdsSent = noOfCmds
				result.cmdsExecuted = noOfCmds
			}
		}
		rows = append(rows, result)
	}
	lastPowerLevel := test.powerLevels[len(test.powerLevels)-1]

	if strings.EqualFold(test.testCategory, "establish") {
		establishRows := test.establish(runner, lastPowerLevel, measurePowerLevel)
		rows = append(rows, establishRows...)
	}

	test.removeRFLink(runner)

	if !runner.Describe && len(rows) > 0 {
		var strRows = make([][]string, 0)
		header := rows[0].ToHeader()
		for _, result := range rows {
			strRows = append(strRows, result.ToRow())
		}
		test.saveResultsToCSV(test.testCategory, header, strRows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}

func (test *rxCommandDynamic) establish(runner *StepRunner, lastPowerLevel float64, measurePowerLevel bool) []rxCommandDynamicResult {
	stepIndex := len(test.ctx.Progress.MeasurementStatus) - 2
	var status string
	var rows = make([]rxCommandDynamicResult, 0)
	runner.Run("Establishing threshold", false, func() {
		repeat := true
		goDown := true
		currentPowerLevel := lastPowerLevel
		noOfCommnads := test.noOfCommandsNominal
		for repeat {
			if goDown {
				currentPowerLevel = currentPowerLevel - 3
				noOfCommnads = test.noOfCommandsNominal
			} else {
				currentPowerLevel = currentPowerLevel + 1
				noOfCommnads = test.noOfCommandsAtThreshold
			}
			status = status + fmt.Sprintf("%.2f: Checking\n", currentPowerLevel)
			test.ctx.Progress.MeasurementValues[stepIndex] = status
			test.ctx.UpdateChannel <- *test.ctx.Progress
			var result rxCommandDynamicResult
			result.receiverPower = currentPowerLevel
			if !strings.EqualFold(test.rxSpec.ModulationScheme, "CDMA") && !strings.EqualFold(test.rxSpec.ModulationScheme, "PM") {
				test.removeRFLinkWithoutRun(runner)
			}
			actualPower := test.setPowerLevelWithoutRun(runner, currentPowerLevel, measurePowerLevel)
			status = status + fmt.Sprintf("Achieved: %.2f\n", actualPower)
			test.ctx.Progress.MeasurementValues[stepIndex] = status
			test.ctx.UpdateChannel <- *test.ctx.Progress
			result.actualPower = actualPower
			lock, agc := test.uplinkWithModulationWithoutRun(runner, false, true)
			if !lock {
				result.lockStatus = "UNLOCK"
				status = status + fmt.Sprintf("%.2f: CAR UNLOCK\n", currentPowerLevel)
				test.ctx.Progress.MeasurementValues[stepIndex] = status
				test.ctx.UpdateChannel <- *test.ctx.Progress
				rows = append(rows, result)
				goDown = false
				continue
			} else {
				result.lockStatus = "LOCK"
				status = status + fmt.Sprintf("%.2f: CAR LOCK\n", currentPowerLevel)
				test.ctx.Progress.MeasurementValues[stepIndex] = status
				test.ctx.UpdateChannel <- *test.ctx.Progress
			}
			result.agcValue = agc

			bsLock := test.checkForBSLockWithoutRun(runner, false)
			if !bsLock {
				result.lockStatus = "BS UNLOCK"
				status = status + fmt.Sprintf("%.2f: BS UNLOCK\n", currentPowerLevel)
				test.ctx.Progress.MeasurementValues[stepIndex] = status
				test.ctx.UpdateChannel <- *test.ctx.Progress
				rows = append(rows, result)
				goDown = false
				continue
			} else {
				result.lockStatus = "LOCK"
				status = status + fmt.Sprintf("%.2f: BS LOCK\n", currentPowerLevel)
				test.ctx.Progress.MeasurementValues[stepIndex] = status
				test.ctx.UpdateChannel <- *test.ctx.Progress
			}
			noOfCmds, ok := test.checkCommandExecutionWithoutRun(runner, noOfCommnads, false)
			if !ok {
				goDown = false
				result.cmdsSent = noOfCommnads
				result.cmdsExecuted = noOfCmds
			} else {
				result.cmdsSent = noOfCmds
				result.cmdsExecuted = noOfCmds
				if !goDown {
					repeat = false
				}
			}
			rows = append(rows, result)
		}
		test.success(status)
	})
	return rows
}
