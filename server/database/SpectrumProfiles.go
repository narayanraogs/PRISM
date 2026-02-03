package database

import "context"

func GetAllSpectrumProfiles() ([]string, bool) {
	ctx := context.Background()
	profiles, err := dbObject.getAllSpectrumProfiles(ctx)
	if err != nil {
		return nil, false
	}
	return profiles, true
}

func GetSpectrumProfile(name string) (SpectrumProfile, bool) {
	ctx := context.Background()
	profile, err := dbObject.getSpectrumProfile(ctx, name)
	if err != nil {
		return SpectrumProfile{}, false
	}
	return profile, true
}
