package database

import (
	"context"
	"strconv"
	"strings"
)

func GetCurrentDownlinkLoss(configName string, testPhase string) (float64, float64, float64, bool) {
	ctx := context.Background()
	var arg getDownlinkLossParams
	arg.ConfigName = configName
	arg.TestPhaseName = testPhase
	loss, err := dbObject.getDownlinkLoss(ctx, arg)
	if err != nil {
		return 0.0, 0.0, 0.0, false
	}
	lines := strings.Split(loss, "\n")
	var common float64
	var sa float64
	var pm float64
	for _, line := range lines {
		if strings.EqualFold(strings.TrimSpace(line), "") {
			continue
		}
		fields := strings.Split(line, ",")
		l, _ := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
		switch strings.ToLower(strings.TrimSpace(fields[3])) {
		case "common":
			common = common + l
		case "sa":
			sa = sa + l
		case "pm":
			pm = pm + l
		}
	}
	return common, common + sa, common + pm, true
}

func GetAllConfigsForDownlink(testPhase string) ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getAllConfigsForDownlinkLoss(ctx, testPhase)
	if err != nil {
		return nil, false
	}
	return cfgs, true
}

func GetDownlinkLossProfile(configName string, testPhase string) (string, bool) {
	ctx := context.Background()
	var arg getDownlinkLossParams
	arg.ConfigName = configName
	arg.TestPhaseName = testPhase
	loss, err := dbObject.getDownlinkLoss(ctx, arg)

	if err != nil {
		return "", false
	}
	return loss, true
}

func UpdateDownlinkLossProfile(configName string, testPhase string, profile string) bool {
	ctx := context.Background()
	var arg updateDownlinkLossParams
	arg.ConfigName = configName
	arg.TestPhaseName = testPhase
	arg.Profile = profile
	err := dbObject.updateDownlinkLoss(ctx, arg)

	if err != nil {
		return false
	}
	return true
}
