package measurements

import (
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"strings"
	"time"
)

func init() {
	executeTest.Register("ModIndex", "", newTpModIndexMeasurement) //Based on Config Type, test to be distinguished
	results.Register("ModIndex", results.NewDefaultProcessor([]string{"Results"}))
}

func newTpModIndexMeasurement() executeTest.Tester {
	var test tpModIndexMeasurement
	return &test.tpBaseTest
}

type tpModIndexMeasurement struct {
	txBaseTest
	rxBaseTest
	tpBaseTest
	subCarriers []string
}

func (test *tpModIndexMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.tpBaseTest.Initialize(init, ctx)
	test.tpBaseTest.impl = test
	test.tpBaseTest.getInstruments()
}

func (test *tpModIndexMeasurement) DBValidate() error {
	readSpecTPFunc := func() error {
		return test.readTpSpecTransponder(test.tpBaseTest.config.TpName.String)
	}
	return test.tpBaseTest.validateAndPrepare(readSpecTPFunc)
}

func (test *tpModIndexMeasurement) getInstruments() {
	test.tpBaseTest.ctx.Progress.Instruments = []string{"SA", "TSM", "GTx"}
	test.tpBaseTest.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.tpBaseTest.ctx.Progress.Instruments))
}

func (test *tpModIndexMeasurement) measure(runner *StepRunner) error {
	var downlinkProfileNames map[string]string
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

	// Measure Mod Index for downlink
	carrierNames := strings.Split(test.tpBaseTest.testCategory, "+")

	for _, name := range carrierNames {
		testDetail, _ := database.GetTestDetails(test.tpBaseTest.configName, "ModIndex", name)
		downlinkProfileNames[name] = testDetail.DLProfileName.String
		txName := test.tpBaseTest.tpSpec.TxName
		subCar := test.tpBaseTest.txSpecSubCarrier[txName].SubCarrierName
		test.subCarriers = append(test.subCarriers, subCar)

		runner.Exec(setGTxModulationOn(gtx, test.rxBaseTest.component))
		test.txBaseTest.setTSMPathForSA(runner)
		_ = test.tpBaseTest.setupSAForDownlinkForDifferentProfiles(runner)
		repeat := 3

		readings := make([]float64, 0, repeat*2)
		for r := 0; r < repeat; r++ {
			runner.Run(
				fmt.Sprintf("Measuring ModIndex of + %s - Reading %d", subCar, r+1),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, test.tpRangingSpec.ToneFrequency))
					if runner.execErr != nil {
						return
					}
					v := resp.Result["modIndexForRight"].Value
					readings = append(readings, v)
					test.tpBaseTest.success(fmt.Sprintf("%.2f", v))
				},
			)

			runner.Run(
				fmt.Sprintf("Measuring ModIndex of -%s - Reading %d", subCar, r+1),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, -1*test.tpRangingSpec.ToneFrequency))
					if runner.execErr != nil {
						return
					}
					v := resp.Result["modIndexForLeft"].Value
					readings = append(readings, v)
					test.tpBaseTest.success(fmt.Sprintf("%.2f", v))
				},
			)
		}
		runner.Run("Capturing Spectrum", true, func() {
			resp := runner.Exec(sa.GetSpectrumDump)
			if runner.execErr != nil {
				return
			}
			caption := fmt.Sprintf("ModIndex Spectrum for %s", subCar)
			test.tpBaseTest.spectra = append(test.tpBaseTest.spectra, reports.Images{
				Caption:  caption,
				FileData: resp.Result["SpectrumDump"].String,
			})
		})

		test.saToNormalMode(runner)

		var sum float64
		for _, v := range readings {
			sum = sum + v
		}
		measured := 0.0
		if len(readings) > 0 {
			measured = sum / float64(len(readings))
		}

		spec := test.tpRangingSpec.UplinkToneMIOnlyRanging
		deviationPct := 0.0

		if !runner.Describe && runner.Err() == nil {

			deviation := measured - spec
			deviationPct = math.Abs(measured-spec) / spec * 100.0
			deviationStatus := "Success"
			if math.Abs(deviation) > test.tpBaseTest.txSpecSubCarrier[subCar].AllowedModIndexDeviation.Float64 {
				deviationStatus = "Error"
			}

			resultData := tpModIndexResult{
				SubCarrier:        subCar,
				ToneFrequency:     test.tpRangingSpec.ToneFrequency,
				SpecifiedModIndex: spec,
				MeasuredModIndex:  resultValue{Value: measured, Status: deviationStatus},
				Deviation:         resultValue{Value: deviationPct, Status: deviationStatus},
			}
			rows := [][]string{resultData.ToRow()}
			header := resultData.ToHeader()
			test.tpBaseTest.saveResultsToCSV("", header, rows)

			test.tpBaseTest.addFinalTestInformation(start)
		}

	}
	return runner.Err()
}
