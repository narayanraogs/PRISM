package driver

import (
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/utils"
	"strconv"
	"strings"
)

type VSA struct {
	SA
	device     vsaDevice
	deviceMake string
}

func (vsa *VSA) LoadDevice(name string) bool {
	if utils.Config.Simulator.VSA {
		vsa.deviceMake = "SimulatedVSA"
		vsa.device = &simulatedVSA{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	vsa.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("N9040B", vsa.deviceMake) {
		vsa.device = &n9040b{}
		lan := vsa.device.loadLANDetails(name)
		cmds := vsa.device.loadCommands()
		vsa.SA.device = &n9040b{}
		lanSA := vsa.SA.device.loadLANDetails(name)
		cmdsSA := vsa.SA.device.loadCommands()
		loaded = lan && cmds && lanSA && cmdsSA
	}
	return loaded
}

func (vsa *VSA) GetAssociatedSA() SA {
	return vsa.SA
}

func (vsa *VSA) SetPulseMode() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.setPulseMode()
}

func (vsa *VSA) SetSpectrumParameters(centerFrequency float64, span float64, rbw float64) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.setSpectrumParameters(centerFrequency, span, rbw)
}

func (vsa *VSA) SetPulseParameters(acquisitionTime float64, YTop float64,
	pdiv float64, analLength float64, reflevel float64, hystlevel float64, points int32, bufferlength int32) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.setPulseParameters(acquisitionTime, YTop, pdiv, analLength, reflevel, hystlevel, points, bufferlength)
}

func (vsa *VSA) StartMeasurement() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.startMeasurement()
}

func (vsa *VSA) StopMeasurement() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.stopMeasurement()
}

func (vsa *VSA) GetPulseParamaeters() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.getPulseParameters()
}

func (vsa *VSA) GetScreenshot(mode string) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.getScreenshot(mode)
}

func (vsa *VSA) SetSpectrogramMode(mode string) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.setSpectrogramMode(mode)
}

func (vsa *VSA) GetPulseAveragePower() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.getPulseAveragePower()
}

func (vsa *VSA) WaitTillFirstPulse() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.waitTillFirstPulse()
}

func (vsa *VSA) SetSelectedTrace(number int) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.setSelectedTrace(number)
}

func (vsa *VSA) RestoreTrace() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.restoreTrace()
}

func (vsa *VSA) GetTraceDumpSA(points int) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.getTraceDumpSA(points)
}

func (vsa *VSA) StartVSA() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.startVSA()
}

func (vsa *VSA) StartSA() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.startSA()
}

func (vsa *VSA) CheckSA() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.checkSA()
}

func (vsa *VSA) CheckVSA() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.checkVSA()
}

func (vsa *VSA) CheckConnection() utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	return vsa.device.getOperationStatus()
}

func (vsa *VSA) MeasureBW(centerFreq float64, firstMarker string, secondMarker string) utils.CommandResponse {
	firstMarkerInt, _ := strconv.Atoi(firstMarker)
	secondMarkerInt, _ := strconv.Atoi(secondMarker)
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := vsa.SA.device.setMaxHold()
	if !resp.Success {
		return resp
	}
	vsa.SA.WaitForSweeps(5)
	resp = vsa.device.sweepOff()
	if !resp.Success {
		return resp
	}
	resp = vsa.SA.device.getPeakFrequencyDeviationFM(firstMarker, secondMarker)
	if !resp.Success {
		return resp
	}
	firstPeak := resp.Result["Frequency1"].Value
	if firstPeak < centerFreq {
		resp = vsa.SA.device.nextPeakLeft(firstMarkerInt)
		if !resp.Success {
			return resp
		}
		resp = vsa.SA.device.nextPeakRight(secondMarkerInt)
		if !resp.Success {
			return resp
		}
	} else {
		resp = vsa.SA.device.nextPeakRight(firstMarkerInt)
		if !resp.Success {
			return resp
		}
		resp = vsa.SA.device.nextPeakLeft(secondMarkerInt)
		if !resp.Success {
			return resp
		}
	}
	resp = vsa.SA.device.getMarkerValue(firstMarkerInt)
	if !resp.Success {
		return resp
	}
	freq1 := resp.Result["MarkerX"].Value
	vsa.SA.WaitForSweeps(1)
	resp = vsa.SA.device.getMarkerValue(secondMarkerInt)
	if !resp.Success {
		return resp
	}
	freq2 := resp.Result["MarkerX"].Value

	resp = vsa.SA.device.setClearWrite()
	if !resp.Success {
		return resp
	}
	resp = vsa.SA.device.continuousSweep()
	if !resp.Success {
		return resp
	}
	vsa.SA.WaitForSweeps(1)
	bandwidth := math.Abs(freq1 - freq2)
	fmt.Println("Bandwidth is...", bandwidth)

	ret := getSuccessResponse()
	ret.Result["Bandwidth"] = utils.CommandResult{
		ResultType: "Value",
		Value:      bandwidth,
	}
	return ret
}

func (vsa *VSA) GetBandwidthMeasurement(peakFreq float64, centreFreq float64) utils.CommandResponse {
	if vsa.device == nil {
		return getDeviceNotAvailable()
	}
	var bandwidth float64
	ppm := 2 * (centreFreq / 1e6)
	upper := centreFreq + ppm
	lower := centreFreq - ppm
	if (peakFreq > lower) && (peakFreq < upper) {
		resp := vsa.MeasureBW(centreFreq, "2", "3")
		if !resp.Success {
			return resp
		}
		bandwidth = resp.Result["Bandwidth"].Value
	} else {
		resp := vsa.MeasureBW(centreFreq, "1", "2")
		if !resp.Success {
			return resp
		}
		bandwidth = resp.Result["Bandwidth"].Value
	}

	vsa.SetAllMarkerOff()
	ret := getSuccessResponse()
	ret.Result["Bandwidth"] = utils.CommandResult{
		ResultType: "Value",
		Value:      bandwidth,
	}
	return ret
}
