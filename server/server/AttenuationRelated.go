package server

import (
	"prismServer/database"
	"prismServer/tne"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getAttnMetadata(c *gin.Context) {
	var attnMetaData AttnMetaData
	deviceProfiles, ok := database.GetDeviceProfileNames()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get device profiles"
		c.JSON(200, attnMetaData)
		return
	}
	attnMetaData.DeviceProfile = deviceProfiles
	rxNames, ok := database.GetReceiverNames()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get receiver names"
		c.JSON(200, attnMetaData)
		return
	}
	attnMetaData.Receiver = rxNames
	spectrumProfileNames, ok := database.GetAllSpectrumProfiles()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get spectrum profile names"
		c.JSON(200, attnMetaData)
		return
	}
	attnMetaData.SprectrumProfile = spectrumProfileNames
	tsmConfigNames, ok := database.GetTSMConfigurations()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get TSM config names"
		c.JSON(200, attnMetaData)
		return
	}
	attnMetaData.TSMConfig = tsmConfigNames
	attnMetaData.GTxComponents = []string{"IFM-1", "IFM-2"}
	attnMetaData.AttnRanges = map[string]AttnRange{
		"TSM": {Max: 63.5, Min: 0, StepSize: 0.1},
		"GTx": {Max: 0, Min: -50, StepSize: 0.5},
		"SG":  {Max: 0, Min: -80, StepSize: 1},
	}
	attnMetaData.OK = true
	attnMetaData.Message = "Successfully retrieved metadata"
	c.JSON(200, attnMetaData)
}

func measureAttn(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	var req AttnRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteJSON(tne.AttnMeasurementStatus{
			Message:   "Unable to read request",
			Error:     true,
			Completed: true,
		})
		return
	}

	if !TryLockOperation() {
		conn.WriteJSON(tne.AttnMeasurementStatus{
			Message:   "System Busy",
			Error:     true,
			Completed: true,
		})
		return
	}
	defer UnlockOperation()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	switch strings.ToLower(req.Type) {
	case "tsm":
		tsmAttn := tne.TSMAttnMeasurement{}
		tsmAttn.Initialize(req.DeviceProfile, req.Receiver, req.SpectrumProfile, req.TSMConfig,
			req.Max, req.Min, req.Step)
		monitor := tsmAttn.GetStatusMonitor()
		go tsmAttn.StartMeasurement()
		go func() {
			defer tsmAttn.Stop()
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
		success := true
		for s := range monitor {
			success = !s.Error
			resp := AttnProgressResponse{
				MeasurementStatus: s,
				Deviations:        nil,
				OK:                true,
				Message:           "Successfully retrieved TSM Measurement",
			}
			if err := conn.WriteJSON(resp); err != nil {
				tsmAttn.Stop()
				return
			}
		}
		if success {
			dev := tsmAttn.GetCorrectedDeviations()
			resp := AttnProgressResponse{
				MeasurementStatus: tne.AttnMeasurementStatus{
					Completed: true,
					Error:     false,
					Message:   "Successfully retrieved TSM Measurement",
				},
				Deviations: dev,
				OK:         true,
				Message:    "Successfully retrieved TSM Measurement",
			}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	case "gtx":
		gtxAttn := tne.GTxAttnMeasurement{}
		gtxAttn.Initialize(req.DeviceProfile, req.Receiver, req.SpectrumProfile, req.Component,
			req.Max, req.Min, req.Step)
		monitor := gtxAttn.GetStatusMonitor()
		go gtxAttn.StartMeasurement()
		go func() {
			defer gtxAttn.Stop()
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
		success := true
		for s := range monitor {
			success = !s.Error
			resp := AttnProgressResponse{
				MeasurementStatus: s,
				Deviations:        nil,
				OK:                true,
				Message:           "Successfully retrieved GTx Measurement",
			}
			if err := conn.WriteJSON(resp); err != nil {
				gtxAttn.Stop()
				return
			}
		}
		if success {
			dev := gtxAttn.GetCorrectedDeviations()
			resp := AttnProgressResponse{
				MeasurementStatus: tne.AttnMeasurementStatus{
					Completed: true,
					Error:     false,
					Message:   "Successfully retrieved GTx Measurement",
				},
				Deviations: dev,
				OK:         true,
				Message:    "Successfully retrieved GTx Measurement",
			}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	case "sg":
		sgPower := tne.SGPowerMeasurement{}
		sgPower.Initialize(req.DeviceProfile, req.Receiver, req.SpectrumProfile,
			req.Max, req.Min, req.Step)
		monitor := sgPower.GetStatusMonitor()
		go sgPower.StartMeasurement()
		go func() {
			defer sgPower.Stop()
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
		success := true
		for s := range monitor {
			success = !s.Error
			resp := AttnProgressResponse{
				MeasurementStatus: s,
				Deviations:        nil,
				OK:                true,
				Message:           "Successfully retrieved SG Measurement",
			}
			if err := conn.WriteJSON(resp); err != nil {
				sgPower.Stop()
				return
			}
		}
		if success {
			dev := sgPower.GetCorrectedDeviations()
			resp := AttnProgressResponse{
				MeasurementStatus: tne.AttnMeasurementStatus{
					Completed: true,
					Error:     false,
					Message:   "Successfully retrieved SG Measurement",
				},
				Deviations: dev,
				OK:         true,
				Message:    "Successfully retrieved SG Measurement",
			}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	}
}
