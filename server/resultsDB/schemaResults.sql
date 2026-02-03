CREATE TABLE "Results" (
    "TestID"       INTEGER     PRIMARY KEY AUTOINCREMENT,
    "SatName"      TEXT   NOT NULL,
    "TestPhase"    TEXT   NOT NULL,
    "TestType"     TEXT   NOT NULL,
    "TestCategory" TEXT ,
    "ConfigName"   TEXT   NOT NULL,
    "Date"         TEXT    NOT NULL,
    "Time"         TEXT    NOT NULL,
    "Remark"       TEXT ,
    "Report"       TEXT        NOT NULL,
    "FilePath"     TEXT  NOT NULL,
    "CSVFilePath"  TEXT 
);


CREATE TABLE "Stability" (
    "ID"   INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Date" TEXT NOT NULL,
    "Time" TEXT NOT NULL
);

CREATE TABLE "StabilityValues" (
    "ID"               INTEGER    PRIMARY KEY AUTOINCREMENT,
    "StabilityID"      INTEGER    NOT NULL,
    "TimeStampInteger" INTEGER    NOT NULL,
    "TimeStamp"        TEXT  NOT NULL,
    "Description"      TEXT NOT NULL,
    "Value"            REAL       NOT NULL
);

CREATE TABLE "CableLosses" (
    "CableID"     INTEGER     PRIMARY KEY AUTOINCREMENT,
    "Date"        TEXT NOT NULL,
    "Time"        TEXT    NOT NULL,
    "CableName"   TEXT NOT NULL,
    "CableLength" INTEGER     NOT NULL,
    "Loss"        TEXT NOT NULL
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

CREATE TABLE "TSMInternalLoss" (
    "LossID"         INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "InputPort"      TEXT    NOT NULL,
    "OutputPort"     TEXT    NOT NULL,
    "PathMnemonic"   TEXT,
    "MeasuredLosses" TEXT    NOT NULL,
    "MeasurementCompleted"	TEXT NOT NULL
);
