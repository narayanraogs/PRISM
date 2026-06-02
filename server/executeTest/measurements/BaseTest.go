package measurements

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/resultsDB"
	"prismServer/utils"
	"strings"
	"time"
)

type specificTester interface {
	measure(runner *StepRunner) error
}

type baseTest struct {
	configName       string
	testName         string
	testCategory     string
	rollbackToBeRead bool
	configChanged    bool
	rollbackRequired bool

	ctx         *executeTest.ExecutionContext
	rollbackMap map[string]utils.CommandResponse
	report      reports.Report
	spectra     []reports.Images
	summary     [][]string
	filenames   []string
	stepNo      int
	reportTime  time.Time

	config    database.Configuration
	tsm       database.TSMConfiguration
	test      database.Test
	startTime time.Time
	impl      specificTester
}

func (test *baseTest) Initialize(init executeTest.Initializer, ctx *executeTest.ExecutionContext) {
	test.rollbackMap = make(map[string]utils.CommandResponse)
	test.ctx = ctx
	test.startTime = time.Now()

	test.configName = init.ConfigName
	test.testName = init.TestName
	test.testCategory = init.TestCategory
	test.ctx.Progress.Configuration = init.ConfigName
	test.ctx.Progress.TestCategory = init.TestCategory
	test.ctx.Progress.TestName = init.TestName

	test.rollbackToBeRead = init.ReadRollback
	test.configChanged = init.ConfigChanged
	test.rollbackRequired = init.RollbackRequired

	test.reportTime = test.report.SetHeader(init.ConfigName, init.TestName, init.TestCategory, utils.GetTestPhase())
	test.report.Remarks = init.Remark
	test.spectra = make([]reports.Images, 0)
	test.filenames = make([]string, 0)
}

func (test *baseTest) readRollback(runner *StepRunner) {
	sa := test.ctx.Selected.SA
	tsm := test.ctx.Selected.TSM

	runner.Run("Reading SA State", true, func() {
		runner.Exec(sa.SetAlignmentOff)
		saResp := runner.Exec(sa.GetSpectrum)
		test.rollbackMap["SA"] = saResp
	})

	runner.Run("Reading TSM State", true, func() {
		time.Sleep(1 * time.Second)
		tsmResp := runner.Exec(tsm.GetDriverPath)
		test.rollbackMap["TSM"] = tsmResp
	})
}

func (test *baseTest) rollback() error {
	if !test.rollbackRequired {
		return nil
	}
	sa := test.ctx.Selected.SA
	tsm := test.ctx.Selected.TSM
	_, ok := test.rollbackMap["SA"]
	if ok {
		center := test.rollbackMap["SA"].Result["CenterFrequency"].Value
		span := test.rollbackMap["SA"].Result["Span"].Value
		rbw := test.rollbackMap["SA"].Result["RBW"].Value
		vbw := test.rollbackMap["SA"].Result["VBW"].Value
		ref := test.rollbackMap["SA"].Result["ReferenceLevel"].Value
		sa.SystemPreset()
		sa.SetSpectrum(center, span, rbw, vbw)
		sa.SetReferenceLevel(ref)
		sa.SetAlignmentOn()
	}
	t, ok := test.rollbackMap["TSM"]
	if ok {
		tsm.SetDriverStatus(t.Result["DriverPath"].String)
	}
	return nil
}

