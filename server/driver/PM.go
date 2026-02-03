package driver

import (
	"prismServer/database"
	"prismServer/utils"
	"strings"
)

type PM struct {
	device     pmDevice
	deviceMake string
}

func (pm *PM) LoadDevice(name string) bool {
	if utils.Config.Simulator.PM {
		pm.deviceMake = "SimulatedPM"
		pm.device = &simulatedPM{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	pm.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("NRX", pm.deviceMake) {
		pm.device = &nrx{}
		lan := pm.device.loadLANDetails(name)
		cmds := pm.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("ML2488B", pm.deviceMake) {
		pm.device = &ml2488b{}
		lan := pm.device.loadLANDetails(name)
		cmds := pm.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("N1912A", pm.deviceMake) {
		pm.device = &n1912a{}
		lan := pm.device.loadLANDetails(name)
		cmds := pm.device.loadCommands()
		loaded = lan && cmds
	}
	return loaded
}

func (pm *PM) CheckConnection() utils.CommandResponse {
	return pm.GetPowerChannelA(true)
}

func (pm *PM) SetChannelA(frequency float64) utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.setChannelA(frequency)
}

func (pm *PM) GetPowerChannelA(connect bool) utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.getPowerChannelA(connect)
}

func (pm *PM) SetChannelB(frequency float64) utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.setChannelB(frequency)
}

func (pm *PM) GetPowerChannelB(connect bool) utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.getPowerChannelB(connect)
}

func (pm *PM) SetChAAverageOff() utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.setChAAverageOff()
}

func (pm *PM) SetChBAverageOff() utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.setChBAverageOff()
}

func (pm *PM) SetChAAverageOn() utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.setChAAverageOn()
}

func (pm *PM) SetChBAverageOn() utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.setChBAverageOn()
}

func (pm *PM) Disconnect() utils.CommandResponse {
	if pm.device == nil {
		return getDeviceNotAvailable()
	}
	return pm.device.disConnect()
}
