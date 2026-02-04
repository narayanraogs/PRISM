package server

import (
	"net/http"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/reports"
	"prismServer/resultsDB"
	"prismServer/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func getSpectrumDumpMetadata(c *gin.Context) {
	var stb SpectrumDumpMetadata
	stb.SpectrumDumpMode = []string{"Spectrum Dump", "Screenshot"}
	stb.Instruments = make(map[string][]string)
	stb.SpectrumProfiles = make([]SpectrumProfile, 0)
	sa, ok := database.GetSAAndVSAList()
	if !ok {
		stb.OK = false
		stb.Message = "SA's not present in Database"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	stb.Instruments["SA"] = sa
	vsa, ok := database.GetVSAList()
	if !ok {
		stb.OK = false
		stb.Message = "VSA's not present in Database"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	stb.Instruments["VSA"] = vsa

	sps, ok := database.GetAllSpectrumProfiles()
	if !ok {
		stb.OK = false
		stb.Message = "Cannot get Spectrum Profiles from Database"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	for _, profile := range sps {
		spec, ok := database.GetSpectrumProfile(profile)
		if !ok {
			continue
		}
		var prof SpectrumProfile
		prof.ProfileName = spec.Name
		prof.CenterFrequency = spec.CenterFrequency
		prof.Span = spec.Span
		prof.RBW = float64(spec.RBW)
		prof.VBW = float64(spec.VBW)
		stb.SpectrumProfiles = append(stb.SpectrumProfiles, prof)
	}

	stb.ScreenshotProfiles = []string{"Screenshot", "Magniture", "Pulse Magniture",
		"Pulse Frequency", "Pulse Phase", "Spectrogram"}

	stb.OK = true
	stb.Message = "Success"
	c.IndentedJSON(http.StatusOK, stb)
}

func setSpectrum(c *gin.Context) {
	var req SetSpectrumRequest
	var ack Ack
	if err := c.BindJSON(&req); err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	var dev driver.SA

	ok := dev.LoadDevice(req.SA)
	if !ok {
		ack.OK = false
		ack.Message = "Unable to Load Device [Database Error]"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	resp := dev.SetSpectrum(req.CenterFrequency, req.Span, req.RBW, req.VBW)
	if !resp.Success {
		ack.OK = false
		ack.Message = resp.ErrorMessage
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	if req.AutoReference {
		dev.SetReferenceNominal()
	} else {
		dev.SetReferenceLevel(req.ReferenceLevel)
	}
	switch strings.ToLower(req.Mode) {
	case "clear write":
		dev.SetNormalMode()
	case "max hold":
		dev.SetMaxHold()
	default:
		dev.SetNormalMode()
	}
	ack.OK = true
	ack.Message = "Success"
	c.IndentedJSON(http.StatusOK, ack)
}

func readSpectrum(c *gin.Context) {
	var req ReadSpectrumRequest
	var ack ReadSpectrumResponse
	if err := c.BindJSON(&req); err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	var dev driver.SA

	ok := dev.LoadDevice(req.SA)
	if !ok {
		ack.OK = false
		ack.Message = "Unable to Load Device [Database Error]"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	resp := dev.GetSpectrum()
	if !resp.Success {
		ack.OK = false
		ack.Message = resp.ErrorMessage
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	ack.CenterFrequency = resp.Result["CenterFrequency"].Value
	ack.Span = resp.Result["Span"].Value
	ack.RBW = resp.Result["RBW"].Value
	ack.VBW = resp.Result["VBW"].Value
	ack.ReferenceLevel = resp.Result["ReferenceLevel"].Value
	ack.OK = true
	ack.Message = "Success"
	c.IndentedJSON(http.StatusOK, ack)
}

func dumpSpectrun(c *gin.Context) {
	var req ReadSpectrumRequest
	var ack ReadSpectrumResponse
	if err := c.BindJSON(&req); err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}

	var dev driver.SA
	ok := dev.LoadDevice(req.SA)
	if !ok {
		ack.OK = false
		ack.Message = "Unable to Load Device [Database Error]"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	response := dev.GetSpectrumDump()
	ack.OK = response.Success
	ack.Message = response.Result["SpectrumDump"].String
	c.IndentedJSON(http.StatusOK, ack)
}

func dumpTrace(c *gin.Context) {
	var req DumpTraceRequest
	var ack ReadSpectrumResponse
	if err := c.BindJSON(&req); err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}

	var dev driver.SA
	ok := dev.LoadDevice(req.SA)
	if !ok {
		ack.OK = false
		ack.Message = "Unable to Load Device [Database Error]"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	response := dev.GetTraceDump(req.TracePoints)
	if !response.Success {
		ack.OK = false
		ack.Message = response.ErrorMessage
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	fileData := response.Result["TraceDump"].String
	response = dev.GetNoOfRowsToSkipInTrace()
	if !response.Success {
		ack.OK = false
		ack.Message = response.ErrorMessage
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	noOfRows := response.Result["NoOfRows"].Integer
	ack.Message, ack.OK = utils.GetTracePlot(fileData, noOfRows)
	c.IndentedJSON(http.StatusOK, ack)
}

func saveSpectrum(c *gin.Context) {
	var req SaveSpectrumRequest
	var ack ReadSpectrumResponse
	if err := c.BindJSON(&req); err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}

	report := reports.Report{}
	report.SetHeader("", "Spectrum Dump", "", utils.GetTestPhase())
	var image reports.Images
	image.FileData = req.Spectrum
	image.Caption = "Captured by user"
	report.SetScreenshots([]reports.Images{image})
	report.SetRemarks(req.Remark)
	filename, err := reports.GenerateResult(report)
	if err != nil {
		ack.OK = false
		ack.Message = "Unable to generate Screenshot PDF"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	resultDir := utils.GetTestResultDirectory()
	resultDir = filepath.Join(resultDir, "SpectrumDump")
	_ = os.MkdirAll(resultDir, 0755)
	fileName := "SpectrumDump"
	fileName = utils.GetTimeStampedFileName(fileName) + ".pdf"
	fileName = filepath.Join(resultDir, fileName)

	err = os.Rename(filename, fileName)
	if err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	fileName = strings.Replace(fileName, utils.Config.BaseFolder, "", 1)

	err = resultsDB.InsertReport(report, fileName, "")
	if err != nil {
		ack.OK = false
		ack.Message = "Unable to save Report in database"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	ack.OK = true
	ack.Message = "Screenshot Saved to " + fileName
	c.IndentedJSON(http.StatusOK, ack)
}

func dumpScreenshot(c *gin.Context) {
	var req DumpScreenshotRequest
	var ack ReadSpectrumResponse
	if err := c.BindJSON(&req); err != nil {
		ack.OK = false
		ack.Message = err.Error()
		c.IndentedJSON(http.StatusOK, ack)
		return
	}

	var dev driver.VSA
	ok := dev.LoadDevice(req.VSA)
	if !ok {
		ack.OK = false
		ack.Message = "Unable to Load Device [Database Error]"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	response := dev.GetScreenshot(req.Mode)
	if !response.Success {
		ack.OK = false
		ack.Message = response.ErrorMessage
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	fileData := response.Result["Screenshot"].String
	ack.OK = response.Success
	ack.Message = fileData
	c.IndentedJSON(http.StatusOK, ack)
}
