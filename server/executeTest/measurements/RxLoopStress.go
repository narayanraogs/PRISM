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
	executeTest.Register("LoopStress", "", newRxLoopStress)
	results.Register("LoopStress", results.NewDefaultProcessor([]string{"Results"}))
}

func newRxLoopStress() executeTest.Tester {
	var test rxLoopStress
	return &test
}

type rxLoopStress struct {
	rxBaseTest
}

func (test *rxLoopStress) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rxBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *rxLoopStress) DBValidate() error {
	return test.validateAndPrepare(test.readFrequencyProfile)
}

func (test *rxLoopStress) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "GTX", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *rxLoopStress) measure(runner *StepRunner) error {
	var rows = make([]rxLoopStressResult, 0)
	start := time.Now()

	gtx := test.ctx.Selected.GTx

	sweepStep := test.frequencyProfile.StepSize.Float64
	if strings.EqualFold(test.testCategory, "Extreme") {
		sweepStep = test.frequencyProfile.MaxFrequency.Float64
	}

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}
	test.setTSMPathForSA(runner)
	test.removeRFLink(runner)

	test.setupSAForUplink(runner)

	test.computeZerodBmDifference(runner)

	actualPower := test.setPowerLevel(runner, test.powerLevels[0], true)
	test.report.AddTestInformation("Power at Rx", fmt.Sprintf("%.2f dBm", actualPower))
	runner.Run("Setting Sweep Parameters", true, func() {
		runner.Exec(setGTxStopSweep(gtx, test.component))
		runner.Exec(setGtxSweepRange(gtx, test.component, test.rxSpec.SweepRange.Float64))
		runner.Exec(setGtxSweepRate(gtx, test.component, test.rxSpec.SweepRate.Float64))
		runner.Exec(setGtxSweepStep(gtx, test.component, sweepStep))
	})
	test.uplinkWithoutModulation(runner, true)
	runner.Run("Getting Loop Stress count for 0 Hz", false, func() {
		_, loop := test.tm.getLoopStressCount()
		rows = append(rows, rxLoopStressResult{0.0, loop})
		test.success(fmt.Sprintf("%.2f", loop))
	})

	waitTime := (sweepStep / test.rxSpec.SweepRate.Float64) + 0.2
	waitDuration := int(math.Ceil(waitTime))
	noOfStepsInOneDirection := int(test.rxSpec.SweepRange.Float64 / sweepStep)
	offset := 0.0

	for i := 0; i < noOfStepsInOneDirection; i++ {
		//positive loop
		offset = offset + sweepStep
		runner.Run(fmt.Sprintf("Reading Loop Stress value for %.2f kHz", offset/1000), false, func() {
			runner.Exec(setGtxTriggerSweep(gtx, test.component))
			runner.Wait(waitDuration)
			_, loop := test.tm.getLoopStressCount()
			rows = append(rows, rxLoopStressResult{offset / 1000, loop})
			test.success(fmt.Sprintf("%.2f", loop))
		})
	}
	runner.Run("Bringing Carrier to Zero", true, func() {
		for i := 0; i < noOfStepsInOneDirection; i++ {
			//positive to Zero loop
			runner.Exec(setGtxTriggerSweep(gtx, test.component))
			runner.Wait(waitDuration)
		}
	})
	offset = 0.0
	for i := 0; i < noOfStepsInOneDirection; i++ {
		//negative loop
		offset = offset - sweepStep
		runner.Run(fmt.Sprintf("Reading Loop Stress value for %.2f kHz", offset/1000), false, func() {
			runner.Exec(setGtxTriggerSweep(gtx, test.component))
			runner.Wait(waitDuration)
			_, loop := test.tm.getLoopStressCount()
			rows = append(rows, rxLoopStressResult{offset / 1000, loop})
			test.success(fmt.Sprintf("%.2f", loop))
		})
	}
	runner.Run("Bringing Carrier to Zero", true, func() {
		runner.Exec(setGTxStartSweep(gtx, test.component))
		runner.Wait(waitDuration * noOfStepsInOneDirection)
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
