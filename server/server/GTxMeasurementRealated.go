package server

import (
	"prismServer/tne"

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
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "abort" {
				gtx.Stop()
			}
		}
	}()

outerFor:
	for {
		select {
		case status := <-status:
			conn.WriteJSON(status)
			if status.Completed {
				break outerFor
			}
		case result := <-results:
			conn.WriteJSON(result)
		}
	}

}
