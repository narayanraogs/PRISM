package database

import (
	"context"
	"database/sql"
	"os"
	"prismServer/logger"

	_ "modernc.org/sqlite"
)

var dbObject *Queries
var db *sql.DB

func Connect(path string) bool {
	if !checkIfDatabaseExists(path) {
		return false
	}
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		logger.Log.Error("Unable to connect to database", "error", err)
		return false
	}
	db.SetMaxOpenConns(1)
	dbObject = New(db)
	return true
}

func checkIfDatabaseExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetSelectedTestPhase() (string, bool) {
	ctx := context.Background()
	tp, err := dbObject.getSelectedTestPhase(ctx)
	if err != nil {
		logger.Log.Warn("Cannot load Selected test phase")
		return "", false
	}
	return tp, true
}
