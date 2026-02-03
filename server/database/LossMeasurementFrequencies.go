package database

import (
	"context"
	"fmt"
	"prismServer/logger"
)

func GetLossMeasurementFrequencies() ([]string, bool) {
	ctx := context.Background()
	values, err := dbObject.getAllLossMeasruementFrequencies(ctx)
	if err != nil {
		logger.Log.Error("Unable to read from Loss Measurement Frequencies", err.Error())
		return []string{}, false
	}
	var freqs = make([]string, 0)
	for _, value := range values {
		var temp string
		temp = value.Description + fmt.Sprintf(";%.2f MHz", value.Frequency/1e6)
		freqs = append(freqs, temp)
	}
	return freqs, true
}

func GetLossMeasurementFrequencyNames() ([]string, bool) {
	ctx := context.Background()
	values, err := dbObject.getAllLossMeasruementFrequencies(ctx)
	if err != nil {
		logger.Log.Error("Unable to read from Loss Measurement Frequencies", err.Error())
		return []string{}, false
	}
	var freqs = make([]string, 0)
	for _, value := range values {
		var temp string
		temp = value.Description
		freqs = append(freqs, temp)
	}
	return freqs, true
}

func GetFrequencyForLossMeasurement(name string) (float64, bool) {
	ctx := context.Background()
	value, err := dbObject.getFrequencyForLossMeasurement(ctx, name)
	if err != nil {
		logger.Log.Error("Unable to read from Loss Measurement Frequencies", err.Error())
		return 0.0, false
	}
	return value, true
}
