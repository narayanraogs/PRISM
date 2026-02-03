package driver

import (
	"prismServer/database"
	"prismServer/utils"
	"strconv"
	"strings"
)

type TSM struct {
	device     tsmDevice
	deviceMake string
}

func (tsm *TSM) LoadDevice(name string) bool {
	if utils.Config.Simulator.TSM {
		tsm.deviceMake = "SimulatedTSM"
		tsm.device = &simulatedTSM{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	tsm.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("TSM", tsm.deviceMake) {
		tsm.device = &indTSM{}
		lan := tsm.device.loadLANDetails(name)
		cmds := tsm.device.loadCommands()
		loaded = lan && cmds
	}
	return loaded
}

func (tsm *TSM) CheckConnection() utils.CommandResponse {
	if tsm.device == nil {
		return getDeviceNotAvailable()
	}
	return tsm.device.getDriverPath()
}

func (tsm *TSM) GetDriverPath() utils.CommandResponse {
	if tsm.device == nil {
		return getDeviceNotAvailable()
	}
	return tsm.device.getDriverPath()
}

func (tsm *TSM) SetDriverStatus(paths string) utils.CommandResponse {
	if tsm.device == nil {
		return getDeviceNotAvailable()
	}
	temp := strings.Split(paths, "!")
	for _, path := range temp {
		path = strings.ReplaceAll(path, "D", "")
		driverNoStr := path[:1]
		driverNo, _ := strconv.Atoi(driverNoStr)
		path = path[1:]
		onStatus := ""
		offStatus := ""
		if strings.Contains(path, "A") {
			bIndex := strings.Index(path, "B")
			tempPath := path
			if bIndex != -1 {
				tempPath = path[:bIndex]
				path = path[bIndex:]
			}
			onStatus = strings.ReplaceAll(tempPath, "A", "")
		}
		if strings.Contains(path, "B") {
			offStatus = strings.ReplaceAll(path, "B", "")
		}
		response := tsm.device.setDriverPath(driverNo, onStatus, offStatus)
		if !response.Success {
			return getErrorResponse("Unable to Communicate with TSM")
		}
	}
	return getSuccessResponse()
}

func (tsm *TSM) GetError() utils.CommandResponse {
	if tsm.device == nil {
		return getDeviceNotAvailable()
	}
	return tsm.device.getError()
}

func (tsm *TSM) GetAttn(attnNo int) utils.CommandResponse {
	if tsm.device == nil {
		return getDeviceNotAvailable()
	}
	return tsm.device.getAttn(attnNo)
}

func (tsm *TSM) SetAttn(attnNo int, value float64) utils.CommandResponse {
	if tsm.device == nil {
		return getDeviceNotAvailable()
	}
	return tsm.device.setAttn(value, attnNo)
}
