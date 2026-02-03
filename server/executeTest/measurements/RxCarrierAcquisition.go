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
	executeTest.Register("CarrierAcquisition", "Normal", newRxCarrierAcquisition)
	executeTest.Register("CarrierAcquisition", "Extreme", newRxCarrierAcquisition)
	results.Register("CarrierAcquisition", results.NewDefaultProcessor([]string{"Results"}))
}

func newRxCarrierAcquisition() executeTest.Tester {
	var test rxCarrierAcquisition
	return &test
}

type rxCarrierAcquisition struct {
	rxBaseTest
}

func (test *rxCarrierAcquisition) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rxBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *rxCarrierAcquisition) DBValidate() error {
	return test.validateAndPrepare(test.readFrequencyProfile)
}

func (test *rxCarrierAcquisition) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "GTX", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *rxCarrierAcquisition) measure(runner *StepRunner) error {
	gtx := test.ctx.Selected.GTx
	sa := test.ctx.Selected.SA
	start := time.Now()
	var resultRows = make([]rxCarrierAcquisitionResult, 0)
	var result rxCarrierAcquisitionResult

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}
	test.setTSMPathForSA(runner)
	test.removeRFLink(runner)
	test.setupSAForUplink(runner)
	test.computeZerodBmDifference(runner)

	modMeasured := test.measureModulation(runner)
	test.report.AddTestInformation("Modulation Measured", modMeasured)

	runner.Run("Getting Unlock AGC Value", false, func() {
		_, agc := test.tm.getLockAndAGCValue()
		test.success(fmt.Sprintf("%.2f", agc))
		result.setFrequency = float64(test.rxSpec.Frequency)
		result.offsetFrequency = 0
		result.agcValue = agc
		result.lockStatus = "Unlock"
		resultRows = append(resultRows, result)

	})

	actualPower := test.setPowerLevel(runner, test.powerLevels[0], true)
	test.report.AddTestInformation("Power at Spacecraft Rx", fmt.Sprintf("%.2f dBm", actualPower))

	lock, agc := test.uplinkWithModulation(runner, true, true)
	_ = test.checkForBSLock(runner, true)
	if runner.Err() == nil {
		result = rxCarrierAcquisitionResult{}
		result.setFrequency = float64(test.rxSpec.Frequency)
		result.offsetFrequency = 0
		result.agcValue = agc
		result.lockStatus = "Lock"
		resultRows = append(resultRows, result)
	}

	stepSize := test.frequencyProfile.StepSize.Float64
	if strings.EqualFold(test.testCategory, "Extreme") {
		stepSize = test.frequencyProfile.MaxFrequency.Float64
	}
	noOfStepsF := test.frequencyProfile.MaxFrequency.Float64 * 2 / stepSize
	noOfSteps := int(noOfStepsF) + 1
	stepFrequency := -1 * test.frequencyProfile.MaxFrequency.Float64
	for i := 0; i < noOfSteps; i++ {
		result = rxCarrierAcquisitionResult{}
		result.setFrequency = float64(test.rxSpec.Frequency) + stepFrequency
		result.offsetFrequency = stepFrequency

		runner.Run(fmt.Sprintf("Setting frequency to %.2f kHz and Checking Lock and AGC", stepFrequency/1000), false, func() {
			if !strings.EqualFold(test.rxSpec.ModulationScheme, "CDMA") {
				runner.Exec(setGTxCarrierOff(gtx, test.component))
				runner.Exec(setGTxModulationOff(gtx, test.component))
			} else {
				runner.Exec(setGTxIdleOff(gtx))
			}
			test.setIntermediateFrequency(runner, stepFrequency)
			runner.Exec(setGTxCarrierOn(gtx, test.component))
			lock, agc = test.tm.getLockAndAGCValue()
			if !lock {
				test.failure("Receiver did not Lock")
				err := fmt.Errorf("receiver did not lock")
				runner.execErr = err
				runner.chainErr = err
			}
			result.agcValue = agc
			result.lockStatus = "Lock"
			test.success(fmt.Sprintf("%.2f", agc))
		})

		if i == 0 || i == noOfSteps-1 {
			runner.Run("Fetching Spectrum", true, func() {
				resp := runner.Exec(sa.GetSpectrumDump)
				if runner.execErr != nil {
					return
				}
				caption := fmt.Sprintf("Uplink Spectrum for %s, %.2f kHz", test.configName, stepFrequency/1000)
				test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
			})
		}

		_ = test.checkForBSLock(runner, true)

		if strings.EqualFold(test.frequencyProfile.CommandingRequired, "yes") {
			noOfCommands, ok := test.checkCommandExecution(runner, test.noOfCommandsNominal, true)
			if ok {
				result.noOfCommandsSent = test.noOfCommandsNominal
				result.noOfCommandsExecuted = noOfCommands
			}
			runner.Exec(sa.SetNormalMode)
		}

		resultRows = append(resultRows, result)
		stepFrequency = stepFrequency + stepSize
	}

	runner.Run("Resetting Intermediate Frequency", true, func() {
		runner.Exec(setGTxCarrierOff(gtx, test.component))
		runner.Exec(setGTxModulationOff(gtx, test.component))
		test.setIntermediateFrequency(runner, 0)
	})

	if !runner.Describe {
		var strRows = make([][]string, 0)
		header := result.ToHeader()
		for _, res := range resultRows {
			strRows = append(strRows, res.ToRow())
		}
		test.saveResultsToCSV(test.testCategory, header, strRows)

		test.addFinalTestInformation(start)
	}
	return runner.Err()
}
