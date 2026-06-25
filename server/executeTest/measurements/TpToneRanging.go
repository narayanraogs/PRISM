package measurements

import (
	"fmt"
	"math"

	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"strings"
	"time"
)

func init() {
	executeTest.Register("ToneRanging", "", newTpToneRangingMeasurement) //Based on Config Type, test to be distinguished
	results.Register("ToneRanging", results.NewDefaultProcessor([]string{"Results"}))
}

func newTpToneRangingMeasurement() executeTest.Tester {
	var test tpToneRangingMeasurement
	return &test.tpBaseTest
}

type tpToneRangingMeasurement struct {
	txBaseTest
	rxBaseTest
	tpBaseTest
}

func (test *tpToneRangingMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.tpBaseTest.Initialize(init, ctx)
	test.tpBaseTest.impl = test
	test.tpBaseTest.getInstruments()
}

func (test *tpToneRangingMeasurement) DBValidate() error {
	readSpecTPFunc := func() error {
		return test.readTpSpecTransponder(test.tpBaseTest.config.TpName.String)
	}
	return test.tpBaseTest.validateAndPrepare(readSpecTPFunc)
}

func (test *tpToneRangingMeasurement) getInstruments() {
	test.tpBaseTest.ctx.Progress.Instruments = []string{"SA", "TSM", "GTx"}
	test.tpBaseTest.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.tpBaseTest.ctx.Progress.Instruments))
}

