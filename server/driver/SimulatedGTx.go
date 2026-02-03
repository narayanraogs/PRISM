package driver

import (
	"prismServer/utils"
)

type simulatedGTx struct {
	connection instrument
}

func (device *simulatedGTx) loadLANDetails(name string) bool {
	return true
}

func (device *simulatedGTx) loadCommands() bool {
	return true
}

func (device *simulatedGTx) initializeDevice(name string) {
}

func (device *simulatedGTx) setCarrierOn(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setCarrierOff(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setModulationOn(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setModulationOff(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setFrequencyDeviationTC(component string, deviation float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setFrequencyDeviationTone(component string, deviation float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setModIndexTC(component string, modIndex float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setModIndexTone(component string, modIndex float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setRangingToneFrequency(frequency float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setOnlyTC(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setOnlyRanging(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setTCAndRanging(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setSweepRate(component string, sweepRate float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setSweepStep(component string, sweepStep float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) triggerSweep(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) sweepHold(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) sweepContinuous(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) enableDoppler(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) startSweep(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) stopSweep(component string) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setSweepRange(component string, sweepRange float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setFrequency(component string, frequency float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setPower(component string, power float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setChipRate(component string, chipRate float64) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) checkConnection() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setDopplerCompensationEnable() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setDopplerCompensationDisable() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) getDeviceTime() utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Time"] = utils.CommandResult{
		ResultType: "Integer",
		Integer:    100,
	}
	return response
}

func (device *simulatedGTx) getDopplerCompensationStatus(component string) utils.CommandResponse {
	response := getSuccessResponse()
	response.Result["Doppler"] = utils.CommandResult{
		ResultType: "Integer",
		Integer:    0,
	}
	return response
}

func (device *simulatedGTx) setIdlePatternOn() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setIdlePatternOff() utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) setDopplerCompensationTable(timeOffset int, frequencies []int, extendedFrequencies []int, times []int) utils.CommandResponse {
	return getSuccessResponse()
}

func (device *simulatedGTx) getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command {
	return make([]utils.Command, 0)
}

func (device *simulatedGTx) communicate(cmds []utils.Command, port string) []string {
	return make([]string, 0)
}
