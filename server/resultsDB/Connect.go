package resultsDB

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var dbObject *Queries
var db *sql.DB

func Connect(path string) bool {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		fmt.Println("Unable to connect to Results database", err.Error())
		return false
	}
	db.SetMaxOpenConns(1)
	dbObject = New(db)
	return true
}
