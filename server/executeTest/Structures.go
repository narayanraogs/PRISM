package executeTest

import (
	"context"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/reports"
	"prismServer/utils"
)

type devices struct {
	SA  driver.SA
	TSM driver.TSM
	GTx driver.GTX
	PM  driver.PM
	SG  driver.SG
	VSA driver.VSA
	PPM driver.PPM
}

func (dev *devices) Load(profile string) bool {
	pm, pmOK := database.GetPMFromDeviceProfile(profile)
	sa, saOK := database.GetSAFromDeviceProfile(profile)
	gtx, gtxOK := database.GetGTxFromDeviceProfile(profile)
	tsm, tsmOK := database.GetTSMFromDeviceProfile(profile)
	sg, sgOK := database.GetSGFromDeviceProfile(profile)
	vsa, vsaOK := database.GetVSAFromDeviceProfile(profile)
	ppm, ppmOK := database.GetPPMFromDeviceProfile(profile)
	if !pmOK || !saOK || !gtxOK || !tsmOK || !sgOK || !vsaOK || !ppmOK {
		return false
	}
	pmOK = dev.PM.LoadDevice(pm)
	saOK = dev.SA.LoadDevice(sa)
	gtxOK = dev.GTx.LoadDevice(gtx)
	tsmOK = dev.TSM.LoadDevice(tsm)
	sgOK = dev.SG.LoadDevice(sg)
	vsaOK = dev.VSA.LoadDevice(vsa)
	ppmOK = dev.PPM.LoadDevice(ppm)
	if !pmOK || !saOK || !gtxOK || !tsmOK || !sgOK || !vsaOK || !ppmOK {
		return false
	}
	return true
}

type ExecutionContext struct {
	Selected      devices
	InputChannel  chan string
	Progress      *SingleTestProgress
	Ui            *UserInteraction
	UpdateChannel chan interface{}
	TestIndex     int
	Ctx           context.Context
}

type Initializer struct {
	ConfigName       string
	TestName         string
	TestCategory     string
	Remark           string
	ReadRollback     bool
	ConfigChanged    bool
	RollbackRequired bool
}

type Tester interface {
	Initialize(init Initializer, ctx *ExecutionContext)
	Rollback() error
	DBValidate() error
	GetRollbackDetails() map[string]utils.CommandResponse
	SetRollbackMap(r map[string]utils.CommandResponse)
	Measure(ctx context.Context) error
	GenerateReport() (map[string]reports.Result, error)
	SetParameters(map[string]interface{}) error
}

type TestStatus struct {
	TestType     string
	TestCategory string
	Config       string
	TestStatus   string
}

type TestResult struct {
	TestName      string
	TestCategory  string
	Configuration string
	Name          []string
	Result        []reports.Result
}

type SingleTestProgress struct {
	TestName            string
	TestCategory        string
	Configuration       string
	CurrentStep         string
	ErrorMessage        string
	DBValidationStatus  string
	Instruments         []string
	InstrumentStatus    []string
	PreTestTMMnemonics  []string
	PreTestTMValues     []string
	MeasurementSteps    []string
	MeasurementValues   []string
	MeasurementStatus   []string
	PostTestTMMnemonics []string
	PostTestTMValues    []string
}

type UserInteraction struct {
	UserConfirmation bool
	UserInput        bool
	Prompt           string
	TimeoutSecs      int
	DefaultValue     string
}

func (ctx *ExecutionContext) AskForInput(prompt string, defaultValue string, timeout int) string {
	ctx.Ui.Prompt = prompt
	ctx.Ui.DefaultValue = defaultValue
	ctx.Ui.TimeoutSecs = timeout
	ctx.Ui.UserInput = true
	ctx.Ui.UserConfirmation = false
	ctx.UpdateChannel <- *ctx.Ui

	var response string
	select {
	case response = <-ctx.InputChannel:
	case <-ctx.Ctx.Done():
		response = "ABORTED"
	}

	ctx.Ui.Prompt = ""
	ctx.Ui.DefaultValue = ""
	ctx.Ui.TimeoutSecs = 0
	ctx.Ui.UserInput = false
	ctx.UpdateChannel <- *ctx.Ui

	return response
}

func (ctx *ExecutionContext) AskForConfirmation(prompt string, timeout int) bool {
	ctx.Ui.Prompt = prompt
	ctx.Ui.TimeoutSecs = timeout
	ctx.Ui.UserInput = false
	ctx.Ui.UserConfirmation = true
	ctx.UpdateChannel <- *ctx.Ui

	var response string
	select {
	case response = <-ctx.InputChannel:
	case <-ctx.Ctx.Done():
		response = "TIMEOUT"
	}

	ctx.Ui.Prompt = ""
	ctx.Ui.TimeoutSecs = 0
	ctx.Ui.UserConfirmation = false
	ctx.UpdateChannel <- *ctx.Ui

	return response != "TIMEOUT"
}

type TestProgressResponse struct {
	TestStatus []TestStatus
	Progress   SingleTestProgress
	Summary    []TestResult
	UI         UserInteraction
	OK         bool
	Message    string
}

func newSingleTestProgress() *SingleTestProgress {
	var single SingleTestProgress
	single.CurrentStep = "DBValidation"
	single.Instruments = make([]string, 0)
	single.InstrumentStatus = make([]string, 0)
	single.PreTestTMMnemonics = make([]string, 0)
	single.PreTestTMValues = make([]string, 0)
	single.MeasurementSteps = make([]string, 0)
	single.MeasurementValues = make([]string, 0)
	single.MeasurementStatus = make([]string, 0)
	single.PostTestTMMnemonics = make([]string, 0)
	single.PostTestTMValues = make([]string, 0)

	return &single
}
