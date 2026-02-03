package utils

import (
	"math"
	"slices"
)

type TSMAttnProvider struct {
	RequiredAttn     []float64
	FixedPadIncluded []float64
	MeasuredAttn     []float64
	Difference       []float64
}

func GetCorrectedProfile(measured TSMAttnProvider, fixedPadValue float64, stepSize float64) TSMAttnProvider {
	minValue := slices.Min(measured.RequiredAttn)
	maxValue := slices.Max(measured.RequiredAttn)
	fixedPadValue = fixedPadValue * -1

	var corrected TSMAttnProvider
	for required := minValue; required <= maxValue; required += stepSize {
		fixed := false
		tempRequired := required
		if tempRequired > fixedPadValue {
			fixed = true
			tempRequired = tempRequired - fixedPadValue
		}
		corrected.RequiredAttn = append(corrected.RequiredAttn, required)
		index := getNearest(measured.MeasuredAttn, tempRequired)
		var measuredValue = 0.0
		if index != -1 {
			measuredValue = measured.MeasuredAttn[index]
			if fixed {
				measuredValue = measuredValue + fixedPadValue
			}
		}
		diffValue := required - measuredValue
		corrected.MeasuredAttn = append(corrected.MeasuredAttn, measuredValue)
		corrected.Difference = append(corrected.Difference, diffValue)
		if fixed {
			corrected.FixedPadIncluded = append(corrected.FixedPadIncluded, 1.0)
		} else {
			corrected.FixedPadIncluded = append(corrected.FixedPadIncluded, 0.0)
		}
	}

	return corrected
}

func getNearest(acheived []float64, required float64) int {
	var diff = make([]float64, 0)
	for _, value := range acheived {
		diff = append(diff, math.Abs(value-required))
	}
	minValue := slices.Min(diff)
	return slices.Index(diff, minValue)
}
