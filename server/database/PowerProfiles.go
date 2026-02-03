package database

import (
	"context"
	"prismServer/utils"
	"strings"
)

func GetPowerLevels(profileName string) ([]float64, bool) {
	ctx := context.Background()
	profile, err := dbObject.getPowerLevels(ctx, profileName)
	if err != nil {
		return nil, false
	}
	temp := strings.Split(profile, ",")
	fValues := utils.GetFloatArray(temp)
	return fValues, true
}

func GetNoOfCommandsInProfile(profileName string) (int, int, bool) {
	ctx := context.Background()
	profile, err := dbObject.getCommandsInPowerProfile(ctx, profileName)
	if err != nil {
		return -1, -1, false
	}

	return int(profile.NoOfCommandsAtThreshold), int(profile.NoOfCommandsAtOtherLevels), true
}
