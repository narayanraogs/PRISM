package driver

import (
	"prismServer/database"
	"prismServer/utils"
	"strings"
)

type PPM struct {
	device     ppmDevice
	deviceMake string
}

func (ppm *PPM) LoadDevice(name string) bool {
	if utils.Config.Simulator.PPM {
		ppm.deviceMake = "SimulatedPPM"
		ppm.device = &simulatedPM{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	ppm.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("NRX", ppm.deviceMake) {
		ppm.device = &nrx{}
		lan := ppm.device.loadLANDetails(name)
		cmds := ppm.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("ML2488B", ppm.deviceMake) {
		ppm.device = &ml2488b{}
		lan := ppm.device.loadLANDetails(name)
		cmds := ppm.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("N1912A", ppm.deviceMake) {
		ppm.device = &n1912a{}
		lan := ppm.device.loadLANDetails(name)
		cmds := ppm.device.loadCommands()
		loaded = lan && cmds
	}
	return loaded
}

func (ppm *PPM) CheckConnection() utils.CommandResponse {
	return ppm.SetChannelFrequency("A", 50000000)
}

func (ppm *PPM) Disconnect() utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.disConnect()
}

func (ppm *PPM) PresetPPM() utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.presetPPM()
}
func (ppm *PPM) SetChannelFrequency(channel string, frequency float64) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.setChannelFrequency(channel, frequency)
}
func (ppm *PPM) GetAveragePower(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getAveragePower(channel, connect)
}
func (ppm *PPM) GetPeakPower(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getPeakPower(channel, connect)
}

func (ppm *PPM) GetFrequency(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getFrequency(channel, connect)
}
func (ppm *PPM) SetPulseParameters(pulseWidth float64, pulsePeriod float64,
	triggerLevel float64, referenceLevel float64, yDiv float64, channel string, preset bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.setPulseParameters(pulseWidth, pulsePeriod,
		triggerLevel, referenceLevel, yDiv, channel, preset)
}
func (ppm *PPM) GetRiseTime(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getRiseTime(channel, connect)
}
func (ppm *PPM) GetFallTime(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getFallTime(channel, connect)
}
func (ppm *PPM) GetPulsePeriod(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getPulsePeriod(channel, connect)
}
func (ppm *PPM) GetPulseOffTime(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getPulseOffTime(channel, connect)
}
func (ppm *PPM) GetPulseWidth(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getPulseWidth(channel, connect)
}
func (ppm *PPM) GetDutyCycle(channel string, connect bool) utils.CommandResponse {
	if ppm.device == nil {
		return getDeviceNotAvailable()
	}
	return ppm.device.getDutyCycle(channel, connect)
}
