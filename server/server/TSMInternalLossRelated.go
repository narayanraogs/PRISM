package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"prismServer/database"
	"prismServer/resultsDB"
	"prismServer/tne"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type lossMeasured struct {
	Frequency []string
	Measured  []string
}

func getTSMInternalLossMetadata(c *gin.Context) {
	var resp TSMInternalLossMetadata
	var ok bool
	resp.DeviceProfile, ok = database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get device profile names"
		c.JSON(http.StatusOK, resp)
		return
	}
	resp.MeasuredLoss, ok = getTSMLossTable()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get read Loss table from database"
		c.JSON(http.StatusOK, resp)
		return
	}
	resp.OK = true
	resp.Message = "Success"
	c.JSON(http.StatusOK, resp)
}

func getTSMLossTable() (TSMInternalLossMeasured, bool) {
	var resp TSMInternalLossMeasured
	loss, err := resultsDB.GetTSMInternalLossStructure()
	if err != nil {
		slog.Error("Failed to get TSM internal loss structure", "error", err)
		return resp, false
	}
	resp.Paths = make([]InternalLossEntry, 0)
	for _, l := range loss {
		if strings.EqualFold(l.PathMnemonic.String, "PM-Measurement") {
			var pm InternalLossPMOrCableEntry
			pm, ok := getPMOrCableLoss(l)
			if !ok {
				slog.Error("Failed to unmarshal PM-Measurement")
				return resp, false
			}
			resp.PM = pm
		} else if strings.EqualFold(l.PathMnemonic.String, "Cable-Measurement") {
			var cable InternalLossPMOrCableEntry
			cable, ok := getPMOrCableLoss(l)
			if !ok {
				slog.Error("Failed to unmarshal Cable-Measurement")
				return resp, false
			}
			resp.Cable = cable
		} else {
			var path InternalLossEntry
			path, ok := getInternalLoss(l)
			if !ok {
				slog.Error("Failed to unmarshal path")
				return resp, false
			}
			resp.Paths = append(resp.Paths, path)
		}
	}
	return resp, true
}

func getPMOrCableLoss(entry resultsDB.TSMInternalLoss) (InternalLossPMOrCableEntry, bool) {
	var pm InternalLossPMOrCableEntry
	var in lossMeasured
	err := json.Unmarshal([]byte(entry.MeasuredLosses), &in)
	if err != nil {
		return pm, false
	}
	pm.Measured = strings.EqualFold(entry.MeasurementCompleted, "yes")
	pm.Frequencies = make([]float64, 0)
	pm.Losses = make([]float64, 0)
	for i := 0; i < len(in.Frequency); i++ {
		freq, err := strconv.ParseFloat(in.Frequency[i], 64)
		if err != nil {
			return pm, false
		}
		loss, _ := strconv.ParseFloat(in.Measured[i], 64)
		pm.Frequencies = append(pm.Frequencies, freq)
		pm.Losses = append(pm.Losses, loss)
	}
	return pm, true
}

func getInternalLoss(entry resultsDB.TSMInternalLoss) (InternalLossEntry, bool) {
	var pm InternalLossEntry
	var in lossMeasured
	err := json.Unmarshal([]byte(entry.MeasuredLosses), &in)
	if err != nil {
		return pm, false
	}
	pm.InputPort = entry.InputPort
	pm.OutputPort = entry.OutputPort
	pm.PathMnemonic = entry.PathMnemonic.String
	pm.Measured = strings.EqualFold(entry.MeasurementCompleted, "yes")
	pm.Frequencies = make([]float64, 0)
	pm.Losses = make([]float64, 0)
	for i := 0; i < len(in.Frequency); i++ {
		freq, err := strconv.ParseFloat(in.Frequency[i], 64)
		if err != nil {
			return pm, false
		}
		loss, _ := strconv.ParseFloat(in.Measured[i], 64)
		pm.Frequencies = append(pm.Frequencies, freq)
		pm.Losses = append(pm.Losses, loss)
	}
	return pm, true
}

func measureTSMInternalLoss(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var req InternalLossMeasurementRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteJSON(tne.MeasurementStatus{
			Message:   "Unable to read request",
			Completed: true,
			Success:   false,
		})
		return
	}

	var tsm tne.TSMInternalLoss
	ok := tsm.Initialize(req.DeviceProfile, req.PMChannel)
	if !ok {
		conn.WriteJSON(tne.RTStatus{
			Message:   "Unable to initialize TSM",
			Error:     true,
			Completed: true,
			Success:   false,
		})
		return
	}
	sts := tsm.GetStatusMonitor()

	go tsm.MeasureForConfig(req.Mode, req.InputPort, req.OutputPort)

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "abort" {
				tsm.Stop()
			}
		}
	}()
	var errorFlag = false
	for s := range sts {
		errorFlag = s.Error
		if err := conn.WriteJSON(s); err != nil {
			return
		}
	}
	if !errorFlag {
		table, _ := getTSMLossTable()
		conn.WriteJSON(table)
	}
}
