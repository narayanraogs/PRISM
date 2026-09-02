package server

import (
	"encoding/json"
	"fmt"
	"prismServer/database"
	"prismServer/resultsDB"
	"prismServer/tne"
	"prismServer/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getCableLossMetadata(c *gin.Context) {
	var resp CableLossMetadata
	resp.Frequencies = make([]string, 0)
	resp.IsPMZeroed = resultsDB.CheckIfCableLossPMReferenceExists()
	names, ok := database.GetLossMeasurementFrequencies()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get frequencies from LossMeasurementFrequencies"
		return
	}
	resp.Frequencies = names
	deviceProfile, ok := database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get device profile names"
		return
	}
	resp.DeviceProfiles = deviceProfile
	cableNames, ok := resultsDB.GetCableNamesForCableLoss()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get cable names"
		return
	}
	resp.ExistingCables = cableNames
	resp.OK = true
	resp.Message = "Success"
	c.JSON(200, resp)
}

func getCableMeasuredDetails(c *gin.Context) {
	var resp CableLossHistoryResponse
	resp.History = make([]tne.CableLossRecord, 0)
	values, err := resultsDB.GetAllCableLosses()
	if err != nil {
		resp.OK = false
		resp.Message = "Unable to get TVAC cable losses"
		return
	}
	for i, value := range values {
		var item tne.CableLossRecord
		item.SlNo = i + 1
		item.CableName = value.CableName
		item.Length = value.CableLength
		item.Date = value.Date
		item.Time = value.Time
		var measurements = make([]tne.MeasurementPoint, 0)
		err := json.Unmarshal([]byte(value.Loss), &measurements)
		if err != nil {
			resp.OK = false
			resp.Message = err.Error()
			c.JSON(200, resp)
			return
		}
		item.Measurements = measurements
		resp.History = append(resp.History, item)

	}
	resp.OK = true
	resp.Message = "Success"
	c.JSON(200, resp)
}

func measureCableLoss(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var req CableLossRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteJSON(tne.RTStatus{
			Message:   "Unable to read request",
			Completed: true,
			Success:   false,
		})
		return
	}

	if !TryLockOperation() {
		conn.WriteJSON(tne.RTStatus{
			Message:   "System Busy",
			Completed: true,
			Success:   false,
		})
		return
	}
	defer UnlockOperation()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	var cableLoss tne.CableLossMeasurement
	cableLoss.Initialize(req.Channel, req.DeviceProfile, req.SelectedFrequencies)
	sts := cableLoss.GetStatusMonitor()

	switch strings.ToLower(req.Action) {
	case "pmreference":
		go cableLoss.MeasurePMReference()
	case "cableloss":
		go cableLoss.MeasureCableLoss(req.CableName, fmt.Sprintf("%.2f", req.CableLength))
	default:
		conn.WriteJSON(utilities.MeasurementStatus{
			Message: "Unknown action: " + req.Action,
			Error:   true,
		})
		return
	}

	go func() {
		defer cableLoss.Stop()
		for {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "abort" {
				return
			}
		}
	}()
	for s := range sts {
		if err := conn.WriteJSON(s); err != nil {
			cableLoss.Stop()
			return
		}
	}
}
