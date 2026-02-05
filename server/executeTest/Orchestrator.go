package executeTest

import (
	"context"
	"fmt"
	"prismServer/utils"
	"strings"
	"time"
)

var GlobalRegistry = make(map[string]func() Tester)

func getKeyForTest(name string, category string) string {
	key := name
	if !strings.EqualFold(strings.TrimSpace(category), "") {
		key = fmt.Sprintf("%s;%s", name, category)
	}
	return key
}

func Register(name string, category string, test func() Tester) {
	key := getKeyForTest(name, category)
	GlobalRegistry[key] = test
}

type Orchestrator struct {
	ConfigNames    []string
	TestTypes      []string
	TestCategories []string
	Remarks        []string
	Parameters     map[string]interface{}
	Progress       TestProgressResponse
	CommChannel    chan TestProgressResponse
	InputChannel   chan string
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewOrchestrator(configs, testTypes, testCategories, remarks []string, extraParams map[string]interface{}, comm chan TestProgressResponse, input chan string) *Orchestrator {
	var o Orchestrator
	o.ConfigNames = configs
	o.TestTypes = testTypes
	o.TestCategories = testCategories
	o.Remarks = remarks
	o.Parameters = extraParams
	o.CommChannel = comm
	o.InputChannel = input
	o.ctx, o.cancel = context.WithCancel(context.Background())

	o.Progress.TestStatus = make([]TestStatus, 0)
	for i := 0; i < len(configs); i++ {
		var t TestStatus
		t.Config = configs[i]
		t.TestType = testTypes[i]
		t.TestCategory = testCategories[i]
		t.TestStatus = "Queued"
		o.Progress.TestStatus = append(o.Progress.TestStatus, t)
	}
	o.Progress.Summary = make([]TestResult, 0)
	o.CommChannel <- o.Progress

	return &o
}

func (o *Orchestrator) Abort() {
	o.cancel()
}

func (o *Orchestrator) RunTests() {
	defer close(o.CommChannel)

	var details map[string]utils.CommandResponse

	for i := range o.TestTypes {
		o.Progress.TestStatus[i].TestStatus = "InProgress"
		o.CommChannel <- o.Progress

		updateChannel := make(chan interface{}, 10)
		readRollback := false
		configChanged := false
		rollbackRequired := false
		if i == 0 {
			readRollback = true
			configChanged = true
		}
		if i > 0 {
			if !strings.EqualFold(o.ConfigNames[i], o.ConfigNames[i-1]) {
				configChanged = true
			}
		}
		if i == len(o.TestTypes)-1 {
			rollbackRequired = true
		}
		var init Initializer
		init.ConfigName = o.ConfigNames[i]
		init.TestName = o.TestTypes[i]
		init.TestCategory = o.TestCategories[i]
		init.Remark = o.Remarks[i]
		init.ReadRollback = readRollback
		init.ConfigChanged = configChanged
		init.RollbackRequired = rollbackRequired

		engine := NewTestExecutor(init, o.Parameters, o.InputChannel, updateChannel)
		if engine == nil {
			o.Progress.TestStatus[i].TestStatus = "Failure"
			o.CommChannel <- o.Progress
			continue
		}
		if i != 0 {
			engine.setRollbackDetails(details)
		}

		go func() {
			for in := range updateChannel {
				switch v := in.(type) {
				case SingleTestProgress:
					o.Progress.Progress = v
				case UserInteraction:
					o.Progress.UI = v
				case TestResult:
					o.Progress.Summary = append(o.Progress.Summary, v)
				}
				o.CommChannel <- o.Progress
			}
		}()

		err := engine.Execute(o.ctx)
		time.Sleep(1 * time.Second)
		if err != nil {
			o.Progress.TestStatus[i].TestStatus = "Failure"
		} else {
			o.Progress.TestStatus[i].TestStatus = "Success"
			if i == 0 {
				details = engine.getRollbackDetails()
			}
		}
		o.CommChannel <- o.Progress

		if o.ctx.Err() != nil {
			for j := i + 1; j < len(o.TestTypes); j++ {
				o.Progress.TestStatus[j].TestStatus = "Aborted"
			}
			o.CommChannel <- o.Progress
			break
		}

		time.Sleep(1 * time.Second)
	}
}
