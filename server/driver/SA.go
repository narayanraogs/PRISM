package driver

import (
	"math"
	"prismServer/database"
	"prismServer/utils"
	"strings"
	"time"
)

type SA struct {
	device     saDevice
	deviceMake string
}

func (sa *SA) LoadDevice(name string) bool {
	if utils.Config.Simulator.SA {
		sa.deviceMake = "SimulatedSA"
		sa.device = &simulatedSA{}
		return true
	}
	dev, ok := database.GetDeviceDetails(name)
	if !ok {
		return false
	}
	sa.deviceMake = dev.DeviceMake
	var loaded = false
	if strings.EqualFold("N9040B", sa.deviceMake) {
		sa.device = &n9040b{}
		lan := sa.device.loadLANDetails(name)
		cmds := sa.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("N9030B", sa.deviceMake) {
		sa.device = &n9030b{}
		lan := sa.device.loadLANDetails(name)
		cmds := sa.device.loadCommands()
		loaded = lan && cmds
	}
	if strings.EqualFold("N9030A", sa.deviceMake) {
		sa.device = &n9030a{}
		lan := sa.device.loadLANDetails(name)
		cmds := sa.device.loadCommands()
		loaded = lan && cmds
	}
	return loaded
}

func (sa *SA) GetCommandDatabase() map[string]utils.Command {
	if sa.device == nil {
		return nil
	}
	return sa.device.getCommandDatabase()
}

func (sa *SA) CheckConnection() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getOperationStatus()
}

func (sa *SA) WaitForSweeps(number int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.device.getSweepTime()
	if !resp.Success {
		return resp
	}
	sweepTime := resp.Result["SweepTime"].Value
	sleepTime := sweepTime * float64(number)
	if sleepTime < 0.1 {
		sleepTime = 0.1
	}
	sleepTime = sleepTime * 1000
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	return getSuccessResponse()
}

func (sa *SA) SetSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setSpectrum(centerFrequency, span, rbw, vbw)
}

func (sa *SA) SetOBWSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setOBWSpectrum(centerFrequency, span, rbw, vbw)

}

func (sa *SA) GetSpectrum() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getSpectrum()

}

func (sa *SA) PeakSearch(maxHold bool, markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	if maxHold {
		resp := sa.SetMaxHold()
		if !resp.Success {
			return resp
		}
	} else {
		resp := sa.SetNormalMode()
		if !resp.Success {
			return resp
		}
	}
	resp := sa.WaitForSweeps(2)
	if !resp.Success {
		return resp
	}

	return sa.device.setMarkerMaxPeak(markerNo)
}

func (sa *SA) SetNormalMode() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setNormalMode(1)
}

func (sa *SA) SetMaxHold() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setMaxHold()
}

func (sa *SA) SetMaxHoldOBW() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setMaxHoldOBW()

}

func (sa *SA) GetFrequencyInCounterMode(markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.SetNormalMode()
	if !resp.Success {
		return resp
	}
	resp = sa.device.setFrequencyCounterMode(markerNo)
	if !resp.Success {
		return resp
	}
	time.Sleep(2 * time.Second)
	return sa.device.getFrequencyInCounterMode(markerNo)
}

func (sa *SA) GetMaxMarkerValue() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.device.setMarkerMaxPeak(1)
	if !resp.Success {
		return resp
	}
	resp = sa.WaitForSweeps(2)
	if !resp.Success {
		return resp
	}
	resp = sa.device.getMarkerValue(1)
	return resp
}

func (sa *SA) SetMarkerNextPeak(markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setMarkerNextPeak(markerNo)
}

func (sa *SA) SetThresholdAndPeakExcursion(threshold float64, excursion float64, markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setPeakThresholdAndExcursion(threshold, excursion, markerNo)
}

func (sa *SA) GetMarkerValue(markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getMarkerValue(markerNo)
}

func (sa *SA) GetPhaseNoiseMarkerY(markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getPhaseNoiseMarkerY(markerNo)
}

func (sa *SA) SetAllMarkerOff() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setAllMarkerOff()
}

func (sa *SA) GetMaxMinPeak() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	var response = getSuccessResponse()
	resp := sa.SetAllMarkerOff()
	if !resp.Success {
		return resp
	}
	resp = sa.SetMaxHold()
	if !resp.Success {
		return resp
	}
	resp = sa.WaitForSweeps(2)
	if !resp.Success {
		return resp
	}

	resp = sa.device.setMarkerMinPeak(1)
	if !resp.Success {
		return resp
	}
	resp = sa.WaitForSweeps(2)
	if !resp.Success {
		return resp
	}
	resp = sa.device.getMarkerValue(1)
	if !resp.Success {
		return resp
	}
	response.Result["MinValue"] = resp.Result["MarkerY"]
	resp = sa.device.setMarkerMaxPeak(1)
	if !resp.Success {
		return resp
	}
	resp = sa.WaitForSweeps(2)
	if !resp.Success {
		return resp
	}
	resp = sa.device.getMarkerValue(1)
	if !resp.Success {
		return resp
	}
	response.Result["MaxValue"] = resp.Result["MarkerY"]
	resp = sa.device.setNormalMode(1)
	if !resp.Success {
		return resp
	}
	return response
}

