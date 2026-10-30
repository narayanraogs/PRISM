package database

import (
	"context"
	"encoding/json"
)

type FrequencyPair struct {
	Input  int64 `json:"Input"`
	Output int64 `json:"Output"`
}

// ParseFrequencyMap parses the JSON-encoded FrequencyMap field.
func (u UpDownConverter) ParseFrequencyMap() ([]FrequencyPair, error) {
	var pairs []FrequencyPair
	if u.FrequencyMap == "" {
		return pairs, nil
	}
	err := json.Unmarshal([]byte(u.FrequencyMap), &pairs)
	return pairs, err
}

// PrimaryInputFreq returns the first frequency pair's input, or 0.
func (u UpDownConverter) PrimaryInputFreq() float64 {
	pairs, err := u.ParseFrequencyMap()
	if err == nil && len(pairs) > 0 {
		return float64(pairs[0].Input)
	}
	return 0
}

// PrimaryOutputFreq returns the first frequency pair's output, or 0.
func (u UpDownConverter) PrimaryOutputFreq() float64 {
	pairs, err := u.ParseFrequencyMap()
	if err == nil && len(pairs) > 0 {
		return float64(pairs[0].Output)
	}
	return 0
}

func GetAllConverterNames() ([]string, error) {
	ctx := context.Background()
	names, err := dbObject.getAllConverterNames(ctx)
	return names, err
}

func GetConverterDetails(converter string) (UpDownConverter, error) {
	ctx := context.Background()
	details, err := dbObject.getConverterDetails(ctx, converter)
	return details, err
}
