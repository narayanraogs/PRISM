package database

import "context"

func GetReceiverNames() ([]string, bool) {
	ctx := context.Background()
	rx, err := dbObject.getAllRxNames(ctx)
	if err != nil {
		return nil, false
	}
	return rx, true
}

func GetRxFrequency(name string) (int64, bool) {
	ctx := context.Background()
	freq, err := dbObject.getRxFrequency(ctx, name)
	if err != nil {
		return 0, false
	}
	return freq, true
}

func GetAllRxWithFrequency(frequency int64) ([]string, bool) {
	ctx := context.Background()
	rxs, err := dbObject.getAllRxForFrequency(ctx, frequency)
	if err != nil {
		return nil, false
	}
	return rxs, true
}

func GetRxDetails(name string) (SpecRx, bool) {
	ctx := context.Background()
	rx, err := dbObject.getRxDetails(ctx, name)
	if err != nil {
		return SpecRx{}, false
	}
	return rx, true
}

func GetRxTMTC(name string) (SpecRxTMTC, bool) {
	ctx := context.Background()
	rx, err := dbObject.getRxTMTC(ctx, name)
	if err != nil {
		return SpecRxTMTC{}, false
	}
	return rx, true
}
