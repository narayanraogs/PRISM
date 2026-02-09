package driver

import "prismServer/utils"

type saDevice interface {
	loadLANDetails(name string) bool
	loadCommands() bool
	initializeDevice(name string)
	getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command
	communicate(cmds []utils.Command, port string) []string
	setSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse
	setOBWSpectrum(centerFrequency float64, span float64, rbw float64, vbw float64) utils.CommandResponse
	getSpectrum() utils.CommandResponse
	setMarkerNormal(markerNo int) utils.CommandResponse
	setMarkerDelta(markerNo int) utils.CommandResponse
	setMarkerMaxPeak(markerNo int) utils.CommandResponse
	setMarkerMinPeak(markerNo int) utils.CommandResponse
	getMarkerValue(markerNo int) utils.CommandResponse
	singleSweep() utils.CommandResponse
	sweepOff() utils.CommandResponse
	continuousSweep() utils.CommandResponse
	restartSweep() utils.CommandResponse
	setFrequencyCounterMode(markerNo int) utils.CommandResponse
	getFrequencyInCounterMode(markerNo int) utils.CommandResponse
	setNormalMode(markerNo int) utils.CommandResponse
	setMaxHold() utils.CommandResponse
	setClearWrite() utils.CommandResponse
	setMaxHoldOBW() utils.CommandResponse
	getSweepTime() utils.CommandResponse
	getSweepTimeOBW() utils.CommandResponse
	setOccupiedBandwidth(bwPercent float64) utils.CommandResponse
	getOccupiedBW() utils.CommandResponse
	setClearScreen() utils.CommandResponse
	getOperationStatus() utils.CommandResponse
	getNoOfRowsToSkipInTrace() utils.CommandResponse
	setAutoAlignOff() utils.CommandResponse
	setAutoAlignOn() utils.CommandResponse
	setAllMarkerOff() utils.CommandResponse
	systemPreset() utils.CommandResponse
	setYScale(scale float64) utils.CommandResponse
	setPeakThresholdAndExcursion(threshold float64, excursion float64, markerNo int) utils.CommandResponse
	nextPeakLeft(markerNo int) utils.CommandResponse
	nextPeakRight(markerNo int) utils.CommandResponse
	shiftMarker(offset float64, markerNo int) utils.CommandResponse
	setMarkerNextPeak(markerNo int) utils.CommandResponse
	setReferenceLevel(level float64) utils.CommandResponse
	getPeakFrequencyDeviationFM(firstMarker string, secondMarker string) utils.CommandResponse
	setPhaseNoiseMode() utils.CommandResponse
	setSAMode() utils.CommandResponse
	setAutoTune() utils.CommandResponse
	setMarkerValuePhaseNoise(value float64, markerNo int) utils.CommandResponse
	setPhaseNoiseMeasurement(startOffset float64, stopOffset float64) utils.CommandResponse
	getPhaseNoiseMarkerY(marker int) utils.CommandResponse
	getSpectrumDump() utils.CommandResponse
	getTraceDump(points int) utils.CommandResponse
	getCommandDatabase() map[string]utils.Command
}

type pmDevice interface {
	loadLANDetails(name string) bool
	loadCommands() bool
	initializeDevice(name string)
	getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command
	communicate(cmds []utils.Command, port string, connect bool) []string
	setChannelA(frequency float64) utils.CommandResponse
	getPowerChannelA(connect bool) utils.CommandResponse
	setChannelB(frequency float64) utils.CommandResponse
	getPowerChannelB(connect bool) utils.CommandResponse
	setChAAverageOff() utils.CommandResponse
	setChAAverageOn() utils.CommandResponse
	setChBAverageOff() utils.CommandResponse
	setChBAverageOn() utils.CommandResponse
	presetPM() utils.CommandResponse
	disConnect() utils.CommandResponse
	getCommandDatabase() map[string]utils.Command
}

type ppmDevice interface {
	loadLANDetails(name string) bool
	loadCommands() bool
	initializeDevice(name string)
	getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command
	communicate(cmds []utils.Command, port string, connect bool) []string
	presetPPM() utils.CommandResponse
	setChannelFrequency(channel string, frequency float64) utils.CommandResponse
	getAveragePower(channel string, connect bool) utils.CommandResponse
	getPeakPower(channel string, connect bool) utils.CommandResponse
	setPulseParameters(pulseWidth float64, pulsePeriod float64,
		triggerLevel float64, referenceLevel float64, yDiv float64, channel string, preset bool) utils.CommandResponse
	getRiseTime(channel string, connect bool) utils.CommandResponse
	getFallTime(channel string, connect bool) utils.CommandResponse
	getPulsePeriod(channel string, connect bool) utils.CommandResponse
	getPulseOffTime(channel string, connect bool) utils.CommandResponse
	getPulseWidth(channel string, connect bool) utils.CommandResponse
	getDutyCycle(channel string, connect bool) utils.CommandResponse
	getFrequency(channel string, connect bool) utils.CommandResponse
	disConnect() utils.CommandResponse
}
type tsmDevice interface {
	loadLANDetails(name string) bool
	loadCommands() bool
	initializeDevice(name string)
	getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command
	communicate(cmds []utils.Command, port string) []string
	getDriverPath() utils.CommandResponse
	setDriverPath(driverNo int, onStatus string, offStatus string) utils.CommandResponse
	getError() utils.CommandResponse
	setAttn(value float64, attnNo int) utils.CommandResponse
	getAttn(attnNo int) utils.CommandResponse
}

