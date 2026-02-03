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
	executeTest.Register("Harmonics", "", newTxHarmonicsMeasurement)
	results.Register("Harmonics", results.NewDefaultProcessor([]string{"Harmonic", "Sub-Harmonic"}))
}

func newTxHarmonicsMeasurement() executeTest.Tester {
	var test txHarmonicsMeasurement
	return &test
}

type txHarmonicsMeasurement struct {
	txBaseTest
}

func (test *txHarmonicsMeasurement) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.txBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *txHarmonicsMeasurement) DBValidate() error {
	readHarmonicsFunc := func() error {
		return test.readTxHarmonicsDetails(test.config.TxName)
	}
	return test.validateAndPrepare(readHarmonicsFunc)
}

func (test *txHarmonicsMeasurement) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *txHarmonicsMeasurement) measure(runner *StepRunner) error {
	start := time.Now()

	var harmonicRows = make([][]string, 0)
	var subHarmonicRows = make([][]string, 0)
	var header []string

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}

	test.setTSMPathForSA(runner)
	carrierPower := test.setupSAForDownlink(runner, true)

	for _, harm := range test.txSpecHarmonic {
		var resultData txHarmonicsResult
		var minPeak float64

		runner.Run(fmt.Sprintf("Measure Harmonic at %.2f MHz", harm.Frequency/1e6), false, func() {
			sa := test.ctx.Selected.SA
			runner.Exec(sa.SetNormalMode)
			runner.Exec(setSpectrum(sa, harm.Frequency, test.downlinkProfile.Span,
				float64(test.downlinkProfile.RBW), float64(test.downlinkProfile.VBW)))

			isCarrPresentResp := runner.Exec(sa.CheckIfCarrierIsPresent)
			if runner.execErr != nil {
				return
			}
			minPeak = isCarrPresentResp.Result["MinValue"].Value

			if runner.Describe || !isCarrPresentResp.Result["Carrier"].Bool {
				test.success("Nil")
				if !runner.Describe {
					test.report.AddTestInformation(fmt.Sprintf("Power Level at %.2f MHz", harm.Frequency/1e6), "NIL")
				}

				resultData = txHarmonicsResult{
					IsNilResult:        true,
					SpecifiedFreqMHz:   harm.Frequency / 1e6,
					SpecificationDBc:   test.txSpec.Harmonics,
					NoiseFloorLevelDBm: minPeak,
				}
			} else {
				runner.Exec(waitForSweeps(sa, 1))
				runner.Exec(peakSearch(sa, true, 1))
				runner.Exec(waitForSweeps(sa, 1))
				peakResp := runner.Exec(getMarkerValue(sa, 1))
				if runner.execErr != nil {
					return
				}

				harmFreq := peakResp.Result["MarkerX"].Value
				diff := math.Abs(harmFreq - harm.Frequency)
				twoPPM := harm.Frequency * 2 * 10e-6

				if diff <= twoPPM {
					harmPower := harm.TotalLossFromTxToSA + peakResp.Result["MarkerY"].Value
					harmLevelFromCarrier := harmPower - carrierPower
					test.report.AddTestInformation(fmt.Sprintf("Power Level at %.2f MHz", harm.Frequency/1e6), fmt.Sprintf("%.2f dB", harmPower))

					harmStatus := "Success"
					if harmLevelFromCarrier > test.txSpec.Harmonics {
						harmStatus = "Error"
					}
					resultData = txHarmonicsResult{
						IsNilResult:        false,
						SpecifiedFreqMHz:   harm.Frequency / 1e6,
						MeasuredFreqMHz:    harmFreq / 1e6,
						SpecificationDBc:   test.txSpec.Harmonics,
						LevelDBc:           resultValue{Value: harmLevelFromCarrier, Status: harmStatus},
						NoiseFloorLevelDBm: minPeak,
					}
					test.success(fmt.Sprintf("%.2f dBc", harmLevelFromCarrier))
				} else {
					test.report.AddTestInformation(fmt.Sprintf("Power Level at %.2f MHz", harm.Frequency/1e6), "NIL")
					resultData = txHarmonicsResult{
						IsNilResult:        true,
						SpecifiedFreqMHz:   harm.Frequency / 1e6,
						SpecificationDBc:   test.txSpec.Harmonics,
						NoiseFloorLevelDBm: minPeak,
					}
					test.success("Nil")
				}
			}
		})

		runner.Run("Reading Noise Floor value", false, func() {
			test.success(fmt.Sprintf("%.2f dBc", minPeak))
		})

		runner.Run("Fetching Spectrum", true, func() {
			sa := test.ctx.Selected.SA
			runner.Exec(waitForSweeps(sa, 5))
			spectrumResp := runner.Exec(sa.GetSpectrumDump)
			if runner.execErr != nil {
				return
			}
			caption := fmt.Sprintf("Harmonics Spectrum at %.6f MHz", harm.Frequency/1e6)
			test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: spectrumResp.Result["SpectrumDump"].String})
		})

		if header == nil {
			header = resultData.ToHeader()
		}

		if strings.EqualFold(harm.HarmonicType, "Harmonics") {
			harmonicRows = append(harmonicRows, resultData.ToRow())
		} else {
			subHarmonicRows = append(subHarmonicRows, resultData.ToRow())
		}
	}

	test.saToNormalMode(runner)

	if !runner.Describe && runner.Err() == nil {
		time.Sleep(500 * time.Millisecond)

		test.saveResultsToCSV("Harmonics", header, harmonicRows)
		test.saveResultsToCSV("SubHarmonics", header, subHarmonicRows)

		test.addFinalTestInformation(start)
	}
	return runner.Err()
}
