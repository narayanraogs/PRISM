package server

import (
	"prismServer/tne"

	"github.com/gin-gonic/gin"
)

func upDownConverterMeasurement(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var req UCDCRequest
	if err := conn.ReadJSON(&req); err != nil {
		conn.WriteJSON(tne.RTStatus{
			Message:   "Unable to read request",
			Completed: true,
			Error:     true,
			Success:   false,
		})
		return
	}

	var uc tne.UpDownConverterMeasurement
	uc.Init(req.DeviceProfile, req.ExternalSGName, req.ConverterName)
	uc.SetInputCableLoss(req.InputCableLoss, req.InputPower)
	uc.SetOutputCableLoss(req.OutputCableLoss)
	uc.SetLOCableLoss(req.LOCableLoss)

	uc.SetPowerSpectrum(req.PowerSpectrum.Span, req.PowerSpectrum.RBW, req.PowerSpectrum.VBW)
	uc.SetFrequencySpectrum(req.FrequencySpectrum.Span, req.FrequencySpectrum.RBW, req.FrequencySpectrum.VBW)
	uc.SetInBandSpectrum(req.InBandSpectrum.Span, req.InBandSpectrum.RBW, req.InBandSpectrum.VBW)
	uc.SetOutBandSpectrum(req.OutBandSpectrum.Span, req.OutBandSpectrum.RBW, req.OutBandSpectrum.VBW)

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "abort" {
				uc.Stop()
			}
		}
	}()

	for _, test := range req.TestsSelected {
		sts, results := uc.GetStatusMonitor()
		switch test {
		case tne.UCDCGainInternalCable:
			uc.OutputGainMeasurement(req.StepSize, true)
		case tne.UCDCGainInternalRadiated:
			uc.OutputGainMeasurement(req.StepSize, false)
		case tne.UCDCFreqMeas:
			uc.OutputFrequencyMeasurement()
		case tne.UCDCHarmonicMeas:
			uc.OutputHarmonicsMeasurement()
		case tne.UCDCSpuriousInBand:
			uc.OutputSpuriousMeasurement(true)
		case tne.UCDCSpuriousOutBand:
			uc.OutputSpuriousMeasurement(false)
		case tne.UCDCLOLeakage:
			uc.LOLeakageMeasurement()
		case tne.UCDCInputLeakage:
			uc.OutputInputLeakageMeasurement()
		case tne.UCDCGainExternalCable:
			uc.OutputExtLOGainMeasurement(req.StepSize, true)
		case tne.UCDCGainExternalRadiated:
			uc.OutputExtLOGainMeasurement(req.StepSize, false)
		case tne.UCDCOutputMonPower:
			uc.MonitorPowerMeasurement(true)
		case tne.UCDCInputMonPower:
			uc.MonitorPowerMeasurement(false)
		case tne.UCDCLOMonPower:
			uc.LOMonFreqPowerMeasurement()
		case tne.UCDCLOMonPhaseNoise:
			uc.LOMonPhaseNoiseMeasurement()
		case tne.UCDCExtLOPowerMatch:
			uc.ExtLOPowerMatch()
		}

		done := false
		for !done {
			select {
			case stsMsg, ok := <-sts:
				if !ok {
					done = true
					break
				}
				if err := conn.WriteJSON(stsMsg); err != nil {
					return
				}
				if stsMsg.Completed {
					done = true
				}
			case result, ok := <-results:
				if !ok {
					continue
				}
				if err := conn.WriteJSON(result); err != nil {
					return
				}
			}
		}
	}
}
