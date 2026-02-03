package driver

import (
	_ "embed"
	"encoding/base64"
	"os"
	"prismServer/utils"
)

type simulatedSA struct {
	connection instrument
}

func (device *simulatedSA) setClearWrite() utils.CommandResponse {
	//TODO implement me
	panic("implement me")
}

func (device *simulatedSA) loadLANDetails(name string) bool {
	return true
}

func (device *simulatedSA) loadCommands() bool {
	return true
}

func (device *simulatedSA) initializeDevice(name string) {
}

func (device *simulatedSA) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	var cmds = make([]utils.Command, 0)
	return cmds
}

func (device *simulatedSA) communicate(cmds []utils.Command, port string) []string {
	return make([]string, 0)
}

func (device *simulatedSA) setSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setOBWSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getSpectrum() utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["CenterFrequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      2e9,
	}
	ret.Result["Span"] = utils.CommandResult{
		ResultType: "Value",
		Value:      1e6,
	}
	ret.Result["RBW"] = utils.CommandResult{
		ResultType: "Value",
		Value:      3e3,
	}
	ret.Result["VBW"] = utils.CommandResult{
		ResultType: "Value",
		Value:      1e3,
	}
	ret.Result["ReferenceLevel"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -10,
	}
	return ret
}

func (device *simulatedSA) setMarkerNormal(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMarkerDelta(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMarkerMaxPeak(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMarkerMinPeak(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getMarkerValue(markerNo int) utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["MarkerY"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	ret.Result["MarkerX"] = utils.CommandResult{
		ResultType: "Value",
		Value:      2e9,
	}
	return ret
}

func (device *simulatedSA) singleSweep() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) sweepOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) continuousSweep() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) restartSweep() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setFrequencyCounterMode(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getFrequencyInCounterMode(markerNo int) utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["Frequency"] = utils.CommandResult{
		ResultType: "Value",
		Value:      2e9,
	}
	return ret
}

func (device *simulatedSA) setNormalMode(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMaxHold() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMaxHoldOBW() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getSweepTime() utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["SweepTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      0.1,
	}
	return ret
}

func (device *simulatedSA) getSweepTimeOBW() utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["SweepTime"] = utils.CommandResult{
		ResultType: "Value",
		Value:      0.1,
	}
	return ret
}

func (device *simulatedSA) setOccupiedBandwidth(bwPercent float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getOccupiedBW() utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["Bandwidth"] = utils.CommandResult{
		ResultType: "Value",
		Value:      120e3,
	}
	ret.Result["Power"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -20,
	}
	return ret
}

func (device *simulatedSA) setClearScreen() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getOperationStatus() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getNoOfRowsToSkipInTrace() utils.CommandResponse {
	var ret = getSuccessResponse()
	ret.Result["NoOfRows"] = utils.CommandResult{
		ResultType: "Value",
		Value:      45.0,
	}
	return ret
}

func (device *simulatedSA) setAutoAlignOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setAutoAlignOn() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setAllMarkerOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) systemPreset() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setYScale(scale float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setPeakThresholdAndExcursion(threshold float64, excursion float64, markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) nextPeakLeft(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) nextPeakRight(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) shiftMarker(offset float64, markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMarkerNextPeak(markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setReferenceLevel(level float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getPeakFrequencyDeviationFM(firstMarker string, secondMarker string) utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["Frequency1"] = utils.CommandResult{
		ResultType: "Value",
		Value:      1.8e6,
	}
	ret.Result["Frequency2"] = utils.CommandResult{
		ResultType: "Value",
		Value:      2.2e6,
	}
	return ret
}

func (device *simulatedSA) setPhaseNoiseMode() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setSAMode() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setAutoTune() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setMarkerValuePhaseNoise(value float64, markerNo int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) setPhaseNoiseMeasurement(startOffset float64, stopOffset float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedSA) getPhaseNoiseMarkerY(marker int) utils.CommandResponse {
	ret := getSuccessResponse()
	ret.Result["MarkerY"] = utils.CommandResult{
		ResultType: "Value",
		Value:      -100,
	}
	return ret
}

func (device *simulatedSA) getSpectrumDump() utils.CommandResponse {
	//todo: replace with a proper image
	data := make([]byte, 100)
	filename := utils.GetTimeStampedFileName("screenshot")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".png"
	_ = os.WriteFile(filename, data, os.ModePerm)

	var encodedImage = base64.StdEncoding.EncodeToString(data)
	ret := getSuccessResponse()
	ret.Result["SpectrumDump"] = utils.CommandResult{
		ResultType: "Image",
		String:     encodedImage,
	}
	return ret
}

func (device *simulatedSA) getTraceDump(points int) utils.CommandResponse {
	//todo: replace with a proper trace
	data := make([]byte, 100)
	filename := utils.GetTimeStampedFileName("trace")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".csv"
	_ = os.WriteFile(filename, data, os.ModePerm)

	fileData := string(data)
	ret := getSuccessResponse()
	ret.Result["TraceDump"] = utils.CommandResult{
		ResultType: "String",
		String:     fileData,
	}
	return ret
}