func (sa *SA) SetReferenceLevel(refLevel float64) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setReferenceLevel(refLevel)
}

func (sa *SA) SetReferenceNominal() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.CheckIfCarrierIsPresent()
	if !resp.Success {
		return resp
	}
	maxValue := resp.Result["MaxValue"].Value
	minValue := resp.Result["MinValue"].Value
	carrierPresent := resp.Result["Carrier"].Bool
	if !carrierPresent {
		return getErrorResponse("No Carrier Available")
	}
	scale := 10.0
	if maxValue-minValue > 80 {
		scale = 15.0
	}
	refLevel := maxValue + 10

	resp = sa.device.setYScale(scale)
	if !resp.Success {
		return resp
	}
	resp = sa.SetReferenceLevel(refLevel)
	if !resp.Success {
		return resp
	}
	resp = getSuccessResponse()
	resp.Result["ReferenceLevel"] = utils.CommandResult{
		ResultType: "Value",
		Value:      refLevel,
	}
	resp.Result["MaxValue"] = utils.CommandResult{
		ResultType: "Value",
		Value:      maxValue,
	}
	resp.Result["MinValue"] = utils.CommandResult{
		ResultType: "Value",
		Value:      minValue,
	}
	return resp
}

func (sa *SA) SystemPreset() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}

	return sa.device.systemPreset()
}

func (sa *SA) SingleSweep() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.singleSweep()
}

func (sa *SA) ContinuousSweep() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.continuousSweep()
}

func (sa *SA) RestartSweep() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.restartSweep()
}

func (sa *SA) GetSpectrumDump() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getSpectrumDump()
}

func (sa *SA) GetTraceDump(points int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getTraceDump(points)
}

func (sa *SA) CheckIfCarrierIsPresent() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.GetMaxMinPeak()
	if !resp.Success {
		return resp
	}
	maxValue := resp.Result["MaxValue"].Value
	minValue := resp.Result["MinValue"].Value
	if maxValue-minValue <= 10 {
		resp.Result["Carrier"] = utils.CommandResult{
			ResultType: "Boolean",
			Bool:       false,
		}
	} else {
		resp.Result["Carrier"] = utils.CommandResult{
			ResultType: "Boolean",
			Bool:       true,
		}
	}
	return resp
}

func (sa *SA) SetPeakThresholdAndExcursion(excursion float64, markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.SetMaxHold()
	if !resp.Success {
		return resp
	}
	resp = sa.GetMaxMinPeak()
	if !resp.Success {
		return resp
	}
	minValue := resp.Result["MinValue"].Value

	return sa.device.setPeakThresholdAndExcursion(minValue+10, excursion, markerNo)
}

func (sa *SA) ShiftMarker(offset float64, markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.shiftMarker(offset, markerNo)
}

func (sa *SA) SetMarkerValuePhaseNoise(value float64, markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setMarkerValuePhaseNoise(value, markerNo)
}

func (sa *SA) GetModIndex(offsetFrequency float64) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.PeakSearch(true, 1)
	if !resp.Success {
		return resp
	}
	resp = sa.GetMarkerValue(1)
	if !resp.Success {
		return resp
	}
	sa.WaitForSweeps(2)
	maxPeak := resp.Result["MarkerY"].Value
	peakFreq := resp.Result["MarkerX"].Value
	resp = sa.device.shiftMarker(peakFreq-offsetFrequency, 1)
	if !resp.Success {
		return resp
	}
	resp = sa.GetMarkerValue(1)
	if !resp.Success {
		return resp
	}
	peakLeft := resp.Result["MarkerY"].Value
	resp = sa.PeakSearch(true, 1)
	if !resp.Success {
		return resp
	}
	sa.WaitForSweeps(2)
	resp = sa.device.shiftMarker(peakFreq+offsetFrequency, 1)
	if !resp.Success {
		return resp
	}
	resp = sa.GetMarkerValue(1)
	if !resp.Success {
		return resp
	}
	peakRight := resp.Result["MarkerY"].Value
	modIndexForLeft, _ := utils.PowerToModIndexConverter(maxPeak - peakLeft)
	modIndexForRight, _ := utils.PowerToModIndexConverter(maxPeak - peakRight)
	var response = getSuccessResponse()
	response.Result["modIndexForLeft"] = utils.CommandResult{
		ResultType: "Value",
		Value:      modIndexForLeft,
	}
	response.Result["modIndexForRight"] = utils.CommandResult{
		ResultType: "Value",
		Value:      modIndexForRight,
	}
	return response

}

