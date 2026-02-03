package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type actionHandler func(c *client, request actionRequest) (message string, ok bool)

var actionHandlers map[string]actionHandler

func init() {
	actionHandlers = map[string]actionHandler{
		"SetTSMRoute": handleSetTSMRoute,
		"SetTSMAttn":  handleSetTSMAttn,
		"SetSpectrum": handleSetSpectrum,

		"MeasureCableLoss":            handleMeasureCableLoss,
		"MeasureTVACCableLoss":        handleMeasureTVACCableLoss,
		"AbortCableLossMeasurement":   handleAbortCableLossMeasurement,
		"MeasureSGPower":              handleMeasureSGPower,
		"AbortSGPowerMeasurement":     handleAbortSGPowerMeasurement,
		"MeasureGTxAttn":              handleMeasureGTxAttn,
		"AbortGTxAttnMeasurement":     handleAbortGTxAttnMeasurement,
		"MeasureTSMAttn":              handleMeasureTSMAttn,
		"AbortTSMAttnMeasurement":     handleAbortTSMAttnMeasurement,
		"MeasureTSMInternalPathLoss":  handleMeasureTSMInternalPathLoss,
		"AbortTSMInternalMeasurement": handleAbortTSMInternalMeasurement,
		"SpectrumDump":                handleSpectrumDump,
		"TraceDump":                   handleTraceDump,
		"Screenshot":                  handleScreenshot,
		"PMPower":                     handlePMPower,
		"StartStability":              handleStartStability,
		"StopStability":               handleStopStability,
		"StartTests":                  handleStartTests,
		"AbortTest":                   handleAbortTest,
		"GTxMeasurement":              handleGTxMeasurement,

		"CreateTSMInternalLossTable": handleCreateTSMInternalLossTable,
		"SelectExistingTestPhase":    handleSelectExistingTestPhase,
		"AddNewTestPhase":            handleAddNewTestPhase,
		"SaveSpectrum":               handleSaveSpectrum,
		"SaveSchedule":               handleSaveSchedule,
		"SaveDownlinkLoss":           handleSaveDownlinkLoss,
		"SaveUplinkLoss":             handleSaveUplinkLoss,
		"RegenerateReport":           handleRegenerateReport,
		"SetPpmParameters":           handleSetPpmParameters,
		"SetVsaParameters":           handleSetVsaParameters,
	}
}

func act(c *gin.Context) {
	var request actionRequest
	var response ackResponse
	response.OK = false
	response.Message = ""

	if err := c.BindJSON(&request); err != nil {
		response.OK = false
		response.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	ss := sessions.getServer(request.ID)
	if ss == nil {
		response.OK = false
		response.Message = "Client Not registered"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	response = conductAction(ss, request)
	c.IndentedJSON(http.StatusOK, response)
}

func conductAction(c *client, request actionRequest) ackResponse {
	if handler, ok := actionHandlers[request.Action]; ok {
		msg, success := handler(c, request)
		return ackResponse{OK: success, Message: msg}
	}
	return ackResponse{OK: false, Message: fmt.Sprintf("Unknown action: %s", request.Action)}
}
