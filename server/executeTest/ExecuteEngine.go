package executeTest

import (
	"context"
	"fmt"
	"prismServer/logger"
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
	logger.Log.Info("Initializing Test Engine", "testName", init.TestName, "testCategory", init.TestCategory, "configName", init.ConfigName, "parameters", params)
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
	ctx.Ctx = context.Background()
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
	e.context.Ctx = ctx
	defer close(e.context.UpdateChannel)
	defer e.executeRollbacks()

	logger.Log.Info("Test Phase Started: DBValidate", "testName", e.context.Progress.TestName)
	err := e.test.DBValidate()
	if err != nil {
		e.context.Progress.DBValidationStatus = "Failed"
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		logger.Log.Error("Test Phase Failed: DBValidate", "testName", e.context.Progress.TestName, "error", err.Error())
		return err
	}

	e.context.Progress.DBValidationStatus = "Success"
	e.context.UpdateChannel <- *e.context.Progress
	logger.Log.Info("Test Phase Completed: DBValidate", "testName", e.context.Progress.TestName)

	logger.Log.Info("Test Phase Started: InstrumentConnection", "testName", e.context.Progress.TestName)
	err = e.checkInstrumentConnection(ctx)
	if err != nil {
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		logger.Log.Error("Test Phase Failed: InstrumentConnection", "testName", e.context.Progress.TestName, "error", err.Error())
		return err
	}
	logger.Log.Info("Test Phase Completed: InstrumentConnection", "testName", e.context.Progress.TestName)

	logger.Log.Info("Test Phase Started: PreTestTM", "testName", e.context.Progress.TestName)
	err = e.getPreTestTM(ctx)
	if err != nil {
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		logger.Log.Error("Test Phase Failed: PreTestTM", "testName", e.context.Progress.TestName, "error", err.Error())
		return err
	}
	logger.Log.Info("Test Phase Completed: PreTestTM", "testName", e.context.Progress.TestName)

	logger.Log.Info("Test Phase Started: Measurement", "testName", e.context.Progress.TestName)
	e.context.Progress.CurrentStep = "Measurement"
	if len(e.context.Progress.MeasurementStatus) == 0 {
		e.context.Progress.MeasurementStatus = append(e.context.Progress.MeasurementStatus, "InProgress")
	} else {
		e.context.Progress.MeasurementStatus[0] = "InProgress"
	}
	e.context.UpdateChannel <- *e.context.Progress

	err = e.test.Measure(ctx)
	if err != nil {
		logger.Log.Error("Test Phase Failed: Measurement", "testName", e.context.Progress.TestName, "error", err.Error(), "progress", e.context.Progress)
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress

		logger.Log.Info("Test Phase Started: GenerateFailureReport", "testName", e.context.Progress.TestName)
		results, rErr := e.test.GenerateFailureReport(err)
		if rErr != nil {
			logger.Log.Error("Test Phase Failed: GenerateFailureReport", "testName", e.context.Progress.TestName, "error", rErr.Error())
		} else {
			logger.Log.Info("Test Phase Completed: GenerateFailureReport", "testName", e.context.Progress.TestName)
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

		e.context.AskForConfirmation("Press Continue to Rollback/Go to Next test", 0)
		return err
	}
	logger.Log.Info("Test Phase Completed: Measurement", "testName", e.context.Progress.TestName, "progress", e.context.Progress)

	logger.Log.Info("Test Phase Started: PostTestTM", "testName", e.context.Progress.TestName)
	err = e.getPostTestTM(ctx)
	if err != nil {
		logger.Log.Error("Test Phase Failed: PostTestTM", "testName", e.context.Progress.TestName, "error", err.Error())
		e.context.Progress.ErrorMessage = err.Error()
		e.context.UpdateChannel <- *e.context.Progress
		return err
	}
	logger.Log.Info("Test Phase Completed: PostTestTM", "testName", e.context.Progress.TestName)

	logger.Log.Info("Test Phase Started: GenerateReport", "testName", e.context.Progress.TestName)
	results, err := e.test.GenerateReport()
	if err != nil {
		logger.Log.Error("Test Phase Failed: GenerateReport", "testName", e.context.Progress.TestName, "error", err.Error())
	} else {
		logger.Log.Info("Test Phase Completed: GenerateReport", "testName", e.context.Progress.TestName, "reportsGenerated", len(results))
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

	logger.Log.Info("Test Lifecycle Completed", "testName", e.context.Progress.TestName)
	e.context.Progress.CurrentStep = "Completed"
	e.context.UpdateChannel <- *e.context.Progress

	return nil
}

func (e *Engine) executeRollbacks() {
	logger.Log.Info("Executing Rollbacks", "testName", e.context.Progress.TestName)
	err := e.test.Rollback()
	if err != nil {
		logger.Log.Error("Error executing rollback steps", "testName", e.context.Progress.TestName, "error", err.Error())
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

		logger.Log.Warn("Auto Fetch Failed, requesting manual input", "testName", e.context.Progress.TestName, "parameter", param)
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
