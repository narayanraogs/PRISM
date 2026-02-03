package database

import (
	"context"
	"strings"
	"time"
)

func GetAllTestPhases() ([]string, bool) {
	ctx := context.Background()
	tp, err := dbObject.getAllTestPhases(ctx)
	if err != nil {
		return nil, false
	}
	return tp, true
}

func SelectExisitingTestPhase(name string) (string, bool) {
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		return "Unable to get Transaction Lock", false
	}
	defer tx.Rollback()
	q := dbObject.WithTx(tx)
	err = q.deselectTestPhase(ctx)
	if err != nil {
		return "Unable to deselect test phase", false
	}
	err = q.selectTestPhase(ctx, name)
	if err != nil {
		return "Unable to select test phase", false
	}
	err = tx.Commit()
	if err != nil {
		return "Unable to Commit Transaction", false
	}
	return "", true
}

func InsertNewTestPhase(name string, copyFrom string) (string, bool) {
	copyFrom = strings.TrimSpace(copyFrom)
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		return "Unable to get Transaction Lock", false
	}
	defer tx.Rollback()
	q := dbObject.WithTx(tx)
	err = q.deselectTestPhase(ctx)
	if err != nil {
		return "Unable to deselect test phase", false
	}
	var arg addNewTestPhaseParams
	arg.Name = name
	arg.Selected = 1
	now := time.Now()
	arg.CreationDate.String = now.Format("02-01-2006")
	arg.CreationDate.Valid = true
	arg.CreationTime.String = now.Format("15:04:05")
	arg.CreationTime.Valid = true
	err = q.addNewTestPhase(ctx, arg)
	if err != nil {
		return "Unable to create test phase", false
	}
	if !strings.EqualFold(copyFrom, "") {
		dlLoss, err := q.getAllDownlinkLossForTestPhase(ctx, copyFrom)
		if err != nil {
			return "Unable to get Downlink losses", false
		}
		for _, d := range dlLoss {
			var dl insertDownlinkLossParams
			dl.ConfigName = d.ConfigName
			dl.TestPhaseName = name
			dl.Profile = d.Profile
			err = q.insertDownlinkLoss(ctx, dl)
		}
		if err != nil {
			return "Unable to insert Downlink losses", false
		}

		ulLoss, err := q.getAllUplinkLossForTestPhase(ctx, copyFrom)
		if err != nil {
			return "Unable to get Uplink losses", false
		}

		for _, u := range ulLoss {
			var ul insertUplinkLossParams
			ul.ConfigName = u.ConfigName
			ul.TestPhaseName = name
			ul.Profile = u.Profile
			err = q.insertUplinkLoss(ctx, ul)
		}
		if err != nil {
			return "Unable to insert Uplink losses", false
		}
	}
	err = tx.Commit()
	if err != nil {
		return "Unable to Commit Transaction", false
	}
	return "", true
}
