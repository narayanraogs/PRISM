package utilities

import (
	"prismServer/driver"
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

func startPMStability(pmName string, info *pmStability, dataChannel chan StabilityUpdate) {
	var pm driver.PM
	ok := pm.LoadDevice(pmName)
	if !ok {
		return
	}
	if info.channelAPresent {
		pm.SetChannelA(info.channelAFrequency)
	}
	if info.channelBPresent {
		pm.SetChannelB(info.channelAFrequency)
	}
	for !info.stop {
		if info.channelAPresent {
			resp := pm.GetPowerChannelA(false)
			if !resp.Success {
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.channelADescription,
				Value:       resp.Result["Power"].Value,
				Timestamp:   time.Now(),
			}
		}
		if info.channelBPresent {
			resp := pm.GetPowerChannelB(false)
			if !resp.Success {
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.channelBDescription,
				Value:       resp.Result["Power"].Value,
				Timestamp:   time.Now(),
			}
		}
	}
}