func (test *tpToneRangingMeasurement) measure(runner *StepRunner) error {

	start := time.Now()
	gtx := test.tpBaseTest.ctx.Selected.GTx
	sa := test.tpBaseTest.ctx.Selected.SA

	if test.tpBaseTest.rollbackToBeRead {
		test.tpBaseTest.readRollback(runner)
	}

	test.rxBaseTest.setTSMPathForSA(runner)
	test.rxBaseTest.removeRFLink(runner)

	test.rxBaseTest.setupSAForUplink(runner)

	test.rxBaseTest.computeZerodBmDifference(runner)

	//Get TTCP tone settings from database and set in ttcp
	//Configure IFM for Only tone

	runner.Run("Setting Cortex Configurations for Ranging", true, func() {
		runner.Exec(setOnlyRanging(gtx, test.rxBaseTest.component))
		runner.Exec(setRangingToneFrequency(gtx, test.tpRangingSpec.ToneFrequency))
	})

	modMeasured := test.tpBaseTest.measureModulationTone(runner)
	test.tpBaseTest.report.AddTestInformation("Modulation Measured", modMeasured)

	//Uplink tone for Nominal Level
	var result rxLockDynamicResult // todo : to be corrected
	result.receiverPower = test.tpBaseTest.powerLevels[0]
	actualPower := test.tpBaseTest.setPowerLevel(runner, test.tpBaseTest.powerLevels[0], true)
	result.actualPower = actualPower
	lock, agc := test.tpBaseTest.uplinkWithoutModulation(runner, true)
	if runner.Err() == nil {
		if !lock {
			result.lockStatus = "UNLOCK"
		} else {
			result.lockStatus = "LOCK"
		}
		result.agcValue = agc
	}

	var rows [][]string
	header := (&tpToneRangingResult{}).ToHeader()

	for i, powerLevel := range test.tpBaseTest.powerLevels {
		runner.Run(fmt.Sprintf("Setting attenuation for : %.2f dBm", powerLevel), false, func() {
			// This string is just for the report, setPowerLevel runs the actual steps.
		})
		actualPower := test.tpBaseTest.setPowerLevel(runner, powerLevel, true)
		
		isFirst := i == 0

		if isFirst {
			runner.Run("Routing to Spacecraft", true, func() {
				runner.Exec(setTSMPath(test.tpBaseTest.ctx.Selected.TSM, test.tpBaseTest.tsm.UplinkToSC.String))
			})
			if strings.ToUpper(test.rxBaseTest.rxSpec.ModulationScheme) == "PM" {
				runner.Run("Sweeping", true, func() {
					runner.Exec(setGTxStopSweep(gtx, test.rxBaseTest.component))
					runner.Exec(setGTxStartSweep(gtx, test.rxBaseTest.component))
					sleepTime := time.Duration((test.rxBaseTest.rxSpec.SweepRange.Float64 * 4) / test.rxBaseTest.rxSpec.SweepRate.Float64) * time.Second
					time.Sleep(sleepTime)
					runner.Exec(setGTxStopSweep(gtx, test.rxBaseTest.component))
				})
			}

			runner.Run("Checking for Lock and AGC", true, func() {
				lockSts, agcValue := test.tpBaseTest.tm.getLockAndAGCValue()
				if !lockSts {
					runner.execErr = fmt.Errorf("Receiver did not Lock")
				} else {
					test.tpBaseTest.success(fmt.Sprintf("AGC Value is %.2f", agcValue))
				}
			})
		}

		if runner.Err() != nil {
			return runner.Err()
		}

		runner.Exec(setGTxModulationOn(gtx, test.rxBaseTest.component))
		test.txBaseTest.setTSMPathForSA(runner)
		_ = test.tpBaseTest.setupSAForDownlinkForDifferentProfiles(runner, test.tpBaseTest.test.DLProfileName.String)

		repeat := 3
		readingsTone := make([]float64, 0, repeat*2)

		for r := 0; r < repeat; r++ {
			runner.Run(
				fmt.Sprintf("Measuring ModIndex of + %s - Reading %d", test.tpRangingSpec.RangingName, r+1),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, test.tpRangingSpec.ToneFrequency))
					if runner.execErr != nil {
						return
					}
					v := resp.Result["modIndexForRight"].Value
					readingsTone = append(readingsTone, v)
					test.tpBaseTest.success(fmt.Sprintf("%.2f", v))
				},
			)

			runner.Run(
				fmt.Sprintf("Measuring ModIndex of -%s - Reading %d", test.tpRangingSpec.RangingName, r+1),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, -1*test.tpRangingSpec.ToneFrequency))
					if runner.execErr != nil {
						return
					}
					v := resp.Result["modIndexForLeft"].Value
					readingsTone = append(readingsTone, v)
					test.tpBaseTest.success(fmt.Sprintf("%.2f", v))
				},
			)
		}

		if isFirst {
			runner.Run("Capturing Spectrum", true, func() {
				resp := runner.Exec(sa.GetSpectrumDump)
				if runner.execErr != nil {
					return
				}
				caption := fmt.Sprintf("ModIndex Spectrum for %s", test.tpRangingSpec.RangingName)
				test.tpBaseTest.spectra = append(test.tpBaseTest.spectra, reports.Images{
					Caption:  caption,
					FileData: resp.Result["SpectrumDump"].String,
				})
			})
		}

		test.saToNormalMode(runner)

		var sumTone float64
		for _, v := range readingsTone {
			sumTone = sumTone + v
		}
		measuredTone := 0.0
		if len(readingsTone) > 0 {
			measuredTone = sumTone / float64(len(readingsTone))
		}

		specTone := test.tpRangingSpec.DownlinkMI
		deviationPctTone := 0.0

		if !runner.Describe && runner.Err() == nil {
			deviation := measuredTone - specTone
			if specTone != 0 {
				deviationPctTone = (math.Abs(deviation) / specTone) * 100.0
			}
			deviationStatus := "Success"
			if math.Abs(deviation) > test.tpRangingSpec.AllowedDownlinkMIDeviation {
				deviationStatus = "Error"
			}

			resultData := tpToneRangingResult{
				ReceiverIPPower:      actualPower,
				SpecUplinkToneMI:     test.tpRangingSpec.UplinkToneMIOnlyRanging,
				MeasuredUplinkTone:   modMeasured,
				SpecDownlinkToneMI:   specTone,
				MeasuredDownlinkTone: resultValue{Value: measuredTone, Status: deviationStatus},
				Deviation:            resultValue{Value: deviationPctTone, Status: deviationStatus},
			}
			rows = append(rows, resultData.ToRow())
		}
	}

	if !runner.Describe && runner.Err() == nil {
		test.tpBaseTest.saveResultsToCSV("", header, rows)
		test.tpBaseTest.addFinalTestInformation(start)
	}

	return runner.Err()
}
