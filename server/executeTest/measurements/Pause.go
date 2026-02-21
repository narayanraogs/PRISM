package measurements

import (
	"context"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/utils"
)

func init() {
	executeTest.Register("Pause", "", newPause)
	results.Register("Pause", results.NewDefaultProcessor(make([]string, 0)))
}

func newPause() executeTest.Tester {
	var test pause
	return &test
}

type pause struct {
	baseTest
}

func (test *pause) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.baseTest.Initialize(init, ctx)
	test.impl = test
	test.getInstruments()
}

func (test *pause) Rollback() error {
	return test.baseTest.rollback()
}

func (test *pause) DBValidate() error {
	descriptions := test.describe(context.Background())
	test.prepareSteps(descriptions)
	return nil
}

func (test *pause) prepareSteps(descriptions []string) {
	test.ctx.Progress.MeasurementSteps = make([]string, 0)
	test.ctx.Progress.MeasurementSteps = append(test.ctx.Progress.MeasurementSteps, descriptions...)
	test.ctx.Progress.MeasurementStatus = utils.GetRepeatedArray("Queued", len(descriptions))
	test.ctx.Progress.MeasurementValues = utils.GetRepeatedArray("", len(descriptions))
}

func (test *pause) getInstruments() {
	test.ctx.Progress.Instruments = []string{}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *pause) measure(runner *StepRunner) error {
	runner.Run("Waitng for user", true, func() {
		test.ctx.AskForConfirmation("Click Continue to proceed", 0)
	})

	return runner.Err()
}