func (sa *SA) SetAlignmentOff() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setAutoAlignOff()
}

func (sa *SA) SetAlignmentOn() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setAutoAlignOn()
}

func (sa *SA) GetNoOfRowsToSkipInTrace() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getNoOfRowsToSkipInTrace()
}

func (sa *SA) GetFrequencyDeviationFM(frequency float64) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.SetMaxHold()
	if !resp.Success {
		return resp
	}
	resp = sa.WaitForSweeps(5)
	if !resp.Success {
		return resp
	}
	resp = sa.SingleSweep()
	if !resp.Success {
		return resp
	}
	resp = sa.WaitForSweeps(2)
	if !resp.Success {
		return resp
	}

	resp = sa.device.getPeakFrequencyDeviationFM("1", "2")
	if !resp.Success {
		return resp
	}
	firstPeak := resp.Result["Frequency1"].Value
	leftMarker := 1
	rightMarker := 2
	if firstPeak > frequency {
		leftMarker = 2
		rightMarker = 1
	}
	resp = sa.device.nextPeakLeft(leftMarker)
	if !resp.Success {
		return resp
	}
	resp = sa.device.nextPeakRight(rightMarker)
	if !resp.Success {
		return resp
	}
	resp = sa.GetMarkerValue(1)
	if !resp.Success {
		return resp
	}
	frequency1 := resp.Result["MarkerX"].Value
	resp = sa.GetMarkerValue(2)
	if !resp.Success {
		return resp
	}
	frequency2 := resp.Result["MarkerX"].Value
	resp = sa.SetNormalMode()
	if !resp.Success {
		return resp
	}

	resp = sa.ContinuousSweep()
	if !resp.Success {
		return resp
	}
	var response = getSuccessResponse()
	response.Result["FrequencyDeviation"] = utils.CommandResult{
		ResultType: "Value",
		Value:      math.Abs(frequency1 - frequency2),
	}
	return response
}

func (sa *SA) SetPhaseNoiseMeasurement() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.device.setPhaseNoiseMode()
	if !resp.Success {
		return resp
	}
	time.Sleep(2 * time.Second)
	if !resp.Success {
		return resp
	}
	time.Sleep(2 * time.Second)
	resp = sa.device.setAutoTune()
	return sa.device.setPhaseNoiseMeasurement(100, 1000000)

}

func (sa *SA) SetSAMode() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setSAMode()
}

func (sa *SA) SetOccupiedBW(bwPercent float64) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.setOccupiedBandwidth(bwPercent)
}

func (sa *SA) GetOccupiedBW() utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	return sa.device.getOccupiedBW()
}

func (sa *SA) GetAllPeaksAbove(excursion float64, markerNo int) utils.CommandResponse {
	if sa.device == nil {
		return getDeviceNotAvailable()
	}
	resp := sa.device.setMaxHold()
	if !resp.Success {
		return resp
	}
	resp = sa.CheckIfCarrierIsPresent()
	if !resp.Success {
		return resp
	}
	//maxPeak := resp.Result["MaxValue"].Value
	minPeak := resp.Result["MinValue"].Value

	resp = sa.device.setPeakThresholdAndExcursion(minPeak, excursion, markerNo)
	if !resp.Success {
		return resp
	}
	peaks := make([]float64, 0)
	frequencies := make([]float64, 0)
	resp = sa.PeakSearch(false, markerNo)
	if !resp.Success {
		return resp
	}
	resp = sa.GetMarkerValue(markerNo)
	if !resp.Success {
		return resp
	}
	peaks = append(peaks, resp.Result["MarkerY"].Value)
	frequencies = append(frequencies, resp.Result["MarkerX"].Value)
	repeat := true
	for repeat == true {
		resp = sa.SetMarkerNextPeak(markerNo)
		if !resp.Success {
			return resp
		}
		if resp.Result["MarkerX"].Value == frequencies[len(frequencies)-1] {
			repeat = false
			continue
		}
		peaks = append(peaks, resp.Result["MarkerY"].Value)
		frequencies = append(frequencies, resp.Result["MarkerX"].Value)
	}

	ret := getSuccessResponse()
	ret.Result["Frequencies"] = utils.CommandResult{
		ResultType: "Values",
		Values:     frequencies,
	}
	ret.Result["Peaks"] = utils.CommandResult{
		ResultType: "Values",
		Values:     peaks,
	}
	ret.Result["MinValue"] = utils.CommandResult{
		ResultType: "Value",
		Value:      minPeak,
	}
	return ret
}
