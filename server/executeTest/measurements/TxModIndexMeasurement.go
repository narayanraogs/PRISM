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
	executeTest.Register("ModIndex", "", newTxModIndexMeasurement)
	results.Register("ModIndex", results.NewDefaultProcessor([]string{"Results"}))
}

func newTxModIndexMeasurement() executeTest.Tester {
	var test txModIndexMeasurement
	return &test
}

type txModIndexMeasurement struct {
	txBaseTest
}

func (test *txModIndexMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *txModIndexMeasurement) DBValidate() error {
	readSubCarrierFunc := func() error {
		return test.readTxSubCarrierDetails(test.config.TxName)
	}
	return test.validateAndPrepare(readSubCarrierFunc)
}

func (test *txModIndexMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txModIndexMeasurement) measure(runner *StepRunner) error {
	//var downlinkProfileNames map[string]string
	start := time.Now()
	sa := test.ctx.Selected.SA

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)

	test.setupSAForDownlink(runner, false)

	carrierNames := strings.Split(test.testCategory, "+")
	const repeat = 3

	//var rows = make([][]string, 0)
	//var header []string
	for _, name := range carrierNames {
		subCar := test.txSpecSubCarrier[name]

		readings := make([]float64, 0, repeat*2)
		for r := 0; r < repeat; r++ {
			runner.Run(
				fmt.Sprintf("Measuring ModIndex of + %s - Reading %d", subCar.SubCarrierName, r+1),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, subCar.Frequency))
					if runner.execErr != nil {
						return
					}
					v := resp.Result["modIndexForRight"].Value
					readings = append(readings, v)
					test.success(fmt.Sprintf("%.2f", v))
				},
			)

			runner.Run(
				fmt.Sprintf("Measuring ModIndex of -%s - Reading %d", subCar.SubCarrierName, r+1),
				false,
				func() {
					resp := runner.Exec(getModIndex(sa, -1*subCar.Frequency))
					if runner.execErr != nil {
						return
					}
					v := resp.Result["modIndexForLeft"].Value
					readings = append(readings, v)
					test.success(fmt.Sprintf("%.2f", v))
				},
			)
		}
		runner.Run("Capturing Spectrum", true, func() {
			resp := runner.Exec(sa.GetSpectrumDump)
			if runner.execErr != nil {
				return
			}
			caption := fmt.Sprintf("ModIndex Spectrum for %s", subCar.SubCarrierName)
			test.spectra = append(test.spectra, reports.Images{
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

		spec := subCar.ModIndex.Float64
		deviationPct := 0.0

		if !runner.Describe && runner.Err() == nil {

			deviation := measured - spec
			deviationPct = math.Abs(measured-spec) / spec * 100.0
			deviationStatus := "Success"
			if math.Abs(deviation) > test.txSpecSubCarrier[subCar.SubCarrierName].AllowedModIndexDeviation.Float64 {
				deviationStatus = "Error"
			}

			resultData := txModIndexResult{
				SubCarrier:          subCar.SubCarrierName,
				SubCarrierFrequency: subCar.Frequency,
				SpecifiedModIndex:   spec,
				MeasuredModIndex:    resultValue{Value: measured, Status: deviationStatus},
				Deviation:           resultValue{Value: deviationPct, Status: deviationStatus},
			}
			rows := [][]string{resultData.ToRow()}
			header := resultData.ToHeader()
			test.saveResultsToCSV("", header, rows)

			test.addFinalTestInformation(start)
		}

	}
	return runner.Err()
}
