package measurements

import (
	"prismServer/driver"
	"prismServer/utils"
)

func setSGFrequency(sg driver.SG, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sg.SetFrequency(frequency)
	}
}

func setSGPower(sg driver.SG, power float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return sg.SetPower(power)
	}
}
