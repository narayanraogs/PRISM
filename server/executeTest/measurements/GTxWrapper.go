package measurements

import (
	"prismServer/driver"
	"prismServer/utils"
)

func setGTxIntermediateFrequency(gtx driver.GTX, component string, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetFrequency(component, frequency)
	}
}

func setGTxModulationOff(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetModulationOff(component)
	}
}

func setGTxModulationOn(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetModulationOn(component)
	}
}

func setGTxCarrierOff(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetCarrierOff(component)
	}
}

func setGTxCarrierOn(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetCarrierOn(component)
	}
}

func getGTxDopplerCompensation(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.GetDopplerCompensationStatus(component)
	}
}

func setGTxDopplerCompensationTable(gtx driver.GTX, timeOffset int, frequencies []int, extendedFrequencies []int, times []int) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetDopplerCompensationTable(timeOffset, frequencies, extendedFrequencies, times)
	}
}

func setGTxEnableDoppler(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.EnableDoppler(component)
	}
}

func setGTxDopplerCompensationEnable(gtx driver.GTX) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetDopplerCompensationEnable()
	}
}

func setGTxDopplerCompensationDisable(gtx driver.GTX) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetDopplerCompensationDisable()
	}
}

func setGTxOnlyTC(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetOnlyTC(component)
	}
}

func setGTxPower(gtx driver.GTX, component string, power float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetPower(component, power)
	}
}

func setGTxIdleOff(gtx driver.GTX) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetIdlePatternOff()
	}
}

func setGTxStopSweep(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.StopSweep(component)
	}
}

func setGTxStartSweep(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.StartSweep(component)
	}
}

func setGtxSweepRange(gtx driver.GTX, component string, sweepRange float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetSweepRange(component, sweepRange)
	}
}

func setGtxSweepRate(gtx driver.GTX, component string, sweepRate float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetSweepRate(component, sweepRate)
	}
}

func setGtxSweepStep(gtx driver.GTX, component string, sweepStep float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetSweepStep(component, sweepStep)
	}
}

func setGtxTriggerSweep(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.TriggerSweep(component)
	}
}

func setGTxModIndexTC(gtx driver.GTX, component string, modIndex float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetModIndexTC(component, modIndex)
	}
}

func setGTxModIndexTone(gtx driver.GTX, component string, toneMI float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetModIndexTone(component, toneMI)
	}
}

func setGTxFrequencyDeviationTC(gtx driver.GTX, component string, freqDev float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetFrequencyDeviationTC(component, freqDev)
	}
}

func setGTxChipRateDSSS(gtx driver.GTX, chipRate float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetChipRate("TCU", chipRate)
	}
}

func setOnlyRanging(gtx driver.GTX, component string) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetOnlyRanging(component)
	}
}

func setRangingToneFrequency(gtx driver.GTX, frequency float64) func() utils.CommandResponse {
	return func() utils.CommandResponse {
		return gtx.SetRangingToneFrequency(frequency)
	}
}
