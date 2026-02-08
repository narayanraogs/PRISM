package server

import (
	"fmt"
	"prismServer/driver"
	"prismServer/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func monitor(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Error upgrading connection:", "Error", err)
		return
	}
	defer conn.Close()
	var req MonitorRequest
	var resp MonitorResponse
	err = conn.ReadJSON(&req)
	if err != nil {
		logger.Log.Error("Error reading initial client registration", "Error", err)
		return
	}
	var sa driver.SA
	var vsa driver.VSA
	var pm driver.PM
	var ppm driver.PPM
	var ok bool
	switch strings.ToLower(req.InstrumentType) {
	case "sa":
		ok = sa.LoadDevice(req.Instrument)
	case "vsa":
		ok = vsa.LoadDevice(req.Instrument)
	case "pm":
		ok = pm.LoadDevice(req.Instrument)
	case "ppm":
		ok = ppm.LoadDevice(req.Instrument)
	}
	if !ok {
		resp.OK = false
		resp.Message = "Error loading device"
		conn.WriteJSON(resp)
		return
	}

	// Detection of client closure
	stop := make(chan struct{})
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(stop)
				return
			}
		}
	}()

	for {
		select {
		case <-stop:
			fmt.Println("Monitor: Client disconnected")
			return
		default:
			switch strings.ToLower(req.InstrumentType) {
			case "sa":
				r := sa.GetSpectrumDump()
				if !r.Success {
					resp.OK = false
					resp.Message = r.ErrorMessage
					conn.WriteJSON(resp)
					return
				}
				resp.Image = r.Result["SpectrumDump"].String
				resp.OK = true
				resp.Message = "Success"
				if err := conn.WriteJSON(resp); err != nil {
					return
				}
			case "vsa":
				r := vsa.GetScreenshot("")
				if !r.Success {
					resp.OK = false
					resp.Message = r.ErrorMessage
					conn.WriteJSON(resp)
					return
				}
				resp.Image = r.Result["Screenshot"].String
				resp.OK = true
				resp.Message = "Success"
				if err := conn.WriteJSON(resp); err != nil {
					return
				}
			case "pm":
				r := pm.GetPowerChannelA(true)
				if !r.Success {
					resp.OK = false
					resp.Message = r.ErrorMessage
					conn.WriteJSON(resp)
					return
				}
				resp.PMChannelA = r.Result["Power"].Value
				r = pm.GetPowerChannelB(true)
				if !r.Success {
					resp.OK = false
					resp.Message = r.ErrorMessage
					conn.WriteJSON(resp)
					return
				}
				resp.PMChannelB = r.Result["Power"].Value
				resp.OK = true
				resp.Message = "Success"
				if err := conn.WriteJSON(resp); err != nil {
					return
				}
			case "ppm":
				r := ppm.GetPeakPower("A", true)
				if !r.Success {
					resp.OK = false
					resp.Message = r.ErrorMessage
					conn.WriteJSON(resp)
					return
				}
				resp.PPMChannelAPeakPower = r.Result["PulsePeakPower"].Value
				resp.PPMChannelAAvgPower = r.Result["PulseAveragePower"].Value
				r = ppm.GetPeakPower("B", true)
				if !r.Success {
					resp.OK = false
					resp.Message = r.ErrorMessage
					conn.WriteJSON(resp)
					return
				}
				resp.PPMChannelBPeakPower = r.Result["PulsePeakPower"].Value
				resp.PPMChannelBAvgPower = r.Result["PulseAveragePower"].Value
				resp.OK = true
				resp.Message = "Success"
				if err := conn.WriteJSON(resp); err != nil {
					return
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}
