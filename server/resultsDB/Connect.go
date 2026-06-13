package resultsDB

import (
	"database/sql"
	"os"

	"prismServer/logger"

	_ "modernc.org/sqlite"
)

var dbObject *Queries
var db *sql.DB

func Connect(path string) bool {
	var createNewSchema = false
	var err error
	if !checkIfDatabaseExists(path) {
		createNewSchema = true
	}
	db, err = sql.Open("sqlite", path)
	if err != nil {
		logger.Log.Error("Unable to connect to Results database", "error", err)
		return false
	}
	db.SetMaxOpenConns(1)

	if createNewSchema {
		if !createSchema(db) {
			return false
		}
	}
	dbObject = New(db)
	return true
}

func checkIfDatabaseExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func createSchema(db *sql.DB) bool {
	_, err := db.Exec(createQuery)
	if err != nil {
		logger.Log.Error("Unable to create schema", "error", err)
		return false
	}
	return true
}

var createQuery = `
CREATE TABLE CableLosses (
    CableID     INTEGER     PRIMARY KEY AUTOINCREMENT,
    Date        TEXT (45)   NOT NULL,
    Time        TEXT (45)   NOT NULL,
    CableName   TEXT (100)  NOT NULL,
    CableLength INTEGER     NOT NULL,
    Loss        TEXT (1000) NOT NULL
);
CREATE TABLE Results (
	TestID INTEGER PRIMARY KEY AUTOINCREMENT, 
	SatName TEXT (255) NOT NULL, 
	TestPhase TEXT (255) NOT NULL, 
	TestType TEXT (255) NOT NULL, 
	TestCategory TEXT (255), 
	ConfigName TEXT (255) NOT NULL, 
	Date TEXT (40) NOT NULL, 
	Time TEXT (40) NOT NULL, 
	Remark TEXT (255), 
	Report TEXT NOT NULL, 
	FilePath TEXT (1000) NOT NULL, 
	CSVFilePath TEXT (1000)
);
CREATE TABLE Stability (
	ID INTEGER PRIMARY KEY AUTOINCREMENT, 
	Date TEXT (45) NOT NULL, 
	Time TEXT (45) NOT NULL
);
CREATE TABLE StabilityValues (
	ID INTEGER PRIMARY KEY AUTOINCREMENT, 
	StabilityID INTEGER NOT NULL, 
	TimeStampInteger INTEGER NOT NULL, 
	TimeStamp TEXT (200) NOT NULL, 
	Description TEXT (255) NOT NULL, 
	Value REAL NOT NULL
);
CREATE TABLE "TSMInternalLoss" (
	"LossID"	INTEGER NOT NULL,
	"InputPort"	TEXT NOT NULL,
	"OutputPort"	TEXT NOT NULL,
	"PathMnemonic"	TEXT,
	"MeasuredLosses"	TEXT NOT NULL,
	"MeasurementCompleted"	TEXT NOT NULL DEFAULT 'No',
	PRIMARY KEY("LossID" AUTOINCREMENT)
);
CREATE TABLE "TVACCableLosses" (
    "CableID"     INTEGER     PRIMARY KEY AUTOINCREMENT,
    "Date"        TEXT NOT NULL,
    "Time"        TEXT    NOT NULL,
    "CableName"   TEXT NOT NULL,
    "TestPhase" TEXT     NOT NULL,
    "Reference"   TEXT NOT NULL,
    "Loss"        TEXT NOT NULL
);
CREATE TABLE "UpDownConverter" (
        "ID"    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
        "Name"  TEXT NOT NULL,
        "TestType"      TEXT NOT NULL,
        "Date"  TEXT,
        "Time"  TEXT,
        "Results"       TEXT NOT NULL
)
`
