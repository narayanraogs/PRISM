package utilities

import (
	"prismServer/driver"
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

func startPMStability(pmName string, info *pmStability, stab *StabilityPlot) {
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
			stab.addPoint(info.channelADescription, resp.Result["Power"].Value)
		}
		if info.channelBPresent {
			resp := pm.GetPowerChannelB(false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.channelBDescription, resp.Result["Power"].Value)
		}
	}
}
