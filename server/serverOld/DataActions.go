package server

import (
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/reports"
	"prismServer/resultsDB"
	"prismServer/utils"
	"strings"
)

func handleCreateTSMInternalLossTable(c *client, request actionRequest) (string, bool) {
	ok := c.global.TSMInternal.CreateNewTable()
	if !ok {
		return "Unable to create TSM Table", false
	}
	return "", true
}

func handleSelectExistingTestPhase(c *client, request actionRequest) (string, bool) {
	tp := request.getParam("TestPhase")
	if tp == nil {
		return "Unable to get Selected TestPhase", false
	}
	return database.SelectExisitingTestPhase(tp[0])
}

func handleAddNewTestPhase(c *client, request actionRequest) (string, bool) {
	tp := request.getParam("TestPhase")
	if tp == nil {
		return "Unable to get Selected TestPhase", false
	}
	return database.InsertNewTestPhase(tp[0], tp[1])
}

func handleSaveSpectrum(c *client, request actionRequest) (string, bool) {
	spectrums := request.getParam("Spectrum")
	if spectrums == nil {
		return "Unable to get Selected Spectrum", false
	}
	remarks := request.getParam("Remark")
	if remarks == nil {
		return "Unable to get Remarks", false
	}
	report := reports.Report{}
	report.SetHeader("", "Spectrum Dump", "", utils.GetTestPhase())
	var image reports.Images
	image.FileData = spectrums[0]
	image.Caption = "Captured by user"
	report.SetScreenshots([]reports.Images{image})
	report.SetRemarks(remarks[0])
	filename, err := reports.GenerateResult(report)
	if err != nil {
		return "Unable to generate Screenshot PDF", false
	}
	resultDir := utils.GetTestResultDirectory()
	resultDir = filepath.Join(resultDir, "SpectrumDump")
	_ = os.MkdirAll(resultDir, 0755)
	fileName := "SpectrumDump"
	fileName = utils.GetTimeStampedFileName(fileName) + ".pdf"
	fileName = filepath.Join(resultDir, fileName)

	err = os.Rename(filename, fileName)
	if err != nil {
		return err.Error(), false
	}
	fileName = strings.Replace(fileName, utils.Config.BaseFolder, "", 1)

	err = resultsDB.InsertReport(report, fileName, "")
	if err != nil {
		return "Unable to save Report in database", false
	}
	return "Screenshot Saved to " + fileName, true
}

func handleSaveSchedule(c *client, request actionRequest) (string, bool) {
	schedules := request.getParam("Schedule")
	if schedules == nil {
		return "Unable to get Schedule", false
	}
	filenames := request.getParam("Filename")
	if filenames == nil {
		return "Unable to get Filename", false
	}
	err := os.WriteFile(filenames[0], []byte(schedules[0]), 0666)
	if err != nil {
		return "Unable to write file", false
	}
	return "", true
}

func handleSaveDownlinkLoss(c *client, request actionRequest) (string, bool) {
	testPhase := request.getParam("TestPhase")
	configuration := request.getParam("Configuration")
	loss := request.getParam("Loss")
	if testPhase == nil || configuration == nil || loss == nil {
		return "Required Parameters are not set", false
	}
	ok := database.UpdateDownlinkLossProfile(configuration[0], testPhase[0], loss[0])
	if !ok {
		return "Not able to update downlink Loss", false
	}
	return "", true
}

func handleSaveUplinkLoss(c *client, request actionRequest) (string, bool) {
	testPhase := request.getParam("TestPhase")
	configuration := request.getParam("Configuration")
	loss := request.getParam("Loss")
	if testPhase == nil || configuration == nil || loss == nil {
		return "Required Parameters are not set", false
	}
	ok := database.UpdateUplinkLossProfile(configuration[0], testPhase[0], loss[0])
	if !ok {
		return "Not able to update uplink Loss", false
	}
	return "", true
}

func handleRegenerateReport(c *client, request actionRequest) (string, bool) {
	date := request.getParam("Date")
	time := request.getParam("Time")
	if date == nil || time == nil {
		return "Time not found in request", false
	}

	pdfPath, err := resultsDB.RegenerateReport(date[0], time[0])
	if err != nil {
		return err.Error(), false
	}

	return "Report regenerated successfully: " + pdfPath, true
}
