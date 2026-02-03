package database

import (
	"context"
)

func GetPulseProfileNames() ([]string, bool) {
	ctx := context.Background()
	pulse, err := dbObject.getAllPulseProfiles(ctx)
	if err != nil {
		return nil, false
	}
	return pulse, true
}

func GetPPMRelatedParameters(profileName string) (PulseProfile, bool) {
	ctx := context.Background()
	ppmParameters, err := dbObject.getPulseParameters(ctx, profileName)
	if err != nil {
		return PulseProfile{}, false
	}
	return ppmParameters, true
}

func GetFullSpec(configName string) (SpecPL, bool) {
	ctx := context.Background()
	pulseSpecs, err := dbObject.getFullPulseSpecs(ctx, configName)
	if err != nil {
		return SpecPL{}, false
	}
	return pulseSpecs, true
}

func GetSpecHRMode(configName string) (SpecPL, bool) {
	ctx := context.Background()
	pulseSpecs, err := dbObject.getPulseSpecHRMode(ctx, configName)
	if err != nil {
		return SpecPL{}, false
	}
	return pulseSpecs, true
}

func GetSpecLRMode(configName string) (SpecPL, bool) {
	ctx := context.Background()
	pulseSpecs, err := dbObject.getPulseSpecLRMode(ctx, configName)
	if err != nil {
		return SpecPL{}, false
	}
	return pulseSpecs, true
}
