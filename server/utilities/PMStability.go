package utilities

import (
	"prismServer/driver"
	"prismServer/logger"
	"time"
)

type pmStability struct {
	channelAPresent     bool
	channelAFrequency   float64
	channelADescription string
	channelBPresent     bool
	channelBFrequency   float64
	channelBDescription string
	stop                bool
}

func startPMStability(id int64, pmName string, info *pmStability, dataChannel chan StabilityUpdate) {
	logger.Log.Info("Starting PM Stability Measurement", "stabilityId", id, "pmName", pmName)
	var pm driver.PM
	ok := pm.LoadDevice(pmName)
	if !ok {
		logger.Log.Error("Failed to load PM device for stability", "stabilityId", id, "pmName", pmName)
		return
	}
	if info.channelAPresent {
		pm.SetChannelA(info.channelAFrequency)
	}
	if info.channelBPresent {
		pm.SetChannelB(info.channelAFrequency)
	}
	pointsMeasured := 0
	for !info.stop {
		if info.channelAPresent {
			resp := pm.GetPowerChannelA(false)
			if !resp.Success {
				logger.Log.Error("PM Stability Measurement Error", "stabilityId", id, "channel", "A", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.channelADescription,
				Value:       resp.Result["Power"].Value,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}
		if info.channelBPresent {
			resp := pm.GetPowerChannelB(false)
			if !resp.Success {
				logger.Log.Error("PM Stability Measurement Error", "stabilityId", id, "channel", "B", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.channelBDescription,
				Value:       resp.Result["Power"].Value,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if pointsMeasured > 0 && pointsMeasured%1000 == 0 {
			logger.Log.Info("PM Stability Heartbeat", "stabilityId", id, "pointsMeasured", pointsMeasured)
		}
	}
	logger.Log.Info("Completed PM Stability Measurement", "stabilityId", id, "totalPointsMeasured", pointsMeasured)
}
