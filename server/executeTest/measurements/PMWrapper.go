package measurements

import (
	"prismServer/driver"
	"prismServer/utils"
)

func readPowerChannelA(pm driver.PM) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return pm.GetPowerChannelA(true)
	}
}

func readPowerChannelB(pm driver.PM) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return pm.GetPowerChannelB(true)
	}
}

func setFrequencyChA(pm driver.PM, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return pm.SetChannelA(frequency)
	}
}

func setFrequencyChB(pm driver.PM, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return pm.SetChannelB(frequency)
	}
}
