package server

import (
	"prismServer/tne"
	"time"

	"github.com/gin-gonic/gin"
)

func conductGTxTne(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	var req GTxTneRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteJSON(tne.RTStatus{
			Message:   "Unable to read request",
			Error:     true,
			Completed: true,
			Success:   false,
		})
		return
	}
	if !TryLockOperation() {
		conn.WriteJSON(tne.RTStatus{
			Message:   "System Busy",
			Error:     true,
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

	var gtx = tne.NewGTxGroundTransmitterMeasurement()
	gtx.SetDevices(req.DeviceProfile, req.Component, req.IntermediateFrequency, req.CableLoss)
	gtx.SetModulationParameters(req.ModulationScheme, req.SubCarrierFrequency, req.FrequencyDeviation, req.ModIndex)
	gtx.SetFrequencySpectrum(req.FrequencySpectrum.Span, req.FrequencySpectrum.RBW, req.FrequencySpectrum.VBW)
	gtx.SetPowerSpectrum(req.PowerSpectrum.Span, req.PowerSpectrum.RBW, req.PowerSpectrum.VBW)
	gtx.SetInBandSpectrum(req.InBandSpectrum.Span, req.InBandSpectrum.RBW, req.InBandSpectrum.VBW)
	gtx.SetOutBandSpectrum(req.OutBandSpectrum.Span, req.OutBandSpectrum.RBW, req.OutBandSpectrum.VBW)

	status, results := gtx.GetStatusMonitor()
	go gtx.StartMeasurement()
	go func() {
		defer gtx.Stop()
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

outerFor:
	for {
		select {
		case status := <-status:
			err := conn.WriteJSON(status)
			if err != nil {
				gtx.Stop()
				break outerFor
			}
			if status.Completed {
				break outerFor
			}
		case result := <-results:
			err := conn.WriteJSON(result)
			if err != nil {
				gtx.Stop()
				break outerFor
			}
		}
	}

}
