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
	executeTest.Register("SimCmdAndRanging", "", newTpSimCmdRangingMeasurement) //Based on Config Type, test to be distinguished
	results.Register("SimCmdAndRanging", results.NewDefaultProcessor([]string{"Results"}))
}

func newTpSimCmdRangingMeasurement() executeTest.Tester {
	var test tpSimCmdRangingMeasurement
	return &test.tpBaseTest
}

type tpSimCmdRangingMeasurement struct {
	txBaseTest
	rxBaseTest
	tpBaseTest
}

func (test *tpSimCmdRangingMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.tpBaseTest.Initialize(init, ctx)
	test.tpBaseTest.impl = test
	test.tpBaseTest.getInstruments()
}

func (test *tpSimCmdRangingMeasurement) DBValidate() error {
	readSpecTPFunc := func() error {
		return test.readTpSpecTransponder(test.tpBaseTest.config.TpName.String)
	}
	return test.tpBaseTest.validateAndPrepare(readSpecTPFunc)
}

func (test *tpSimCmdRangingMeasurement) getInstruments() {
	test.tpBaseTest.ctx.Progress.Instruments = []string{"SA", "TSM", "GTx"}
	test.tpBaseTest.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.tpBaseTest.ctx.Progress.Instruments))
}

func (test *tpSimCmdRangingMeasurement) measure(runner *StepRunner) error {

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
		runner.Exec(setTCAndRanging(gtx, test.rxBaseTest.component))
		runner.Exec(setRangingToneFrequency(gtx, test.tpRangingSpec.ToneFrequency))
	})

	var measuredModIndxTC, measuredModIndxTone string

	if strings.ToUpper(test.rxBaseTest.rxSpec.ModulationScheme) == "PM" {
		runner.Run("Measuring Deviation in Mod Index of TC", false, func() {
			runner.Exec(setGTxModIndexTone(gtx, test.rxBaseTest.component, test.tpRangingSpec.UplinkToneMISimultaneousCmdAndRanging.Float64))
			runner.Exec(setGTxModIndexTC(gtx, test.rxBaseTest.component, test.tpRangingSpec.TCMISimultaneousCmdAndRanging))
			runner.Exec(setGTxModulationOn(gtx, test.rxBaseTest.component))
			test.tpBaseTest.ctx.AskForConfirmation("Issue one command to enable modulation", 0)

			resp := runner.Exec(getModIndex(sa, test.rxBaseTest.rxSpec.TCSubCarrierFrequency))
			if runner.Err() == nil {
				v := (resp.Result["modIndexForLeft"].Value + resp.Result["modIndexForRight"].Value) / 2
				measuredModIndxTC = fmt.Sprintf("%.2f rad", v)
				diff := math.Abs(test.tpRangingSpec.TCMISimultaneousCmdAndRanging - v)
				tolerance := 0.3 * test.tpRangingSpec.TCMISimultaneousCmdAndRanging
				if diff <= tolerance {
					test.tpBaseTest.success(fmt.Sprintf("deviation is %.2f", diff))
				} else {
					test.tpBaseTest.failure("Mod Index Deviation beyond Tolerance")
					runner.execErr = fmt.Errorf("mod index deviation beyond tolerance")
				}
			}
		})

		runner.Run("Measuring Deviation in Mod Index of uplink Tone", false, func() {
			resp := runner.Exec(getModIndex(sa, test.tpRangingSpec.ToneFrequency))
			if runner.Err() == nil {
				v := (resp.Result["modIndexForLeft"].Value + resp.Result["modIndexForRight"].Value) / 2
				measuredModIndxTone = fmt.Sprintf("%.2f rad", v)
				diff := math.Abs(test.tpRangingSpec.UplinkToneMISimultaneousCmdAndRanging.Float64 - v)
				tolerance := 0.3 * test.tpRangingSpec.UplinkToneMISimultaneousCmdAndRanging.Float64
				if diff <= tolerance {
					test.tpBaseTest.success(fmt.Sprintf("deviation is %.2f", diff))
				} else {
					test.tpBaseTest.failure("Mod Index Deviation beyond Tolerance")
					runner.execErr = fmt.Errorf("mod index deviation beyond tolerance")
				}
			}
		})
	}

	runner.Run("Capturing Uplink Modulated Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.Err() == nil {
			caption := fmt.Sprintf("Uplink Spectrum")
			test.tpBaseTest.spectra = append(test.tpBaseTest.spectra, reports.Images{
				Caption:  caption,
				FileData: resp.Result["SpectrumDump"].String,
			})
		}
	})

	var rows [][]string
	header := (&tpSimCmdRangingResult{}).ToHeader()

	for i, powerLevel := range test.tpBaseTest.powerLevels {
		runner.Run(fmt.Sprintf("Setting attenuation for : %.2f dBm", powerLevel), false, func() {
		})

		runner.Exec(setGTxModulationOff(gtx, test.rxBaseTest.component))
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
					sleepTime := time.Duration((test.rxBaseTest.rxSpec.SweepRange.Float64*4)/test.rxBaseTest.rxSpec.SweepRate.Float64) * time.Second
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
		test.tpBaseTest.ctx.AskForConfirmation("Issue one command to enable modulation", 0)

		var bsLock bool
		runner.Run("Checking for Bit Sync Lock Status", true, func() {
			bsLock = test.rxBaseTest.tm.checkRxBitSyncLock()
			if !bsLock {
				runner.execErr = fmt.Errorf("Bitsync Not locked")
				test.tpBaseTest.failure("Bitsync Not locked")
			} else {
				test.tpBaseTest.success("Lock")
			}
		})

		if runner.Err() != nil {
			return runner.Err()
		}

		var commandsSent, commandsExecuted int
		var ccBefore int

		runner.Run("Setting up Command Execution", true, func() {
			ccBefore = test.rxBaseTest.tm.getCommandCounter()
			tempStr := test.tpBaseTest.ctx.AskForInput("Enter number of commands to be sent", "", 0)
			fmt.Sscanf(tempStr, "%d", &commandsSent)

			test.tpBaseTest.ctx.AskForConfirmation(fmt.Sprintf("Issue %d commands to spacecraft", commandsSent), 0)
		})

		test.txBaseTest.setTSMPathForSA(runner)
		runner.Run("Setting SA profile for Downlink", true, func() {
			_ = test.tpBaseTest.setupSAForDownlinkForDifferentProfiles(runner, test.tpBaseTest.test.DLProfileName.String)
		})

		repeat := 3
		readingsTone := make([]float64, 0, repeat*2)

		for r := 0; r < repeat; r++ {
			runner.Run(
				fmt.Sprintf("Measuring Mod Index"),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, test.tpRangingSpec.ToneFrequency))
					if runner.execErr != nil {
						return
					}
					v1 := resp.Result["modIndexForRight"].Value

					resp2 := runner.Exec(getModIndex(sa, -1*test.tpRangingSpec.ToneFrequency))
					if runner.execErr != nil {
						return
					}
					v2 := resp2.Result["modIndexForLeft"].Value
					readingsTone = append(readingsTone, v1, v2)
					test.tpBaseTest.success(fmt.Sprintf("%.2f", (v1+v2)/2))
				},
			)
		}

		if isFirst {
			runner.Run("Capturing Downlink Spectrum", true, func() {
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
		deviationStatus := "Success"

		runner.Run("Comparing Mod Index with downlink spec", false, func() {
			deviation := measuredTone - specTone
			if specTone != 0 {
				deviationPctTone = (math.Abs(deviation) / specTone) * 100.0
			}
			
			if math.Abs(deviation) > test.tpRangingSpec.AllowedDownlinkMIDeviation {
				deviationStatus = "Error"
			}

			var msg string
			if deviationStatus == "Success" {
				msg = fmt.Sprintf("%.2f%% [In-Spec]", deviationPctTone)
			} else {
				msg = fmt.Sprintf("%.2f%% [Out-Of-Spec]", deviationPctTone)
			}
			test.tpBaseTest.success(msg)
		})

		runner.Run("Commands Executed", false, func() {
			test.tpBaseTest.ctx.AskForConfirmation(fmt.Sprintf("Press continue after %d commands are issued", commandsSent), 0)

			ccAfter := test.rxBaseTest.tm.getCommandCounter()
			if ccAfter >= ccBefore+commandsSent {
				commandsExecuted = commandsSent
				test.tpBaseTest.success(fmt.Sprintf("%d Commands sent successfully", commandsSent))
			} else {
				noExec := test.tpBaseTest.ctx.AskForInput(fmt.Sprintf("Cannot detect if minimum %d are executed.\n Enter Number of Commands Executed", commandsSent), "", 0)
				fmt.Sscanf(noExec, "%d", &commandsExecuted)
				if commandsExecuted >= commandsSent {
					test.tpBaseTest.success(fmt.Sprintf("%d Commands sent successfully", commandsSent))
				} else {
					test.tpBaseTest.failure(fmt.Sprintf("Total no of Commands Executed: %d", commandsExecuted))
					runner.execErr = fmt.Errorf("commands not executed properly")
				}
			}
			
			resultData := tpSimCmdRangingResult{
				ReceiverIPPower:      actualPower,
				SpecUplinkTCMI:       test.tpRangingSpec.TCMISimultaneousCmdAndRanging,
				MeasuredUplinkTCMI:   measuredModIndxTC,
				SpecUplinkToneMI:     test.tpRangingSpec.UplinkToneMISimultaneousCmdAndRanging.Float64,
				MeasuredUplinkTone:   measuredModIndxTone,
				SpecDownlinkToneMI:   specTone,
				MeasuredDownlinkTone: resultValue{Value: measuredTone, Status: deviationStatus},
				CommandsSent:         commandsSent,
				CommandsExecuted:     commandsExecuted,
			}
			rows = append(rows, resultData.ToRow())
		})

		runner.Exec(setGTxModulationOff(gtx, test.rxBaseTest.component))
	}

	runner.Run("Removing RF Uplink", true, func() {
		test.rxBaseTest.removeRFLink(runner)
	})

	if !runner.Describe && runner.Err() == nil {
		test.tpBaseTest.saveResultsToCSV("", header, rows)
		test.tpBaseTest.addFinalTestInformation(start)
	}

	return runner.Err()
}