func (test *baseTest) DBValidate() error {
	test.ctx.Progress.CurrentStep = "DBValidation"
	test.ctx.Progress.DBValidationStatus = "InProgress"
	test.ctx.UpdateChannel <- *test.ctx.Progress

	var ok bool
	test.test, ok = database.GetTestDetails(test.configName, test.testName, test.testCategory)
	if !ok {
		return fmt.Errorf("unable to read Test Details for %s, %s %s", test.configName, test.testName, test.testCategory)
	}

	test.config, ok = database.GetConfigurationDetails(test.configName)
	if !ok {
		return fmt.Errorf("unable to read Configuration Details for %s", test.configName)
	}

	test.tsm, ok = database.GetTSMPathDetails(test.config.TSMConfigurationName)
	if !ok {
		return fmt.Errorf("unable to read TSM Configuration Details for %s", test.config.TSMConfigurationName)
	}

	if !strings.EqualFold(strings.TrimSpace(test.test.TMProfileName.String), "") {
		profile, ok := database.GetTMProfile(test.test.TMProfileName.String)
		if !ok {
			return fmt.Errorf("unable to read TM  Details for %s", test.test.TMProfileName.String)
		}
		if !strings.EqualFold(strings.TrimSpace(profile.PreRequisiteTM.String), "") {
			temp := strings.Split(profile.PreRequisiteTM.String, ",")
			for _, t := range temp {
				test.ctx.Progress.PreTestTMMnemonics = append(test.ctx.Progress.PreTestTMMnemonics, strings.TrimSpace(t))
			}
			test.ctx.Progress.PreTestTMValues = utils.GetRepeatedArray("", len(test.ctx.Progress.PreTestTMMnemonics))
		}
		if !strings.EqualFold(strings.TrimSpace(profile.LogTM.String), "") {
			temp := strings.Split(profile.LogTM.String, ",")
			for _, t := range temp {
				test.ctx.Progress.PostTestTMMnemonics = append(test.ctx.Progress.PostTestTMMnemonics, strings.TrimSpace(t))
			}
			test.ctx.Progress.PostTestTMValues = utils.GetRepeatedArray("", len(test.ctx.Progress.PostTestTMMnemonics))
		}
	}

	ok = test.ctx.Selected.Load(test.config.DeviceProfileName)
	if !ok {
		return fmt.Errorf("unable to load devices")
	}
	return nil
}

func (test *baseTest) getInstruments() {
	test.ctx.Progress.Instruments = []string{"PM", "SA", "SG", "GTX", "TSM", "PPM"}
	test.ctx.Progress.InstrumentStatus = utils.GetRepeatedArray("InProgress", len(test.ctx.Progress.Instruments))
}

func (test *baseTest) success(message string) {
	test.ctx.Progress.MeasurementValues[test.stepNo] = message
	test.ctx.Progress.MeasurementStatus[test.stepNo] = "Success"
	test.stepNo = test.stepNo + 1
	if test.stepNo < len(test.ctx.Progress.MeasurementSteps) {
		test.ctx.Progress.MeasurementStatus[test.stepNo] = "InProgress"
	}
	test.ctx.UpdateChannel <- *test.ctx.Progress
}

func (test *baseTest) failure(message string) {
	test.ctx.Progress.MeasurementValues[test.stepNo] = message
	test.ctx.Progress.MeasurementStatus[test.stepNo] = "Failed"
	test.ctx.Progress.ErrorMessage = message
	test.ctx.UpdateChannel <- *test.ctx.Progress
}

func (test *baseTest) GetRollbackDetails() map[string]utils.CommandResponse {
	return test.rollbackMap
}

func (test *baseTest) SetRollbackMap(r map[string]utils.CommandResponse) {
	test.rollbackMap = make(map[string]utils.CommandResponse)
	for k, v := range r {
		test.rollbackMap[k] = v
	}
}

