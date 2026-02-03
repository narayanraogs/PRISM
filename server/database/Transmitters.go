package database

import (
	"context"
)

func GetTxFrequency(name string) (int64, bool) {
	ctx := context.Background()
	freq, err := dbObject.getTxFrequency(ctx, name)
	if err != nil {
		return 0, false
	}
	return freq, true
}

func GetTxSpecs(name string) (SpecTx, bool) {
	ctx := context.Background()
	tx, err := dbObject.getTxSpecs(ctx, name)
	if err != nil {
		return SpecTx{}, false
	}
	return tx, true
}

func GetTxHarmonicsDetails(name string) ([]SpecTxHarmonic, bool) {
	ctx := context.Background()
	var harms = make([]SpecTxHarmonic, 0)

	txHarm, err := dbObject.getTxHarmonics(ctx, name)
	if err != nil {
		return []SpecTxHarmonic{}, false
	}
	for _, harm := range txHarm {
		harms = append(harms, harm)
	}
	return harms, true
}

func GetTxSubCarriersDetails(name string) ([]SpecTxSubCarrier, bool) {
	ctx := context.Background()
	var subCarriers = make([]SpecTxSubCarrier, 0)

	txsubCarriers, err := dbObject.getSpexTxSubCarriersDetails(ctx, name)
	if err != nil {
		return []SpecTxSubCarrier{}, false
	}
	for _, carrier := range txsubCarriers {
		subCarriers = append(subCarriers, carrier)
	}
	return subCarriers, true
}


