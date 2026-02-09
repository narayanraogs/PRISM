package driver

import (
	"prismServer/database"
	"prismServer/utils"
	"strings"
)

type SG struct {
	device     sgDevice
	deviceMake string
}

func (sg *SG) LoadDevice(name string) bool {
	if utils.Config.Simulator.SG {
		sg.deviceMake = "SimulatedSG"
		sg.device = &simulatedSG{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	sg.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("E8257D", sg.deviceMake) {
		sg.device = &e82x7d{}
		lan := sg.device.loadLANDetails(name)
		cmds := sg.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("E8267D", sg.deviceMake) {
		sg.device = &e82x7d{}
		lan := sg.device.loadLANDetails(name)
		cmds := sg.device.loadCommands()
		loaded = lan && cmds
	}
	return loaded
}

func (sg *SG) CheckConnection() utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.getModulationStatus()
}

func (sg *SG) GetCommandDatabase() map[string]utils.Command {
	if sg.device == nil {
		return nil
	}
	return sg.device.getCommandDatabase()
}

func (sg *SG) SetRFOn() utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.setRFOn()
}

func (sg *SG) SetRFOff() utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.setRFOff()
}

func (sg *SG) SetModOn() utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.setModOn()
}

func (sg *SG) SetModOff() utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.setModOff()
}

func (sg *SG) SetPower(power float64) utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.setPower(power)
}

func (sg *SG) SetFrequency(frequency float64) utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.setFrequency(frequency)
}

func (sg *SG) GetModStatus() utils.CommandResponse {
	if sg.device == nil {
		return getDeviceNotAvailable()
	}
	return sg.device.getModulationStatus()
}
