package database

import (
	"context"
	"strconv"
	"strings"
)

func GetCurrentUplinkLoss(configName string, testPhase string) (float64, float64, float64, float64, bool) {
	ctx := context.Background()
	var arg getUplinkLossParams
	arg.ConfigName = configName
	arg.TestPhaseName = testPhase
	loss, err := dbObject.getUplinkLoss(ctx, arg)
	if err != nil {
		return 0.0, 0.0, 0.0, 0.0, false
	}
	lines := strings.Split(loss, "\n")
	var common float64
	var sa float64
	var pm float64
	var sc float64
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
		case "sc":
			sc = sc + l
		}
	}
	return common, common + sa, common + pm, common + sc, true
}

func GetAllConfigsForUplinkLoss(testPhase string) ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getAllConfigsForUplinkLoss(ctx, testPhase)
	if err != nil {
		return nil, false
	}
	return cfgs, true
}

func GetUplinkLossProfile(configName string, testPhase string) (string, bool) {
	ctx := context.Background()
	var arg getUplinkLossParams
	arg.ConfigName = configName
	arg.TestPhaseName = testPhase
	loss, err := dbObject.getUplinkLoss(ctx, arg)

	if err != nil {
		return "", false
	}
	return loss, true
}

func UpdateUplinkLossProfile(configName string, testPhase string, profile string) bool {
	ctx := context.Background()
	var arg updateUplinkLossParams
	arg.ConfigName = configName
	arg.TestPhaseName = testPhase
	arg.Profile = profile
	err := dbObject.updateUplinkLoss(ctx, arg)

	if err != nil {
		return false
	}
	return true
}
