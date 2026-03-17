package resultsDB

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"prismServer/executeTest/results"
	"prismServer/reports"
	"prismServer/utils"
	"strings"
)

func InsertReport(report reports.Report, pdfPath string, csvPath string) error {
	ctx := context.Background()
	var args insertResultsParams
	args.SatName = report.Header.Spacecraft
	args.TestPhase = report.Header.TestPhase
	args.TestType = report.Header.TestType
	args.TestCategory.String = report.Header.TestCategory
	args.TestCategory.Valid = true
	args.ConfigName = report.Header.Config
	args.Date = report.Header.Date
	args.Time = report.Header.Time
	args.Remark.String = report.Remarks
	args.Remark.Valid = true
	args.FilePath = pdfPath
	args.CSVFilePath.String = csvPath
	args.CSVFilePath.Valid = true

	data, err := json.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	args.Report = string(data)

	return dbObject.insertResults(ctx, args)
}

func GetResultsTable(tp string, config string, testName string, testCategory string, date string) ([][]string, error) {
	ctx := context.Background()
	var args getResultsParams

	args.TestPhase = tp
	if strings.EqualFold(tp, "all") {
		args.TestPhase = "%"
	}
	args.TestType = testName
	if strings.EqualFold(testName, "all") {
		args.TestType = "%"
	}
	args.TestCategory.String = testCategory
	if strings.EqualFold(testCategory, "all") {
		args.TestCategory.String = "%"
	}
	args.TestCategory.Valid = true
	args.ConfigName = config
	if strings.EqualFold(config, "all") {
		args.ConfigName = "%"
	}
	args.Date = date
	if strings.EqualFold(date, "all") {
		args.Date = "%"
	}
	res, err := dbObject.getResults(ctx, args)
	if err != nil {
		return nil, err
	}
	var values = make([][]string, 0)
	for _, result := range res {
		var row = make([]string, 0)
		row = append(row, result.ConfigName, result.TestType, result.TestCategory.String)
		row = append(row, result.Date, result.Time, result.Remark.String)
		values = append(values, row)
	}
	return values, nil
}
func GetAllResults() ([]Result, error) {
	ctx := context.Background()
	var args getResultsParams

	args.TestPhase = "%"
	args.TestType = "%"
	args.TestCategory.String = "%"
	args.TestCategory.Valid = true
	args.ConfigName = "%"
	args.Date = "%"

	res, err := dbObject.getResults(ctx, args)
	return res, err
}

func GetReportPDF(date string, time string) (string, error) {
	ctx := context.Background()
	var args getSingleResultParams
	args.Date = date
	args.Time = time
	result, err := dbObject.getSingleResult(ctx, args)
	if err != nil {
		return "", err
	}
	path := result.FilePath
	path = filepath.Join(utils.Config.BaseFolder, path)
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(file), nil
}

func RegenerateReport(date string, time string) (string, error) {
	ctx := context.Background()
	var args getSingleResultParams
	args.Date = date
	args.Time = time
	storedResult, err := dbObject.getSingleResult(ctx, args)
	if err != nil {
		return "", err
	}

	var reportData reports.Report
	err = json.Unmarshal([]byte(storedResult.Report), &reportData)
	if err != nil {
		return "", err
	}

	var absolutePaths []string
	for _, relPath := range reportData.Filenames {
		absPath := filepath.Join(utils.Config.BaseFolder, relPath)
		absolutePaths = append(absolutePaths, absPath)
	}

	regeneratedResults, order, plots, err := results.GenerateResults(reportData.Header.TestType, absolutePaths)
	if err != nil {
		return "", err
	}

	reportData.Results = regeneratedResults
	reportData.Order = order
	reportData.Screenshots = append(reportData.Screenshots, plots...)

	newPDFPath, err := reports.GenerateResult(reportData, true, true, true, true, true)
	if err != nil {
		return "", err
	}

	updatedReportJSON, err := json.MarshalIndent(reportData, "", " ")
	if err != nil {
		return "", err
	}

	var updateArgs updateResultParams
	updateArgs.Date = date
	updateArgs.Time = time
	updateArgs.Report = string(updatedReportJSON)
	updateArgs.FilePath = newPDFPath

	err = dbObject.updateResult(ctx, updateArgs)
	if err != nil {
		return "", err
	}

	return newPDFPath, nil
}

func GetOfflineTestPhases() ([]string, error) {
	ctx := context.Background()
	return dbObject.getOfflineTestPhases(ctx)
}
