package database

import "context"

func GetFrequencyProfile(name string) (FrequencyProfile, bool) {
	ctx := context.Background()
	freq, err := dbObject.getFrequencyProfile(ctx, name)
	if err != nil {
		return FrequencyProfile{}, false
	}
	return freq, true
}