func (test *baseTest) GenerateReport() (map[string]reports.Result, error) {

	generateResults, order, plots, err := results.GenerateResults(test.testName, test.filenames)
	if err != nil {
		return nil, err
	}
	test.report.SetOrder(order)
	for k := range generateResults {
		test.report.SetResults(k, generateResults[k].Header, generateResults[k].Data)
	}
	for _, plot := range plots {
		test.report.Screenshots = append(test.report.Screenshots, plot)
	}

	var relativeFilenames []string
	for _, absPath := range test.filenames {
		relPath := strings.Replace(absPath, utils.Config.BaseFolder, "", 1)
		relativeFilenames = append(relativeFilenames, relPath)
	}
	test.report.Filenames = relativeFilenames

	test.report.SetPreRequisiteTM(test.ctx.Progress.PreTestTMMnemonics, test.ctx.Progress.PreTestTMValues)
	test.report.SetPostTestTM(test.ctx.Progress.PostTestTMMnemonics, test.ctx.Progress.PostTestTMValues)
	totalTime := time.Since(test.startTime)
	totalTime = totalTime.Round(time.Second)
	seconds := int64(totalTime.Seconds())
	minutes := seconds / 60
	seconds = seconds % 60
	test.report.AddTestInformation("Total Time", fmt.Sprintf("%d:%d", minutes, seconds))
	pdfPath, err := reports.GenerateResult(test.report, true, true, true, true, true)
	if err != nil {
		return nil, err
	}

	csvPath := test.getRelativeFilenames()
	resultDir := utils.GetTestResultDirectory()
	resultDir = filepath.Join(resultDir, test.testName)
	_ = os.MkdirAll(resultDir, 0755)
	fileName := test.testName
	if strings.TrimSpace(test.testCategory) != "" {
		fileName = fileName + "-" + test.testCategory
	}
	fileName = fileName + "-" + test.configName
	fileName = utils.GetOldTimeStampedFileName(fileName, test.reportTime) + ".pdf"
	fileName = filepath.Join(resultDir, fileName)

	err = os.Rename(pdfPath, fileName)
	if err != nil {
		return nil, err
	}
	fileName = strings.Replace(fileName, utils.Config.BaseFolder, "", 1)

	err = resultsDB.InsertReport(test.report, fileName, csvPath)
	if err != nil {
		return nil, err
	}

	return generateResults, nil
}

func (test *baseTest) GenerateFailureReport(failErr error) (map[string]reports.Result, error) {
	test.report.OK = false
	test.report.Message = failErr.Error()

	header := []string{"Measurement Step", "Value"}
	var data [][]reports.DataCell

	steps := test.ctx.Progress.MeasurementSteps
	values := test.ctx.Progress.MeasurementValues

	for i := 0; i < len(steps); i++ {
		stepName := steps[i]
		val := ""
		if i < len(values) {
			val = values[i]
		}
		row := []reports.DataCell{
			reports.GetDataCell(stepName),
			reports.GetDataCell(val),
		}
		data = append(data, row)
	}

	errorRow := []reports.DataCell{
		reports.GetDataCell("Failure Error"),
		reports.GetDataCell(failErr.Error()),
	}
	errorRow[0].SetError()
	errorRow[1].SetError()
	data = append(data, errorRow)

	test.report.SetOrder([]string{"Failed Execution Steps"})
	test.report.SetResults("Failed Execution Steps", header, data)

	test.report.SetPreRequisiteTM(test.ctx.Progress.PreTestTMMnemonics, test.ctx.Progress.PreTestTMValues)
	test.report.SetPostTestTM(test.ctx.Progress.PostTestTMMnemonics, test.ctx.Progress.PostTestTMValues)

	totalTime := time.Since(test.startTime).Round(time.Second)
	seconds := int64(totalTime.Seconds())
	minutes := seconds / 60
	seconds = seconds % 60
	test.report.AddTestInformation("Total Time", fmt.Sprintf("%d:%d", minutes, seconds))

	pdfPath, err := reports.GenerateResult(test.report, true, true, true, true, true)
	if err != nil {
		return nil, err
	}

	resultDir := utils.GetTestResultDirectory()
	resultDir = filepath.Join(resultDir, test.testName)
	_ = os.MkdirAll(resultDir, 0755)
	fileName := test.testName
	if strings.TrimSpace(test.testCategory) != "" {
		fileName = fileName + "-" + test.testCategory
	}
	fileName = fileName + "-" + test.configName
	fileName = utils.GetOldTimeStampedFileName(fileName, test.reportTime) + "-FAILED.pdf"
	fileName = filepath.Join(resultDir, fileName)

	err = os.Rename(pdfPath, fileName)
	if err != nil {
		return nil, err
	}
	fileName = strings.Replace(fileName, utils.Config.BaseFolder, "", 1)

	csvPath := test.getRelativeFilenames()
	err = resultsDB.InsertReport(test.report, fileName, csvPath)
	if err != nil {
		return nil, err
	}

	return test.report.Results, nil
}

