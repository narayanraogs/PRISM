package driver

import (
	"prismServer/database"
	"prismServer/utils"
	"strings"
)

type GTX struct {
	device     gtxDevice
	deviceMake string
}

func (gtx *GTX) LoadDevice(name string) bool {
	if utils.Config.Simulator.GTx {
		gtx.deviceMake = "SimulatedGTx"
		gtx.device = &simulatedGTx{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	gtx.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("Cortex", gtx.deviceMake) {
		gtx.device = &cortex{}
		lan := gtx.device.loadLANDetails(name)
		cmds := gtx.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("TTCP", gtx.deviceMake) {
		gtx.device = &ttcp{}
		lan := gtx.device.loadLANDetails(name)
		cmds := gtx.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("DataPattern", gtx.deviceMake) {
		gtx.device = &dp{}
		lan := gtx.device.loadLANDetails(name)
		cmds := gtx.device.loadCommands()
		loaded = lan && cmds
	}
	return loaded
}

func (gtx *GTX) CheckConnection() utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.checkConnection()
}

func (gtx *GTX) SetCarrierOn(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setCarrierOn(component)
}

func (gtx *GTX) SetCarrierOff(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setCarrierOff(component)
}

func (gtx *GTX) SetModulationOn(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setModulationOn(component)
}

func (gtx *GTX) SetModulationOff(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setModulationOff(component)
}

func (gtx *GTX) SetFrequencyDeviationTC(component string, deviation float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setFrequencyDeviationTC(component, deviation)
}

func (gtx *GTX) SetFrequencyDeviationTone(component string, deviation float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setFrequencyDeviationTone(component, deviation)
}

func (gtx *GTX) SetModIndexTC(component string, modIndex float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setModIndexTC(component, modIndex)
}

func (gtx *GTX) SetModIndexTone(component string, modIndex float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setModIndexTone(component, modIndex)
}

func (gtx *GTX) SetRangingToneFrequency(frequency float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setRangingToneFrequency(frequency)
}

func (gtx *GTX) SetOnlyTC(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setOnlyTC(component)
}

func (gtx *GTX) SetOnlyRanging(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setOnlyRanging(component)
}

func (gtx *GTX) SetSweepRate(component string, sweepRate float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setSweepRate(component, sweepRate)
}

func (gtx *GTX) SetSweepRange(component string, sweepRange float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setSweepRange(component, sweepRange)
}

func (gtx *GTX) SetSweepStep(component string, sweepStep float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setSweepStep(component, sweepStep)
}

func (gtx *GTX) StartSweep(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.startSweep(component)
}

func (gtx *GTX) StopSweep(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.stopSweep(component)
}

func (gtx *GTX) TriggerSweep(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.triggerSweep(component)
}

func (gtx *GTX) SetPower(component string, power float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setPower(component, power)
}

func (gtx *GTX) SetFrequency(component string, frequency float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setFrequency(component, frequency)
}

func (gtx *GTX) SetChipRate(component string, chipRate float64) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setChipRate(component, chipRate)
}

func (gtx *GTX) SetIdlePatternOff() utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setIdlePatternOff()
}

func (gtx *GTX) SetIdlePatternOn() utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setIdlePatternOn()
}

func (gtx *GTX) SetDopplerCompensationEnable() utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setDopplerCompensationEnable()
}

func (gtx *GTX) SetDopplerCompensationDisable() utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setDopplerCompensationDisable()
}

func (gtx *GTX) EnableDoppler(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.enableDoppler(component)
}

func (gtx *GTX) GetDopplerCompensationStatus(component string) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.getDopplerCompensationStatus(component)
}

func (gtx *GTX) SetDopplerCompensationTable(timeOffset int, freqs []int, extFreqs []int, times []int) utils.CommandResponse {
	if gtx.device == nil {
		return getDeviceNotAvailable()
	}
	return gtx.device.setDopplerCompensationTable(timeOffset, freqs, extFreqs, times)
}