type sgDevice interface {
	loadLANDetails(name string) bool
	loadCommands() bool
	initializeDevice(name string)
	getCommands(mnemonics []string, arguments []string, replace []string) []utils.Command
	communicate(cmds []utils.Command, port string) []string
	setRFOn() utils.CommandResponse
	setRFOff() utils.CommandResponse
	setModOn() utils.CommandResponse
	setModOff() utils.CommandResponse
	getModulationStatus() utils.CommandResponse
	setFrequency(value float64) utils.CommandResponse
	setPower(value float64) utils.CommandResponse
	getCommandDatabase() map[string]utils.Command
}

type gtxDevice interface {
	loadLANDetails(name string) bool
	loadCommands() bool
	initializeDevice(name string)
	setCarrierOn(component string) utils.CommandResponse
	setCarrierOff(component string) utils.CommandResponse
	setModulationOn(component string) utils.CommandResponse
	setModulationOff(component string) utils.CommandResponse
	setFrequencyDeviationTC(component string, deviation float64) utils.CommandResponse
	setFrequencyDeviationTone(component string, deviation float64) utils.CommandResponse
	setModIndexTC(component string, modIndex float64) utils.CommandResponse
	setModIndexTone(component string, modIndex float64) utils.CommandResponse
	setRangingToneFrequency(frequency float64) utils.CommandResponse
	setOnlyTC(component string) utils.CommandResponse
	setOnlyRanging(component string) utils.CommandResponse
	setTCAndRanging(component string) utils.CommandResponse
	setSweepRate(component string, sweepRate float64) utils.CommandResponse
	setSweepStep(component string, sweepStep float64) utils.CommandResponse
	triggerSweep(component string) utils.CommandResponse
	enableDoppler(component string) utils.CommandResponse
	startSweep(component string) utils.CommandResponse
	stopSweep(component string) utils.CommandResponse
	setSweepRange(component string, sweepRange float64) utils.CommandResponse
	setFrequency(component string, frequency float64) utils.CommandResponse
	setPower(component string, power float64) utils.CommandResponse
	setChipRate(component string, chipRate float64) utils.CommandResponse
	checkConnection() utils.CommandResponse
	setDopplerCompensationEnable() utils.CommandResponse
	setDopplerCompensationDisable() utils.CommandResponse
	getDeviceTime() utils.CommandResponse
	getDopplerCompensationStatus(component string) utils.CommandResponse
	setIdlePatternOn() utils.CommandResponse
	setIdlePatternOff() utils.CommandResponse
	setDopplerCompensationTable(timeOffset int, frequencies []int, extendedFrequencies []int, times []int) utils.CommandResponse
	sweepHold(component string) utils.CommandResponse
	sweepContinuous(component string) utils.CommandResponse
	communicate(cmds []utils.Command, port string) []string
}

type vsaDevice interface {
	saDevice
	setPulseMode() utils.CommandResponse
	setSpectrumParameters(centerFrequency float64, span float64, rbw float64) utils.CommandResponse
	setPulseParameters(acqtime float64, YTop float64, pdiv float64, analLength float64, reflevel float64,
		hystlevel float64, points int32, bufferlength int32) utils.CommandResponse
	startMeasurement() utils.CommandResponse
	stopMeasurement() utils.CommandResponse
	setSpectrogramMode(mode string) utils.CommandResponse
	getPulseParameters() utils.CommandResponse
	getPulseAveragePower() utils.CommandResponse
	waitTillFirstPulse() utils.CommandResponse
	getScreenshot(mode string) utils.CommandResponse
	setSelectedTrace(number int) utils.CommandResponse
	restoreTrace() utils.CommandResponse
	getSpectrumDumpSA() utils.CommandResponse
	getTraceDumpSA(points int) utils.CommandResponse
	startVSA() utils.CommandResponse
	startSA() utils.CommandResponse
	checkSA() utils.CommandResponse
	checkVSA() utils.CommandResponse
	waitForSweeps(i int) utils.CommandResponse
}
