package measurements

import (
	"fmt"
	"math"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/utils"
	"strings"
	"time"
)

func init() {
	executeTest.Register("LockDynamic", "", newRxLockDynamic)
	results.Register("LockDynamic", results.NewDefaultProcessor([]string{"Results"}))
}

func newRxLockDynamic() executeTest.Tester {
	var test rxLockDynamic
	return &test
}

type rxLockDynamic struct {
	rxBaseTest
}

func (test *rxLockDynamic) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rxBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *rxLockDynamic) DBValidate() error {
	return test.validateAndPrepare(nil)
}

func (test *rxLockDynamic) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "GTX", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *rxLockDynamic) measure(runner *StepRunner) error {
	var rows = make([]rxLockDynamicResult, 0)
	start := time.Now()

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}
	test.setTSMPathForSA(runner)
	test.removeRFLink(runner)

	test.setupSAForUplink(runner)

	test.computeZerodBmDifference(runner)

	for _, powerLevel := range test.powerLevels {
		var result rxLockDynamicResult
		result.receiverPower = powerLevel
		test.removeRFLink(runner)
		actualPower := test.setPowerLevel(runner, powerLevel, true)
		result.actualPower = actualPower
		lock, agc := test.uplinkWithoutModulation(runner, true)
		if runner.Err() == nil {
			if !lock {
				result.lockStatus = "UNLOCK"
			} else {
				result.lockStatus = "LOCK"
			}
			result.agcValue = agc
		}
		rows = append(rows, result)
	}
	lastPowerLevel := test.powerLevels[len(test.powerLevels)-1]

	if strings.EqualFold(test.testCategory, "establish") {
		establishRows := test.establish(runner, lastPowerLevel)
		rows = append(rows, establishRows...)
	}

	runner.Run("Continuous sweep monitor for Last level", false, func() {

		noOfSweeps := utils.Config.TestRelated.SweepsForSustainedLock
		timeOut := float64(noOfSweeps) * test.rxSpec.SweepRange.Float64 * 4 / test.rxSpec.SweepRate.Float64
		timeOut = math.Ceil(timeOut)
		var tempChan = make(chan bool, 1)
		go test.tm.getContinuousLockMonitor(int(timeOut), tempChan)
		for i := 0; i < noOfSweeps; i++ {
			test.sweep(runner)
		}
		locked := <-tempChan
		if !locked {
			test.failure("Lock Failed")
			err := fmt.Errorf("sustained lock failed")
			runner.execErr = err
			runner.chainErr = err
			return
		}
		test.success("Locked")
	})

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

func (test *rxLockDynamic) establish(runner *StepRunner, lastPowerLevel float64) []rxLockDynamicResult {
	stepIndex := len(test.ctx.Progress.MeasurementStatus) - 2
	var status string
	var rows = make([]rxLockDynamicResult, 0)
	runner.Run("Establishing threshold", false, func() {
		repeat := true
		goDown := true
		currentPowerLevel := lastPowerLevel
		for repeat {
			if goDown {
				currentPowerLevel = currentPowerLevel - 3
			} else {
				currentPowerLevel = currentPowerLevel + 1
			}
			status = status + fmt.Sprintf("%.2f: Checking\n", currentPowerLevel)
			test.ctx.Progress.MeasurementValues[stepIndex] = status
			test.ctx.UpdateChannel <- *test.ctx.Progress
			var result rxLockDynamicResult
			result.receiverPower = currentPowerLevel
			test.removeRFLinkWithoutRun(runner)
			actualPower := test.setPowerLevelWithoutRun(runner, currentPowerLevel, true)
			status = status + fmt.Sprintf("Achieved: %.2f\n", actualPower)
			test.ctx.Progress.MeasurementValues[stepIndex] = status
			test.ctx.UpdateChannel <- *test.ctx.Progress
			result.actualPower = actualPower
			lock, agc := test.uplinkWithoutModulationWithoutRun(runner, false)
			if !lock {
				result.lockStatus = "UNLOCK"
				status = status + fmt.Sprintf("%.2f: UNLOCK\n", currentPowerLevel)
				test.ctx.Progress.MeasurementValues[stepIndex] = status
				test.ctx.UpdateChannel <- *test.ctx.Progress
				goDown = false
			} else {
				result.lockStatus = "LOCK"
				status = status + fmt.Sprintf("%.2f: LOCK\n", currentPowerLevel)
				test.ctx.Progress.MeasurementValues[stepIndex] = status
				test.ctx.UpdateChannel <- *test.ctx.Progress
				if !goDown {
					repeat = false
				}
			}
			result.agcValue = agc
			rows = append(rows, result)
		}
		test.success(status)
	})
	return rows
}