func (test *baseTest) getRelativeFilenames() string {
	var relativeFilePaths = make([]string, 0)
	for _, filename := range test.filenames {
		fn := strings.Replace(filename, utils.Config.BaseFolder, "", 1)
		relativeFilePaths = append(relativeFilePaths, fn)
	}
	return strings.Join(relativeFilePaths, ";")
}
func (test *baseTest) getFileData(header []string, rows [][]string) []byte {
	var builder strings.Builder
	builder.WriteString(strings.Join(header, ","))
	builder.WriteString("\n")
	for _, row := range rows {
		builder.WriteString(strings.Join(row, ","))
		builder.WriteString("\n")
	}
	fileData := builder.String()
	return []byte(fileData)
}

func (test *baseTest) addFinalTestInformation(start time.Time) {
	totalTime := time.Since(start)
	totalTime = totalTime.Round(time.Second)
	seconds := int64(totalTime.Seconds())
	minutes := seconds / 60
	seconds = seconds % 60
	test.report.AddTestInformation("Measurement Time", fmt.Sprintf("%02d:%02d", minutes, seconds))
	test.report.SetScreenshots(test.spectra)
}

func (test *baseTest) saveResultsAndCSV(fileIdentifier string, path string) {
	csvDir := utils.GetCSVResultDirectory()
	csvDir = filepath.Join(csvDir, test.testName)
	_ = os.MkdirAll(csvDir, 0755)
	fileName := test.testName
	if strings.TrimSpace(test.testCategory) != "" {
		fileName += "-" + test.testCategory
	}
	if fileIdentifier != "" {
		fileName += "-" + fileIdentifier
	}
	fileName += "-" + test.configName

	fileName = utils.GetOldTimeStampedFileName(fileName, test.reportTime) + ".csv"
	fullPath := filepath.Join(csvDir, fileName)

	_ = os.Rename(path, fullPath)

	test.filenames = append(test.filenames, fullPath)
}

func (test *baseTest) saveResultsToCSV(fileIdentifier string, header []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}
	csvDir := utils.GetCSVResultDirectory()
	csvDir = filepath.Join(csvDir, test.testName)
	_ = os.MkdirAll(csvDir, 0755)
	fileName := test.testName
	if strings.TrimSpace(test.testCategory) != "" {
		fileName += "-" + test.testCategory
	}
	if fileIdentifier != "" {
		fileName += "-" + fileIdentifier
	}
	fileName += "-" + test.configName

	fileName = utils.GetOldTimeStampedFileName(fileName, test.reportTime) + ".csv"
	fullPath := filepath.Join(csvDir, fileName)

	data := test.getFileData(header, rows)
	_ = os.WriteFile(fullPath, data, 0666)

	test.filenames = append(test.filenames, fullPath)
}

func (test *baseTest) describe(ctx context.Context) []string {
	runner := &StepRunner{
		Describe: true,
		Test:     test,
		Ctx:      ctx,
	}
	_ = test.impl.measure(runner)
	return runner.Descriptions
}

func (test *baseTest) Measure(ctx context.Context) error {
	runner := &StepRunner{
		Describe: false,
		Test:     test,
		Ctx:      ctx,
	}
	return test.impl.measure(runner)
}

func (test *baseTest) SetParameters(params map[string]interface{}) error {
	return nil
}
