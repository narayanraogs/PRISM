package server

import (
	"encoding/json"
	"prismServer/database"
	"prismServer/resultsDB"
	"prismServer/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getTVACCableLossMetadata(c *gin.Context) {
	var resp TVACCableLossMetadata
	resp.Frequencies = make([]float64, 0)
	resp.IsPMZeroed = resultsDB.CheckIfTVACCableLossPMReferenceExists()
	names, ok := database.GetLossMeasurementFrequencyNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get frequencies from LossMeasurementFrequencies"
		return
	}
	for _, name := range names {
		frequency, ok := database.GetFrequencyForLossMeasurement(name)
		if !ok {
			resp.OK = false
			resp.Message = "Unable to get frequency for " + name
			return
		}
		resp.Frequencies = append(resp.Frequencies, frequency)
	}
	deviceProfile, ok := database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get device profile names"
		return
	}
	resp.DeviceProfiles = deviceProfile
	cableNames, ok := resultsDB.GetCableNames()
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

func getTVACCableMeasuredDetails(c *gin.Context) {
	var resp TVACCableLossResponse
	values, ok := resultsDB.GetAllTVACCableLosses()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get TVAC cable losses"
		return
	}
	resp.IsPMZeroed = resultsDB.CheckIfTVACCableLossPMReferenceExists()
	resp.History = make([]utilities.TVACCableLossRecord, 0)
	for i, value := range values {
		var record utilities.TVACCableLossRecord
		record.SlNo = i + 1
		record.CableName = value.CableName
		temp := strings.Split(value.TestPhase, ";")
		record.Phase = temp[0]
		if len(temp) > 1 {
			record.CycleName = temp[1]
		} else {
			record.CycleName = ""
		}

		record.Date = value.Date
		record.Time = value.Time
		record.IsReference = strings.EqualFold(value.Reference, "1")
		var measurements = make([]utilities.MeasurementPoint, 0)
		err := json.Unmarshal([]byte(value.Loss), &measurements)
		if err != nil {
			resp.OK = false
			resp.Message = err.Error()
			c.JSON(200, resp)
			return
		}
		record.Measurements = measurements
		resp.History = append(resp.History, record)
		if i == len(values)-1 {
			resp.LatestRecord = record
		}
	}
	resp.OK = true
	resp.Message = "Success"
	c.JSON(200, resp)
}

func measureTVACCableLoss(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var req TVACCableLossRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteJSON(utilities.MeasurementStatus{
			Message: "Unable to read request",
			Error:   true,
		})
		return
	}

	if !TryLockOperation() {
		conn.WriteJSON(utilities.MeasurementStatus{
			Message: "System Busy",
			Error:   true,
		})
		return
	}
	defer UnlockOperation()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	var tvac utilities.TVACCableLossMeasurement
	testPhase := req.Phase + ";" + req.CycleName
	tvac.Initialize(req.Channel, req.DeviceProfile, testPhase)
	sts := tvac.GetStatusMonitor()

	switch strings.ToLower(req.Action) {
	case "pmreference":
		go tvac.MeasurePMReference()
	case "cablereference":
		go tvac.MeasureTVACReference(req.CableName, testPhase)
	case "cableloss":
		go tvac.MeasureTVACCableLoss(req.CableName, testPhase)
	default:
		conn.WriteJSON(utilities.MeasurementStatus{
			Message: "Unknown action: " + req.Action,
			Error:   true,
		})
		return
	}

	go func() {
		defer tvac.Stop()
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
			tvac.Stop()
			return
		}
	}
}
