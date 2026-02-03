package measurements

import (
	"prismServer/driver"
	"prismServer/utils"
)

func setSpectrumParameters(vsa driver.VSA, freq float64, span float64, rbw float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return vsa.SetSpectrumParameters(freq, span, rbw)
	}
}

func setPulseParametersForVSA(vsa driver.VSA, analysistime float64, YTop float64,
	pdiv float64, analLength float64, reflevel float64, hystlevel float64, points int32, bufferlength int32) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return vsa.SetPulseParameters(analysistime, YTop, pdiv, analLength, reflevel, hystlevel, points, bufferlength)
	}
}

func getScreenshot(vsa driver.VSA, mode string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return vsa.GetScreenshot(mode)
	}
}

func setSpectrogramMode(vsa driver.VSA, mode string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return vsa.SetSpectrogramMode(mode)
	}
}

func getPulseBandwidth(vsa driver.VSA, peakFreq float64, centreFreq float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return vsa.GetBandwidthMeasurement(peakFreq, centreFreq)
	}
}
