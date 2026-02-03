package measurements

import (
	"prismServer/driver"
	"prismServer/utils"
)

func setChannelFrequency(ppm driver.PPM, channel string, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.SetChannelFrequency(channel, frequency)
	}
}
func setPulseParameters(ppm driver.PPM, pulseWidth float64, pulsePeriod float64,
	triggerLevel float64, referenceLevel float64, yDiv float64, channel string, preset bool) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.SetPulseParameters(pulseWidth, pulsePeriod, triggerLevel, referenceLevel, yDiv, channel, preset)
	}
}

func getPeakPower(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetPeakPower(channel, true)
	}
}

func getPulseWidth(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetPulseWidth(channel, true)
	}
}

func getPulsePeriod(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetPulsePeriod(channel, true)
	}
}

func getPulseOffTime(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetPulseOffTime(channel, true)
	}
}

func getRiseTime(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetRiseTime(channel, true)
	}
}

func getFallTime(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetFallTime(channel, true)
	}
}

func getDutyCycle(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetDutyCycle(channel, true)
	}
}

func getAveragePower(ppm driver.PPM, channel string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return ppm.GetAveragePower(channel, true)
	}
}
