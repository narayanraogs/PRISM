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
	executeTest.Register("RFUplink", "Full", newRxRFUplink)
	executeTest.Register("RFUplink", "Fast", newRxRFUplink)
	executeTest.Register("RFUplink", "Full-Doppler", newRxRFUplink)
	executeTest.Register("RFUplink", "Fast-Doppler", newRxRFUplink)
	results.Register("RFUplink", results.NewDefaultProcessor([]string{"Results"}))
}

func newRxRFUplink() executeTest.Tester {
	var test rxRFUplink
	return &test
}

type rxRFUplink struct {
	rxBaseTest
	nominalPowerLevel float64
	category          string
}

func (test *rxRFUplink) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rxBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *rxRFUplink) DBValidate() error {
	test.category = test.testCategory
	test.testCategory = ""
	err := test.validateAndPrepare(test.readFrequencyProfile)
	if err != nil {
		return err
	}
	if strings.Contains(test.category, "Doppler") {
		test.report.AddTestInformation("Doppler Enabled", "true")
		test.report.AddTestInformation("Total Entries", fmt.Sprintf("%d", len(test.dopplerFrequencies)))
		test.report.AddTestInformation("First Frequency", fmt.Sprintf("%d", test.dopplerFrequencies[0]))
		test.report.AddTestInformation("Last Frequency", fmt.Sprintf("%d", test.dopplerFrequencies[len(test.dopplerFrequencies)-1]))
	}
	test.testCategory = test.category
	test.rollbackRequired = false
	return nil
}

func (test *rxRFUplink) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "GTX", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *rxRFUplink) measure(runner *StepRunner) error {
	start := time.Now()
	var result rxRFUplinkResult
	result.setPower = test.nominalPowerLevel
	result.modulation = test.rxSpec.ModulationScheme
	result.frequencyDeviationSpec = test.rxSpec.FrequencyDeviationFM.Float64
	result.modIndexSpec = test.rxSpec.TCModIndex.Float64

	sa := test.ctx.Selected.SA

	if test.rollbackToBeRead {
		test.readRollback(runner)
	}
	test.setTSMPathForSA(runner)
	test.removeRFLink(runner)

	test.setupSAForUplink(runner)

	test.computeZerodBmDifference(runner)

	if !strings.Contains(test.category, "Fast") {
		result.measured = test.measureModulation(runner)
	} else {
		result.measured = "NA"
	}

	actualPower := test.setPowerLevel(runner, test.nominalPowerLevel, true)
	result.actualPower = actualPower

	if strings.Contains(test.category, "Doppler") {
		test.enableDopplerCompensation(runner)
	}

	lock, agc := test.uplinkWithModulation(runner, true, true)
	if runner.Err() == nil {
		result.lockStatus = lock
		result.agc = agc
	}
	_ = test.checkForBSLock(runner, true)

	runner.Run("Fetching Spectrum", true, func() {
		resp := runner.Exec(sa.GetSpectrumDump)
		if runner.execErr != nil {
			return
		}
		caption := fmt.Sprintf("Uplink Spectrum for %s", test.configName)
		test.spectra = append(test.spectra, reports.Images{Caption: caption, FileData: resp.Result["SpectrumDump"].String})
	})

	if !runner.Describe {
		var strRows = make([][]string, 0)
		header := result.ToHeader()
		strRows = append(strRows, result.ToRow())
		test.saveResultsToCSV(test.testCategory, header, strRows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}

func (test *rxRFUplink) SetParameters(params map[string]interface{}) error {
	value, ok := params["NominalPower"]
	if !ok {
		return fmt.Errorf("unable to get nominal power parameter")
	}
	test.nominalPowerLevel = value.(float64)
	return nil
}
