package database

import (
	"context"
	"database/sql"
)

func GetDownlinkPowerProfile(profileName sql.NullString) (DownlinkPowerProfile, bool) {
	ctx := context.Background()
	profile, err := dbObject.getDownlinkPowerProfile(ctx, profileName)
	if err != nil {
		return DownlinkPowerProfile{}, false
	}
	return profile, true
}

func GetPulsePowerProfile(configName string, testType string, testCategory string) (string, bool) {
	ctx := context.Background()
	var arg getTestWithCategoryParams
	arg.ConfigName = configName
	arg.TestType = testType
	arg.TestCategory = testCategory
	profile, err := dbObject.getTestWithCategory(ctx, arg)
	if err != nil {
		return "Unable to get Pulse Power Profile", false
	}
	return profile.PowerProfileName.String, true
}

func GetTRMParameters(profileName sql.NullString) (TRMProfile, bool) {
	ctx := context.Background()
	trmParameters, err := dbObject.getTRMParameters(ctx, profileName)
	if err != nil {
		return TRMProfile{}, false
	}
	return trmParameters, true
}
