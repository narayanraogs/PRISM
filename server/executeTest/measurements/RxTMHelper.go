package measurements

import (
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/tm"
	"strconv"
	"strings"
	"time"
)

type rxTM struct {
	tmtc database.SpecRxTMTC
	rx   database.SpecRx
	ctx  *executeTest.ExecutionContext
}

func (r *rxTM) initialize(tmtc database.SpecRxTMTC, rx database.SpecRx, ctx *executeTest.ExecutionContext) {
	r.tmtc = tmtc
	r.rx = rx
	r.ctx = ctx
}

func (r *rxTM) getTMValues(mnemonics []string) map[string]tm.Parameter {
	var tbr = make(map[string]tm.Parameter)
	var tmChannel = make(chan tm.Parameter, len(mnemonics))
	tm.Fetch(mnemonics, tmChannel)
	for param := range tmChannel {
		if !param.OK {
			prompt := "Unable to get value for " + param.Param + ". Please provide value"
			value := r.ctx.AskForInput(prompt, "", 0)
			param.StringV = value
			param.FloatV = value
			time.Sleep(500 * time.Millisecond)
		}
		tbr[param.Param] = param
	}
	return tbr
}

func (r *rxTM) checkIfLocked(p tm.Parameter, bs bool) bool {
	var floatValue float64
	var err error
	if bs {
		floatValue, err = strconv.ParseFloat(r.tmtc.BSLockStatusValue, 64)
	} else {
		floatValue, err = strconv.ParseFloat(r.tmtc.LockStatusValue.String, 64)
	}
	strComp := err != nil

	if strComp {
		if bs {
			return strings.EqualFold(r.tmtc.BSLockStatusValue, p.StringV)
		} else {
			return strings.EqualFold(r.tmtc.LockStatusValue.String, p.StringV)
		}
	}
	received, _ := strconv.ParseFloat(p.FloatV, 64)
	return received > floatValue
}

func (r *rxTM) checkRxLock() bool {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, r.tmtc.LockStatusMnemonic.String)
	values := r.getTMValues(mnemonics)
	return r.checkIfLocked(values[r.tmtc.LockStatusMnemonic.String], false)
}

func (r *rxTM) getLockAndAGCValue() (bool, float64) {
	var mnemonics = make([]string, 0)
	var checkLock = false
	if r.tmtc.LockStatusMnemonic.Valid && !strings.EqualFold(r.rx.ModulationScheme, "FM") {
		mnemonics = append(mnemonics, r.tmtc.LockStatusMnemonic.String)
		checkLock = true
	}
	mnemonics = append(mnemonics, r.tmtc.AGCMnemonic)
	values := r.getTMValues(mnemonics)
	var lockParam tm.Parameter
	lock := true
	if checkLock {
		lockParam = values[r.tmtc.LockStatusMnemonic.String]
		lock = r.checkIfLocked(lockParam, false)
	}
	var agcParam = values[r.tmtc.AGCMnemonic]
	agc, _ := strconv.ParseFloat(agcParam.FloatV, 64)
	return lock, agc
}

func (r *rxTM) checkRxBitSyncLock() bool {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, r.tmtc.BSLockStatusMnemonic)
	values := r.getTMValues(mnemonics)
	return r.checkIfLocked(values[r.tmtc.BSLockStatusMnemonic], true)
}

func (r *rxTM) getLoopStressCount() (bool, float64) {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, r.tmtc.LockStatusMnemonic.String)
	mnemonics = append(mnemonics, r.tmtc.LoopStressMnemonic.String)
	values := r.getTMValues(mnemonics)
	var lockParam = values[r.tmtc.LockStatusMnemonic.String]
	var lsParam = values[r.tmtc.LoopStressMnemonic.String]
	lock := r.checkIfLocked(lockParam, false)
	fVal, _ := strconv.ParseFloat(lsParam.FloatV, 64)
	return lock, fVal
}

func (r *rxTM) getCommandCounter() int {
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, r.tmtc.CommandCounterMnemonic)
	values := r.getTMValues(mnemonics)
	fVal, _ := strconv.ParseFloat(values[r.tmtc.CommandCounterMnemonic].FloatV, 64)
	return int(fVal)
}

func (r *rxTM) getContinuousLockMonitor(timeoutInSecs int, output chan bool) {
	defer close(output)
	var mnemonics = make([]string, 0)
	mnemonics = append(mnemonics, r.tmtc.LockStatusMnemonic.String)
	var tmChannel = make(chan tm.Parameter, len(mnemonics))
	tm.Subscribe(mnemonics, tmChannel, true)
	autoFailed := false
	lock := true
outerFor:
	for {
		select {
		case <-time.After(time.Duration(timeoutInSecs) * time.Second):
			break outerFor
		case param := <-tmChannel:
			if !param.OK {
				autoFailed = true
				break outerFor
			}
			lock = lock && r.checkIfLocked(param, false)
		}
	}
	if autoFailed {
		r.ctx.Ui.Prompt = "Cannot monitor Continuous Lock. Enter 'Yes' if Receiver Locked through out"
		r.ctx.Ui.UserInput = true
		r.ctx.UpdateChannel <- r.ctx.Ui
		value := <-r.ctx.InputChannel
		output <- strings.EqualFold(value, "yes")
	}
	output <- lock
}
