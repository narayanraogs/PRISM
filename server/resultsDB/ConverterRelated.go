package resultsDB

import (
	"context"
	"database/sql"
	"time"
)

func InsertUpDownConverterResult(name string, testType string, results string) error {
	ctx := context.Background()
	var arg insertUpDownConverterResultParams
	arg.Name = name
	arg.TestType = testType
	arg.Results = results
	currentTime := time.Now()
	arg.Date = sql.NullString{String: currentTime.Format("02-01-2006"), Valid: true}
	arg.Time = sql.NullString{String: currentTime.Format("15:04:05"), Valid: true}
	dbObject.insertUpDownConverterResult(ctx, arg)
	return nil
}

func GetUpDownConverterResult(name string, testType string) (UpDownConverter, error) {
	ctx := context.Background()
	var arg getUpDownConverterResultParams
	arg.Name = name
	arg.TestType = testType
	value, err := dbObject.getUpDownConverterResult(ctx, arg)
	return value, err
}

func GetAllResultsForConverter(name string) ([]UpDownConverter, error) {
	ctx := context.Background()
	values, err := dbObject.getAllResultsForConverter(ctx, name)
	return values, err
}

func GetUpDownConverterResultWithDateAndTime(name string, date string, time string) (UpDownConverter, error) {
	ctx := context.Background()
	var arg getUpDownConverterResultWithDateAndTimeParams
	arg.Name = name
	arg.Date = sql.NullString{String: date, Valid: true}
	arg.Time = sql.NullString{String: time, Valid: true}
	value, err := dbObject.getUpDownConverterResultWithDateAndTime(ctx, arg)
	return value, err
}
