package measurements

import (
	"context"
	"fmt"
	"prismServer/utils"
	"time"
)

type StepRunner struct {
	Describe     bool
	Descriptions []string
	Test         *baseTest
	Ctx          context.Context
	execErr      error
	chainErr     error
}

func (r *StepRunner) Exec(fn func() utils.CommandResponse) utils.CommandResponse {
	if r.execErr != nil {
		return utils.CommandResponse{} // Don't execute if a previous Exec in this Run failed
	}

	select {
	case <-r.Ctx.Done():
		r.execErr = fmt.Errorf("user aborted")
		return utils.CommandResponse{}
	default:
		resp := fn()
		if !resp.Success {
			r.execErr = fmt.Errorf("%s", resp.ErrorMessage)
		}
		return resp
	}
}

func (r *StepRunner) Run(desc string, success bool, fn func()) {
	if r.chainErr != nil {
		return
	}

	if r.Describe {
		r.Descriptions = append(r.Descriptions, desc)
		return
	}

	r.execErr = nil

	if r.Test.stepNo < len(r.Test.ctx.Progress.MeasurementStatus) {
		r.Test.ctx.Progress.MeasurementStatus[r.Test.stepNo] = "InProgress"
	}
	r.Test.ctx.UpdateChannel <- *r.Test.ctx.Progress

	fn()

	if r.execErr != nil {
		r.Test.failure(r.execErr.Error())
		r.chainErr = r.execErr
		return
	}
	if success {
		r.Test.success("Success")
	}
}

func (r *StepRunner) Err() error {
	return r.chainErr
}

func (r *StepRunner) SetError(err error) {
	r.execErr = err
	r.Test.failure(r.execErr.Error())
	r.chainErr = r.execErr
}

func (r *StepRunner) Wait(seconds int) utils.CommandResponse {
	for i := 0; i < seconds; i++ {
		select {
		case <-r.Ctx.Done():
			r.execErr = fmt.Errorf("user aborted")
			return utils.CommandResponse{}
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return utils.CommandResponse{
		Success: true,
	}
}
