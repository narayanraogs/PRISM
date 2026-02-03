package measurements

import (
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/utils"
	"time"
)

func init() {
	executeTest.Register("RFUplinkRemoval", "", newRxRFUplinkRemoval)
	results.Register("RFUplinkRemoval", results.NewDefaultProcessor([]string{"Results"}))
}

func newRxRFUplinkRemoval() executeTest.Tester {
	var test rxRFUplinkRemoval
	return &test
}

type rxRFUplinkRemoval struct {
	rxBaseTest
}

func (test *rxRFUplinkRemoval) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rxBaseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *rxRFUplinkRemoval) DBValidate() error {
	test.testName = "RFUplink"
	err := test.validateAndPrepare(nil)
	if err != nil {
		return err
	}
	test.rollbackRequired = false
	test.testName = "RFUplinkRemoval"
	return nil
}

func (test *rxRFUplinkRemoval) getInstruments() {
	test.ctx.Progress.Instruments = []string{"SA", "GTX", "TSM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *rxRFUplinkRemoval) measure(runner *StepRunner) error {
	start := time.Now()
	var result rxRFUplinkResult
	result.modulation = test.rxSpec.ModulationScheme

	test.setTSMPathForSA(runner)
	test.removeRFLink(runner)

	if !runner.Describe {
		var strRows = make([][]string, 0)
		header := result.ToHeader()
		strRows = append(strRows, result.ToRow())
		test.saveResultsToCSV(test.testCategory, header, strRows)

		test.addFinalTestInformation(start)
	}

	return runner.Err()
}
