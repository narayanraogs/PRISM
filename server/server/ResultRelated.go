package server

import (
	"net/http"
	"prismServer/reports"
	"prismServer/resultsDB"
	"prismServer/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func getResultMetadata(c *gin.Context) {
	var resp ReportsResponse
	results, err := resultsDB.GetAllResults()
	if err != nil {
		resp.OK = false
		resp.Message = "Unable to get results"
		c.JSON(200, resp)
		return
	}
	for _, result := range results {
		resp.Reports = append(resp.Reports, ReportMetadata{
			Date:         result.Date,
			Time:         result.Time,
			TestType:     result.TestType,
			Config:       result.ConfigName,
			TestCategory: result.TestCategory.String,
			Phase:        result.TestPhase,
			Remarks:      result.Remark.String,
			VSAUsed:      strings.EqualFold(result.TestCategory.String, "vsa"),
			PPMUsed:      strings.EqualFold(result.TestCategory.String, "ppm"),
			Success:      !strings.Contains(result.Report, `"OK": false`) && !strings.Contains(result.Report, `"OK":false`),
		})
	}
	resp.AllPPMParams = utils.GetAllPpmParameters()
	resp.SelectedPPMParams = utils.GetSelectedPPMParams()
	resp.AllVSAParams = utils.GetAllVsaParameters()
	resp.SelectedVSAParams = utils.GetSelectedVSAParams()
	resp.OK = true
	resp.Message = "Success"

	c.JSON(200, resp)
}

func getReportPDF(c *gin.Context) {
	var req ReportPDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	report, err := resultsDB.GetReportPDF(req.Date, req.Time)
	if err != nil {
		c.IndentedJSON(http.StatusOK, Ack{OK: false, Message: "Failed to fetch report"})
		return
	}
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: report})
}

func regenerateReport(c *gin.Context) {
	var req RegenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	utils.SetSelectedPPMParameter(req.PPMParameters)
	utils.SetSelectedVSAParameter(req.VSAParameters)
	report, err := resultsDB.RegenerateReport(req.Date, req.Time)
	if err != nil {
		c.IndentedJSON(http.StatusOK, Ack{OK: false, Message: "Failed to regenerate report"})
		return
	}
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: report})

}

func getStabilityReports(c *gin.Context) {
	var resp StabilityReportsMetadata
	resp = getStabilityReportsMetadata()
	c.JSON(200, resp)
}

func getReportsData(c *gin.Context) {
	var req ReportsDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	var resp ReportsDataResponse
	resp.Reports = make([]reports.Report, 0)
	for _, session := range req.Sessions {
		report, err := resultsDB.GetReportJSON(session.Date, session.Time)
		if err != nil {
			continue
		}
		resp.Reports = append(resp.Reports, report)
	}
	resp.OK = true
	resp.Message = "Success"
	c.JSON(http.StatusOK, resp)
}

