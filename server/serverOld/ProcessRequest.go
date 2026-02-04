package server

import (
	"fmt"
	"net/http"
	"prismServer/resultsDB"

	"github.com/gin-gonic/gin"
)

func register(c *gin.Context) {
	var request emptyRequest
	var response ackResponse
	response.OK = false
	response.Message = ""

	if err := c.BindJSON(&request); err != nil {
		response.OK = false
		response.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	fmt.Println("Client ID", request.ID)
	s := sessions.getServer(request.ID)
	if s == nil {
		s = sessions.createServer(request.ID)
	}
	response.OK = true

	c.IndentedJSON(http.StatusOK, response)
}

func getTableValues(c *gin.Context) {
	var request valueRequest
	var response tableValueResponse
	response.OK = false
	response.Message = ""

	if err := c.BindJSON(&request); err != nil {
		response.OK = false
		response.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	s := sessions.getServer(request.ID)
	if s == nil {
		response.OK = false
		response.Message = "Client Not registered"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	response = processTableValueRequest(s, request)
	c.IndentedJSON(http.StatusOK, response)
}

func processTableValueRequest(c *client, request valueRequest) tableValueResponse {
	var values [][]string
	var ok bool
	var message string
	switch request.ParameterName {
	case "CableLossTable":
		values, ok = resultsDB.GetAllCableLosses()
		if ok {
			values = values[2:]
		}
		message = "Unable to read from Cable Loss Table"
	case "TSMInternalLossTable":
		values, ok = resultsDB.GetTSMInternalLossTable()
		if len(values) == 0 {
			ok = false
		}
		message = "No entries in TSM loss table, click Re-Generate table"
	case "TVACCableLossTable":
		//values, ok = resultsDB.GetAllTVACCableLosses()
		values = make([][]string, 24)
		if ok {
			values = values[2:]
		}
		message = "Unable to read from TVAC Cable Loss Table"
	case "OfflineResultsTable":
		values, message, ok = getOfflineResultsTable(c)
	}
	var val tableValueResponse
	if !ok {
		val.OK = false
		val.Values = make([][]string, 0)
		val.Message = message
	} else {
		val.OK = true
		val.Values = values
	}
	return val
}

func getOfflineResultsTable(c *client) ([][]string, string, bool) {
	tp, tpOk := c.global.GetParams("OfflineTestPhase")
	cfg, cfgOk := c.global.GetParams("OfflineConfig")
	tn, tnOk := c.global.GetParams("OfflineTestName")
	tc, tcOk := c.global.GetParams("OfflineTestCategory")
	dt, dtOk := c.global.GetParams("OfflineDate")
	if !tpOk || !cfgOk || !tnOk || !tcOk || !dtOk {
		return nil, "Required Parameter Not Set", false
	}
	table, err := resultsDB.GetResultsTable(tp[0], cfg[0], tn[0], tc[0], dt[0])
	if err != nil {
		return nil, err.Error(), false
	}
	return table, "", true
}
