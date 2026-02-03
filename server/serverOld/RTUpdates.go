package server

import (
	"fmt"
	"prismServer/driver"
	"prismServer/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type rtUpdateHandler func(conn *websocket.Conn, c *client)

var rtUpdateHandlers map[string]rtUpdateHandler

func init() {
	rtUpdateHandlers = map[string]rtUpdateHandler{
		"CurrentTSMDriverStatus":         getCurrentTSMStatus,
		"CableLossMeasurementStatus":     getCableLossMeasurementStatus,
		"TVACCableLossMeasurementStatus": getTVACCableLossMeasurementStatus,
		"SGPowerMeasurementStatus":       getSGPowerMeasurementStatus,
		"GTxPowerMeasurementStatus":      getGTxPowerMeasurementStatus,
		"TSMAttnMeasurementStatus":       getTSMAttnMeasurementStatus,
		"TSMInternalMeasurementStatus":   getTSMInternalMeasurementStatus,
		"GTxMeasurementStatus":           getGTxMeasurementStatus,
	}
}

func rtUpdate(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	var req actionRequest
	err = conn.ReadJSON(&req)
	if err != nil {
		logger.Log.Error("Error reading initial client registration:", err)
		return
	}
	cid := sessions.getServer(req.ID)
	if cid == nil {
		_ = conn.WriteJSON(gin.H{"Error": "Client Not registered"})
		return
	}

	handler, ok := rtUpdateHandlers[req.Action]
	if !ok {
		_ = conn.WriteJSON(gin.H{"Error": fmt.Sprintf("Unknown action: %s", req.Action)})
		return
	}

	handler(conn, cid)
}

func getCurrentTSMStatus(conn *websocket.Conn, cid *client) {
	tsmNames, read := cid.global.GetParams("SelectedTSM")
	if !read {
		_ = conn.WriteJSON(gin.H{"OK": false, "Message": "TSM Not selected"})
		return
	}
	var tsm driver.TSM
	ok := tsm.LoadDevice(tsmNames[0])
	fmt.Println("TSM Selected", tsmNames)
	if !ok {
		_ = conn.WriteJSON(gin.H{"OK": false, "Message": "Unable to read TSM Details from Database"})
		return
	}
	response := tsm.GetDriverPath()
	if !response.Success {
		_ = conn.WriteJSON(gin.H{"OK": false, "Message": "Unable to Connect to TSM to read Driver Status"})
		return
	}
	values := make([]string, 0)
	values = append(values, response.Result["DriverPath"].String)
	time.Sleep(1 * time.Second)
	response = tsm.GetAttn(1)
	if !response.Success {
		_ = conn.WriteJSON(gin.H{"OK": false, "Message": "Unable to Connect to TSM to read Attn 1"})
		return
	}
	values = append(values, response.Result["Attn"].String)
	time.Sleep(1 * time.Second)
	response = tsm.GetAttn(2)
	if !response.Success {
		_ = conn.WriteJSON(gin.H{"OK": false, "Message": "Unable to Connect to TSM to read Attn 2"})
		return
	}
	values = append(values, response.Result["Attn"].String)

	var tsmStatus parameterValue
	tsmStatus.Name = "CurrentTSMStatus"
	tsmStatus.Values = values

	var value getResponse
	value.OK = true
	value.Values = append(value.Values, tsmStatus)

	_ = conn.WriteJSON(value)
}

func getCableLossMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.global.CLM.GetStatusMonitor()
	for msg := range monitor {
		var values = []string{msg.Message}
		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "CableLossMeasurementStatus"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}

func getTVACCableLossMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.global.TCLM.GetStatusMonitor()
	for msg := range monitor {
		var values = []string{msg.Message}
		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "TVACCableLossMeasurementStatus"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}

func getSGPowerMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.global.SGPower.GetStatusMonitor()
	for msg := range monitor {
		var temp = make([]string, 0)
		for _, line := range msg.CurrentStatus {
			temp = append(temp, strings.Join(line, ","))
		}
		var values = []string{strings.Join(temp, ";;;"), msg.Message}

		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "SGPowerMeasurementStatus"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}

func getGTxPowerMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.global.GTXAttn.GetStatusMonitor()
	for msg := range monitor {
		var temp = make([]string, 0)
		for _, line := range msg.CurrentStatus {
			temp = append(temp, strings.Join(line, ","))
		}
		var values = []string{strings.Join(temp, ";;;"), msg.Message}

		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "GTxPowerMeasurementStatus"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}

func getTSMAttnMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.global.TSMAttn.GetStatusMonitor()
	for msg := range monitor {
		var temp = make([]string, 0)
		for _, line := range msg.CurrentStatus {
			temp = append(temp, strings.Join(line, ","))
		}
		var values = []string{strings.Join(temp, ";;;"), msg.Message}

		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "TSMAttnMeasurementStatus"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}

func getTSMInternalMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.global.TSMInternal.GetStatusMonitor()
	for msg := range monitor {
		var values = []string{msg.Message}
		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "TSMInternalMeasurementStatus"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}

func getGTxMeasurementStatus(conn *websocket.Conn, cid *client) {
	monitor := cid.gtxMeasurement.GetStatusMonitor()
	go cid.gtxMeasurement.StartMeasurement()
	for msg := range monitor {
		var temp = make([]string, 0)
		for _, line := range msg.CurrentStatus {
			temp = append(temp, strings.Join(line, ","))
		}
		var values = []string{strings.Join(temp, ";;;"), msg.Message}

		if msg.Completed {
			values = append(values, "Completed")
		} else {
			values = append(values, "In-Progress")
		}
		if msg.Success {
			values = append(values, "Success")
		} else {
			values = append(values, "Failed")
		}

		var value parameterValue
		value.Name = "GTxMeasurement"
		value.Values = values
		_ = conn.WriteJSON(value)
	}
}
