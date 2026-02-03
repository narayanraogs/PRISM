package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type PLParameterSpec struct {
	Name             string
	Unit             string
	ExpectedValue    float64
	AllowedDeviation float64
	Operation        func(float64) (float64, string)
}

func GetPLFrequency(name string) (float64, bool) {
	ctx := context.Background()
	freq, err := dbObject.getPLFrequency(ctx, name)
	if err != nil {
		return 0, false
	}
	return freq, true
}

func GetPayloadSpec(name string, testPhase string, mode string) (map[string]PLParameterSpec, bool) {
	_, saLoss, pmLoss, _ := GetCurrentDownlinkLoss(name, testPhase)
	var saLossCalculator = func(value float64) (float64, string) {
		totalLoss := value + saLoss
		return totalLoss, fmt.Sprintf("%.2f", totalLoss)
	}
	var pmLossCalculator = func(value float64) (float64, string) {
		totalLoss := value + pmLoss
		return totalLoss, fmt.Sprintf("%.2f", totalLoss)
	}
	var noOp = func(value float64) (float64, string) {
		return value, strconv.FormatFloat(value, 'G', 4, 64)
	}
	var timeCalculator = func(value float64) (float64, string) {
		value = value * 1e6
		return value, strconv.FormatFloat(value, 'G', 4, 64)
	}
	var mhzConverter = func(value float64) (float64, string) {
		value = value / 1e6
		return value, fmt.Sprintf("%.6f", value)
	}

	var tbr = make(map[string]PLParameterSpec)
	ctx := context.Background()
	var pl SpecPL
	var err error
	if strings.EqualFold(mode, "HR") {
		pl, err = dbObject.getPulseSpecHRMode(ctx, name)
		if err != nil {
			return tbr, false
		}
	} else if strings.EqualFold(mode, "LR") {
		pl, err = dbObject.getPulseSpecLRMode(ctx, name)
		if err != nil {
			return tbr, false
		}
	} else {
		pl, err = dbObject.getFullPulseSpecs(ctx, name)
		if err != nil {
			return tbr, false
		}
	}

	tbr["PeakPowerPPM"] = PLParameterSpec{
		Name:             "PeakPower",
		Unit:             "dBm",
		ExpectedValue:    pl.PeakPower.Float64,
		AllowedDeviation: pl.PeakPowerTolerance.Float64,
		Operation:        pmLossCalculator,
	}
	tbr["PeakPowerVSA"] = PLParameterSpec{
		Name:             "PeakPower",
		Unit:             "dBm",
		ExpectedValue:    pl.PeakPower.Float64,
		AllowedDeviation: pl.PeakPowerTolerance.Float64,
		Operation:        saLossCalculator,
	}
	tbr["PulseAveragePowerPPM"] = PLParameterSpec{
		Name:             "PulseAveragePower",
		Unit:             "dBm",
		ExpectedValue:    pl.AveragePower.Float64,
		AllowedDeviation: pl.AveragePowerTolerance.Float64,
		Operation:        pmLossCalculator,
	}
	tbr["PulseAveragePowerVSA"] = PLParameterSpec{
		Name:             "PulseAveragePower",
		Unit:             "dBm",
		ExpectedValue:    pl.AveragePower.Float64,
		AllowedDeviation: pl.AveragePowerTolerance.Float64,
		Operation:        saLossCalculator,
	}
	tbr["AverageTxPowerPPM"] = PLParameterSpec{
		Name:             "AverageTxPower",
		Unit:             "dBm",
		ExpectedValue:    pl.AverageTxPower.Float64,
		AllowedDeviation: pl.AverageTxPowerTolerance.Float64,
		Operation:        pmLossCalculator,
	}
	tbr["AverageTxPowerVSA"] = PLParameterSpec{
		Name:             "AverageTxPower",
		Unit:             "dBm",
		ExpectedValue:    pl.AverageTxPower.Float64,
		AllowedDeviation: pl.AverageTxPowerTolerance.Float64,
		Operation:        saLossCalculator,
	}
	tbr["DutyCycle"] = PLParameterSpec{
		Name:             "DutyCycle",
		Unit:             "%",
		ExpectedValue:    pl.DutyCycle.Float64,
		AllowedDeviation: pl.DutyCycleTolerance.Float64,
		Operation:        noOp,
	}
	tbr["PulsePeriod"] = PLParameterSpec{
		Name:             "PulsePeriod",
		Unit:             "\u00B5s",
		ExpectedValue:    pl.PulsePeriod,
		AllowedDeviation: pl.PulsePeriodTolerance.Float64,
		Operation:        timeCalculator,
	}
	tbr["PulseWidth"] = PLParameterSpec{
		Name:             "PulseWidth",
		Unit:             "\u00B5s",
		ExpectedValue:    pl.PulseWidth,
		AllowedDeviation: pl.PulseWidthTolerance.Float64,
		Operation:        timeCalculator,
	}
	tbr["PulseSeparation"] = PLParameterSpec{
		Name:             "PulseSeparation",
		Unit:             "\u00B5s",
		ExpectedValue:    pl.PulseSeperation.Float64,
		AllowedDeviation: pl.PulseSeperationTolerance.Float64,
		Operation:        timeCalculator,
	}
	tbr["RiseTime"] = PLParameterSpec{
		Name:             "RiseTime",
		Unit:             "\u00B5s",
		ExpectedValue:    pl.RiseTime.Float64,
		AllowedDeviation: pl.RiseTimeTolerance.Float64,
		Operation:        timeCalculator,
	}
	tbr["FallTime"] = PLParameterSpec{
		Name:             "FallTime",
		Unit:             "\u00B5s",
		ExpectedValue:    pl.FallTime.Float64,
		AllowedDeviation: pl.FallTimeTolerance.Float64,
		Operation:        timeCalculator,
	}
	tbr["ReplicaPeriod"] = PLParameterSpec{
		Name:             "ReplicaPeriod",
		Unit:             "\u00B5s",
		ExpectedValue:    pl.ReplicaPeriod.Float64,
		AllowedDeviation: pl.ReplicaPeriodTolerance.Float64,
		Operation:        timeCalculator,
	}
	tbr["ChirpBandwidth"] = PLParameterSpec{
		Name:             "ChirpBandwidth",
		Unit:             "MHz",
		ExpectedValue:    pl.ChirpBandwidth.Float64,
		AllowedDeviation: pl.ChirpBandwidthTolerance.Float64,
		Operation:        mhzConverter,
	}
	tbr["Bandwidth"] = PLParameterSpec{
		Name:             "Bandwidth",
		Unit:             "MHz",
		ExpectedValue:    pl.ChirpBandwidth.Float64,
		AllowedDeviation: pl.ChirpBandwidthTolerance.Float64,
		Operation:        mhzConverter,
	}
	tbr["RepetitionRate"] = PLParameterSpec{
		Name:             "RepetitionRate",
		Unit:             "Hz",
		ExpectedValue:    pl.RepetitionRate.Float64,
		AllowedDeviation: pl.RepetitionRateTolerance.Float64,
		Operation:        noOp,
	}
	tbr["ReplicaRate"] = PLParameterSpec{
		Name:             "ReplicaRate",
		Unit:             "Hz",
		ExpectedValue:    pl.ReplicaRate.Float64,
		AllowedDeviation: pl.ReplicaRateTolerance.Float64,
		Operation:        noOp,
	}
	tbr["FrequencyShift"] = PLParameterSpec{
		Name:             "FrequencyShift",
		Unit:             "Hz",
		ExpectedValue:    pl.FrequencyShift.Float64,
		AllowedDeviation: pl.FrequencyShiftTolerance.Float64,
		Operation:        noOp,
	}
	tbr["Droop"] = PLParameterSpec{
		Name:             "Droop",
		Unit:             "dB",
		ExpectedValue:    pl.Droop.Float64,
		AllowedDeviation: pl.DroopTolerance.Float64,
		Operation:        noOp,
	}
	tbr["Phase"] = PLParameterSpec{
		Name:             "Phase",
		Unit:             "Deg",
		ExpectedValue:    pl.Phase.Float64,
		AllowedDeviation: pl.PhaseTolerance.Float64,
		Operation:        noOp,
	}
	tbr["Overshoot"] = PLParameterSpec{
		Name:             "Overshoot",
		Unit:             "dB",
		ExpectedValue:    pl.Overshoot.Float64,
		AllowedDeviation: pl.OvershootTolerance.Float64,
		Operation:        noOp,
	}
	tbr["ChirpRate"] = PLParameterSpec{
		Name:             "ChirpRate",
		Unit:             "MHz/\u00B5s",
		ExpectedValue:    pl.ChirpRate.Float64,
		AllowedDeviation: pl.ChirpRateTolerance.Float64,
		Operation:        noOp,
	}
	tbr["ChirpRateDeviation"] = PLParameterSpec{
		Name:             "ChirpRateDeviation",
		Unit:             "%",
		ExpectedValue:    pl.ChirpRateDeviation.Float64,
		AllowedDeviation: pl.ChirpRateDeviationTolerance.Float64,
		Operation:        noOp,
	}
	tbr["Ripple"] = PLParameterSpec{
		Name:             "Ripple",
		Unit:             "dB",
		ExpectedValue:    pl.Ripple.Float64,
		AllowedDeviation: pl.RippleTolerance.Float64,
		Operation:        noOp,
	}
	tbr["Frequency"] = PLParameterSpec{
		Name:             "Frequency",
		Unit:             "MHz",
		ExpectedValue:    pl.CenterFrequency,
		AllowedDeviation: pl.CenterFrequency * 2e-6,
		Operation:        mhzConverter,
	}
	tbr["UplinkPower"] = PLParameterSpec{
		Name:             "UplinkPower",
		Unit:             "dBm",
		ExpectedValue:    pl.UplinkPower,
		AllowedDeviation: 0.5,
		Operation:        noOp,
	}
	return tbr, true
}
