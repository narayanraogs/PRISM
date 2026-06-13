package resultsDB

import (
	"context"
	"time"

	"prismServer/logger"
)

type StabilityPoint struct {
	TimeStampInt int64
	TimeStamp    string
	Description  string
	Value        float64
}

func StartNewStability() (int64, bool) {
	var arg insertStabilityParams
	now := time.Now()
	ctx := context.Background()
	arg.Date = now.Format("02-01-2006")
	arg.Time = now.Format("15:04:05")
	id, err := dbObject.insertStability(ctx, arg)
	if err != nil {
		return -1, false
	}
	return id, true
}

func InsertPoints(id int64, points []StabilityPoint) bool {
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		logger.Log.Error("Unable to get Transaction Lock", "error", err)
		return false
	}
	defer tx.Rollback()
	q := dbObject.WithTx(tx)
	for _, p := range points {
		var arg insertStabilityPointsParams
		arg.StabilityID = id
		arg.Description = p.Description
		arg.TimeStamp = p.TimeStamp
		arg.TimeStampInteger = p.TimeStampInt
		arg.Value = p.Value
		err = q.insertStabilityPoints(ctx, arg)
		if err != nil {
			return false
		}
	}
	err = tx.Commit()
	if err != nil {
		return false
	}
	return true
}

func GetStabilitySessions() ([]Stability, error) {
	ctx := context.Background()
	rows, err := dbObject.getStabilitySessions(ctx)
	return rows, err
}

func GetStabilityPoints(id int64, parameter string) ([]StabilityPoint, error) {
	ctx := context.Background()
	var arg getStabilityPointsParams
	arg.StabilityID = id
	arg.Description = parameter
	rows, err := dbObject.getStabilityPoints(ctx, arg)
	var points []StabilityPoint
	for _, row := range rows {
		points = append(points, StabilityPoint{
			TimeStampInt: row.TimeStampInteger,
			TimeStamp:    row.TimeStamp,
			Description:  row.Description,
			Value:        row.Value,
		})
	}
	return points, err
}

func GetStabilityParameters(id int64) ([]string, error) {
	ctx := context.Background()
	params, err := dbObject.getStabilityParameters(ctx, id)
	return params, err
}
