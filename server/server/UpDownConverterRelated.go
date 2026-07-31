package server

import (
	"encoding/json"
	"net/http"
	"prismServer/resultsDB"
	"prismServer/tne"

	"github.com/gin-gonic/gin"
)

func getUCDCResults(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"ok": false, "message": "name is required"})
		return
	}

	history, err := resultsDB.GetAllResultsForConverter(name)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "Failed to get history"})
		return
	}

	type ResultEntry struct {
		ID       int64                `json:"id"`
		Name     string               `json:"name"`
		TestType string               `json:"testType"`
		Date     string               `json:"date"`
		Time     string               `json:"time"`
		Results  tne.ConvertorResults `json:"results"`
	}

	var resp struct {
		OK      bool          `json:"ok"`
		History []ResultEntry `json:"history"`
	}
	resp.OK = true
	resp.History = make([]ResultEntry, 0)

	for _, h := range history {
		var res tne.ConvertorResults
		err := json.Unmarshal([]byte(h.Results), &res)
		if err != nil {
			continue
		}
		resp.History = append(resp.History, ResultEntry{
			ID:       h.ID,
			Name:     h.Name,
			TestType: h.TestType,
			Date:     h.Date.String,
			Time:     h.Time.String,
			Results:  res,
		})
	}

	c.IndentedJSON(http.StatusOK, resp)
}

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
	sts, _ := uc.GetStatusMonitor()
	if !uc.Init(req.DeviceProfile, req.ExternalSGName, req.ConverterName) {
		for stsMsg := range sts {
			conn.WriteJSON(stsMsg)
		}
		return
	}
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
			go uc.OutputGainMeasurement(req.StepSize, true)
		case tne.UCDCGainInternalRadiated:
			go uc.OutputGainMeasurement(req.StepSize, false)
		case tne.UCDCFreqMeas:
			go uc.OutputFrequencyMeasurement()
		case tne.UCDCHarmonicMeas:
			go uc.OutputHarmonicsMeasurement()
		case tne.UCDCSpuriousInBand:
			go uc.OutputSpuriousMeasurement(true)
		case tne.UCDCSpuriousOutBand:
			go uc.OutputSpuriousMeasurement(false)
		case tne.UCDCLOLeakage:
			go uc.LOLeakageMeasurement()
		case tne.UCDCInputLeakage:
			go uc.OutputInputLeakageMeasurement()
		case tne.UCDCGainExternalCable:
			go uc.OutputExtLOGainMeasurement(req.StepSize, true)
		case tne.UCDCGainExternalRadiated:
			go uc.OutputExtLOGainMeasurement(req.StepSize, false)
		case tne.UCDCOutputMonPower:
			go uc.MonitorPowerMeasurement(true)
		case tne.UCDCInputMonPower:
			go uc.MonitorPowerMeasurement(false)
		case tne.UCDCLOMonPower:
			go uc.LOMonFreqPowerMeasurement()
		case tne.UCDCLOMonPhaseNoise:
			go uc.LOMonPhaseNoiseMeasurement()
		case tne.UCDCExtLOPowerMatch:
			go uc.ExtLOPowerMatch()
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

func upDownConverterResult(c *gin.Context) {
	type resultRequest struct {
		Name  string
		Dates []string
		Times []string
	}

	var req resultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var ack Ack
		ack.OK = false
		ack.Message = "Invalid request"
		c.IndentedJSON(http.StatusBadRequest, ack)
		return
	}

	uc := tne.UpDownConverterMeasurement{}
	pdf, ok := uc.GeneratePDF(req.Name, req.Dates, req.Times)
	var ack Ack
	ack.OK = ok
	ack.Message = pdf
	c.IndentedJSON(http.StatusOK, ack)
}
