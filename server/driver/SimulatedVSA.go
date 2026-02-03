package driver

import (
	"encoding/base64"
	"os"
	"prismServer/utils"
)

type simulatedVSA struct {
	simulatedSA
}

func (device *simulatedVSA) setClearWrite() utils.CommandResponse {
	//TODO implement me
	panic("implement me")
}

func (device *simulatedVSA) getPeakFrequencyDeviationFM(firstMarker string, secondMarker string) utils.CommandResponse {
	//TODO implement me
	panic("implement me")
}

func (device *simulatedVSA) waitForSweeps(i int) utils.CommandResponse {
	//TODO implement me
	panic("implement me")
}

func (device *simulatedVSA) setPulseMode() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) setSpectrumParameters(centerFrequency float64, span float64, rbw float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) setPulseParameters(acqtime float64, YTop float64, pdiv float64, analLength float64, reflevel float64,
	hystlevel float64, points int32, bufferlength int32) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) startMeasurement() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) stopMeasurement() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) setSpectrogramMode(mode string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) getPulseParameters() utils.CommandResponse {
	//todo: write a pulse file to the below path
	//err := os.WriteFile(utils.Config.BaseFolder+"/temp/temp.csv", []byte(fileData.String()), os.ModePerm)
	return getSuccessResponse()
}

func (device *simulatedVSA) getPulseAveragePower() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) waitTillFirstPulse() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) getScreenshot(mode string) utils.CommandResponse {
	//todo: use an actual screenshot
	data := make([]byte, 100)

	filename := utils.GetTimeStampedFileName("screenshot")
	filename = utils.Config.BaseFolder + "/screenshots/" + filename + ".png"
	_ = os.WriteFile(filename, data, os.ModePerm)

	var encodedImage = base64.StdEncoding.EncodeToString(data)
	ret := getSuccessResponse()
	ret.Result["Screenshot"] = utils.CommandResult{
		ResultType: "Image",
		String:     encodedImage,
	}
	return ret
}
func (device *simulatedVSA) setSelectedTrace(number int) utils.CommandResponse {
	return getSuccessResponse()
}
func (device *simulatedVSA) restoreTrace() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) getSpectrumDumpSA() utils.CommandResponse {
	return device.simulatedSA.getSpectrumDump()
}

func (device *simulatedVSA) getTraceDumpSA(points int) utils.CommandResponse {
	return device.simulatedSA.getTraceDump(points)
}
func (device *simulatedVSA) startVSA() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) startSA() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) checkSA() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedVSA) checkVSA() utils.CommandResponse {
	return getSuccessResponse()
}
