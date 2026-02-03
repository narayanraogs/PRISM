package executeTest

import (
	"context"
	"fmt"
	"prismServer/reports"
	"prismServer/tm"
	"prismServer/utils"
	"slices"
	"strings"
	"time"
)

type Engine struct {
	test    Tester
	params  map[string]interface{}
	context *ExecutionContext
}

func NewTestExecutor(init Initializer, params map[string]interface{}, input chan string, updateChannel chan interface{}) *Engine {
	var e Engine
	var t func() Tester
	var ok bool

	key := getKeyForTest(init.TestName, init.TestCategory)
	t, ok = GlobalRegistry[key]
	if !ok {
		key = getKeyForTest(init.TestName, "")
		t, ok = GlobalRegistry[key]
		if !ok {
			return nil
		}
	}
	e.test = t()

	var ctx ExecutionContext
	ctx.InputChannel = input
	ctx.UpdateChannel = updateChannel
	ctx.Progress = newSingleTestProgress()
	ctx.Progress.TestCategory = init.TestCategory
	ctx.Progress.TestName = init.TestName
	ctx.Progress.Configuration = init.ConfigName
	ctx.Ui = &UserInteraction{}
	e.context = &ctx
	e.test.Initialize(init, e.context)
	err := e.test.SetParameters(params)
	if err != nil {
		return nil
	}
	e.context.UpdateChannel <- *e.context.Progress
	return &e
}

func (e *Engine) getRollbackDetails() map[string]utils.CommandResponse {
	return e.test.GetRollbackDetails()
}

func (e *Engine) setRollbackDetails(details map[string]utils.CommandResponse) {
	e.test.SetRollbackMap(details)
}

func (e *Engine) Execute(ctx context.Context) error {
	defer close(e.context.UpdateChannel)
	defer e.executeRollbacks()

	fmt.Println("DBValidate")
	err := e.test.DBValidate()
	if err != nil {
		e.context.Progress.DBValidationStatus = "Failed"
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		return err
	}

	e.context.Progress.DBValidationStatus = "Success"
	e.context.UpdateChannel <- *e.context.Progress

	fmt.Println("InstrumentConnection")
	err = e.checkInstrumentConnection(ctx)
	if err != nil {
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		return err
	}

	fmt.Println("Pre-Test")
	err = e.getPreTestTM(ctx)
	if err != nil {
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		return err
	}

	fmt.Println("Measurement")
	e.context.Progress.CurrentStep = "Measurement"
	if len(e.context.Progress.MeasurementStatus) == 0 {
		e.context.Progress.MeasurementStatus = append(e.context.Progress.MeasurementStatus, "InProgress")
	} else {
		e.context.Progress.MeasurementStatus[0] = "InProgress"
	}
	e.context.UpdateChannel <- *e.context.Progress

	err = e.test.Measure(ctx)
	if err != nil {
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		e.context.AskForConfirmation("Press Continue to Rollback/Go to Next test", 0)
		return err
	}

	fmt.Println("Post-Test")
	err = e.getPostTestTM(ctx)
	if err != nil {
		fmt.Println("Exiting in post test")
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		return err
	}

	fmt.Println("Generate Report")
	results, err := e.test.GenerateReport()
	if err != nil {
		fmt.Println("Error during report Generation", err.Error())
	} else {
		fmt.Printf("Generated %d reports for summary\n", len(results))
		var sts TestResult
		sts.Name = make([]string, 0)
		sts.Result = make([]reports.Result, 0)
		sts.TestName = e.context.Progress.TestName
		sts.TestCategory = e.context.Progress.TestCategory
		sts.Configuration = e.context.Progress.Configuration
		for k, v := range results {
			sts.Name = append(sts.Name, k)
			sts.Result = append(sts.Result, v)
		}
		e.context.UpdateChannel <- sts
	}

	fmt.Println("Completed")
	e.context.Progress.CurrentStep = "Completed"
	e.context.UpdateChannel <- *e.context.Progress

	return nil
}

func (e *Engine) executeRollbacks() {
	fmt.Println("Executing rollbacks")
	err := e.test.Rollback()
	if err != nil {
		fmt.Println("Error executing rollback steps:", err)
	}
}

func (e *Engine) getPreTestTM(ctx context.Context) error {
	e.context.Progress.CurrentStep = "PreTestTM"
	e.context.UpdateChannel <- *e.context.Progress
	return e.getTMData(ctx, e.context.Progress.PreTestTMMnemonics, e.context.Progress.PreTestTMValues)
}

func (e *Engine) getPostTestTM(ctx context.Context) error {
	e.context.Progress.CurrentStep = "PostTestTM"
	e.context.UpdateChannel <- *e.context.Progress
	return e.getTMData(ctx, e.context.Progress.PostTestTMMnemonics, e.context.Progress.PostTestTMValues)
}

func (e *Engine) getTMData(ctx context.Context, mnemonics []string, values []string) error {
	if len(mnemonics) == 0 {
		return nil
	}
	var tmChannel = make(chan tm.Parameter, 10)
	go tm.Fetch(mnemonics, tmChannel)

TMSubscriptionLoop:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("user aborted")
		case param, ok := <-tmChannel:
			if !ok {
				break TMSubscriptionLoop
			}
			if param.OK {
				mnemonic := param.Param
				if !strings.EqualFold(strings.TrimSpace(param.Stream), "") {
					mnemonic = param.Stream + ":" + param.Param
				}
				index := slices.Index(mnemonics, mnemonic)
				if index != -1 {
					if param.OK {
						values[index] = param.StringV
						e.context.UpdateChannel <- *e.context.Progress
					}
				}
			}
		}
	}

	for slices.Contains(values, "") {
		index := slices.Index(values, "")
		param := mnemonics[index]

		e.context.Ui.Prompt = "Auto Fetch failed for " + param + ". Please enter value manually"
		e.context.Ui.UserConfirmation = false
		e.context.Ui.UserInput = true
		e.context.UpdateChannel <- *e.context.Ui

		select {
		case <-ctx.Done():
			return fmt.Errorf("user aborted")
		case value := <-e.context.InputChannel:
			values[index] = value
			e.context.Ui.UserInput = false
			e.context.Ui.Prompt = ""
			e.context.UpdateChannel <- *e.context.Ui
			e.context.UpdateChannel <- *e.context.Progress
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}
