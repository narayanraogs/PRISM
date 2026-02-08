package server

import (
	"net/http"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/executeTest/measurements"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getLinkStatus(c *gin.Context) {
	var request RFUplinkRequest
	var rfu LinkStatus
	var ok bool
	var tsmSelected string

	if err := c.BindJSON(&request); err != nil {
		rfu.OK = false
		rfu.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, rfu)
		return
	}

	tsms, ok := database.GetTSMList()
	if ok {
		if strings.EqualFold(request.TSMSelected, "") {
			tsmSelected = tsms[0]
		} else {
			tsmSelected = request.TSMSelected
		}
	} else {
		tsmSelected = ""
	}

	if tsmSelected == "" {
		rfu.TSMConnected = false
		rfu.RemoveConfigs = []string{}
		rfu.SwitchStatus = []string{}
		rfu.Attn1Value = 0
		rfu.Attn2Value = 0
		rfu.OK = false
		rfu.Message = "No TSM Selected"
		c.IndentedJSON(http.StatusOK, rfu)
		return
	}

	rfu.RemoveConfigs = measurements.GetAllActiveUplinkConfigs(tsmSelected)

	var tsm driver.TSM
	ok = tsm.LoadDevice(tsmSelected)
	if !ok {
		rfu.TSMConnected = false
		rfu.SwitchStatus = []string{}
		rfu.Attn1Value = 0
		rfu.Attn2Value = 0
		c.JSON(200, rfu)
		return
	}
	resp := tsm.CheckConnection()
	rfu.TSMConnected = resp.Success
	if !resp.Success {
		rfu.SwitchStatus = []string{}
		rfu.Attn1Value = 0
		rfu.Attn2Value = 0
		c.JSON(200, rfu)
		return
	}
	resp = tsm.GetDriverPath()
	if !resp.Success {
		rfu.SwitchStatus = []string{}

	} else {
		rfu.SwitchStatus = strings.Split(resp.Result["DriverPath"].String, "!")
	}
	resp = tsm.GetAttn(1)
	if !resp.Success {
		rfu.Attn1Value = 0
	} else {
		value := resp.Result["Attn"].String
		value = strings.TrimSpace(value)
		value = strings.ReplaceAll(value, "dB", "")
		rfu.Attn1Value, _ = strconv.ParseFloat(value, 64)
	}
	resp = tsm.GetAttn(2)
	if !resp.Success {
		rfu.Attn2Value = 0
	} else {
		value := resp.Result["Attn"].String
		value = strings.TrimSpace(value)
		value = strings.ReplaceAll(value, "dB", "")
		rfu.Attn2Value, _ = strconv.ParseFloat(value, 64)
	}
	rfu.OK = true
	rfu.Message = "Success"
	c.IndentedJSON(http.StatusOK, rfu)
}

func setTSMRoute(c *gin.Context) {
	var request TSMRouteRequest
	var ack Ack
	if err := c.BindJSON(&request); err != nil {
		ack.OK = false
		ack.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	var tsm driver.TSM
	ok := tsm.LoadDevice(request.TSMSelected)
	if !ok {
		ack.OK = false
		ack.Message = "Failed to load TSM"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	resp := tsm.SetDriverStatus(request.Mnemonic)
	if !resp.Success {
		ack.OK = false
		ack.Message = "Failed to set TSM route"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	ack.OK = true
	ack.Message = "Success"
	c.IndentedJSON(http.StatusOK, ack)
}

func setTSMAttn(c *gin.Context) {
	var request TSMSetAttn
	var ack Ack
	if err := c.BindJSON(&request); err != nil {
		ack.OK = false
		ack.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	var tsm driver.TSM
	ok := tsm.LoadDevice(request.TSMSelected)
	if !ok {
		ack.OK = false
		ack.Message = "Failed to load TSM"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	resp := tsm.SetAttn(request.AttnNo, request.AttnValue)
	if !resp.Success {
		ack.OK = false
		ack.Message = "Failed to set TSM Attn"
		c.IndentedJSON(http.StatusOK, ack)
		return
	}
	ack.OK = true
	ack.Message = "Success"
	c.IndentedJSON(http.StatusOK, ack)
}
