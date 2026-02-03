CREATE TABLE "Configurations" (
    "ID"                    INTEGER   PRIMARY KEY AUTOINCREMENT,
    "ConfigName"            TEXT UNIQUE NOT NULL,
    "ConfigType"            TEXT  NOT NULL,
    "RxName"                TEXT ,
    "TxName"                TEXT ,
    "TpName"                TEXT ,
    "PayloadName"           TEXT ,
    "TSMConfigurationName"  TEXT  NOT NULL,
    "CortexIFM"             INTEGER,
    "IntermediateFrequency" INTEGER,
    "ProgrammableAttnUsed"  TEXT ,
    "DeviceProfileName"     TEXT  NOT NULL
);


CREATE TABLE "DeviceProfile" (
    "ID"                INTEGER   PRIMARY KEY AUTOINCREMENT,
    "DeviceProfileName" TEXT  UNIQUE Not Null,
    "SAName"            TEXT ,
    "VSAName"           TEXT ,
    "PMName"            TEXT ,
    "PPMName"           TEXT ,
    "TSMName"           TEXT ,
    "GTxName"           TEXT ,
    "SGName"            TEXT ,
    "VSGName"           TEXT
);

CREATE TABLE "Devices" (
    "ID"                   INTEGER   PRIMARY KEY AUTOINCREMENT,
    "DeviceName"           TEXT  UNIQUE NOT NULL,
    "DeviceMake"           TEXT  NOT NULL,
    "DeviceType"           TEXT  NOT NULL,
    "IPAddress"            TEXT  NOT NULL,
    "ControlPort"          INTEGER   NOT NULL,
    "AlternateControlPort" INTEGER,
    "ReadPort"             INTEGER,
    "DopplerPort"          INTEGER,
    "TimeoutInMillisecs"   INTEGER   NOT NULL
);

CREATE TABLE "DownlinkLoss" (
    "ID"            INTEGER     PRIMARY KEY AUTOINCREMENT,
    "ConfigName"    TEXT Not Null,
    "TestPhaseName" TEXT Not Null,
    "Profile"       TEXT Not Null,
    UNIQUE (
        "ConfigName",
        "TestPhaseName"
    )
);

CREATE TABLE "DownlinkPowerProfile" (
    "ID"         INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Name"       TEXT  UNIQUE,
    "PMChannel"  TEXT      NOT NULL,
    "OccupiedBW" INTEGER   NOT NULL
);

CREATE TABLE "FrequencyProfile" (
    "ID"                 INTEGER    PRIMARY KEY AUTOINCREMENT,
    "Name"               TEXT   UNIQUE NOT NULL,
    "MaxFrequency"       REAL,
    "StepSize"           REAL,
    "CommandingRequired" TEXT   NOT NULL,
    "DopplerFile"        TEXT
);

CREATE TABLE "LossMeasurementFrequencies" (
    "ID"          INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Description" TEXT  UNIQUE NOT NULL,
    "Frequency"   REAL Not Null
);

CREATE TABLE "PowerProfile" (
    "ID"          INTEGER    PRIMARY KEY AUTOINCREMENT,
    "Name"        TEXT   UNIQUE NOT NULL,
    "PowerLevels" TEXT  NOT NULL,
    "NoOfCommandsAtThreshold" INTEGER NOT NULL,
    "NoOfCommandsAtOtherLevels" INTEGER NOT NULL
);

CREATE TABLE "PulseProfile" (
    "ID"                INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Name"              TEXT  UNIQUE NOT NULL,
    "TransientON"       REAL      NOT NULL,
    "IQOn"              INTEGER   NOT NULL,
    "AcquisitionTime"      INTEGER   NOT NULL,
    "SweepTime"         REAL      NOT NULL,
    "SweepCount"        INTEGER   NOT NULL,
    "FilterType"        TEXT  NOT NULL,
    "FilterBandwidth"   REAL      NOT NULL,
    "YTop"              REAL      NOT NULL,
    "ThresholdLevel"    REAL      NOT NULL,
    "Hysterisis"        REAL      NOT NULL,
    "PPMTriggerLevel"   REAL      NOT NULL,
    "PPMReferenceLevel" REAL      NOT NULL,
    "PPMYDivision"      REAL      NOT NULL,
    "PPMChannel"        TEXT      NOT NULL
);

CREATE TABLE "SpecPL" (
    "SpecID"                      INTEGER PRIMARY KEY AUTOINCREMENT,
    "ConfigName"                  TEXT    NOT NULL,
    "ResolutionMode"              TEXT,
    "OnTime"                      REAL    NOT NULL,
    "CenterFrequency"             NUMERIC NOT NULL,
    "UplinkPower"                 REAL    NOT NULL,
    "PeakPower"                   REAL,
    "PeakPowerTolerance"          REAL,
    "AveragePower"                REAL,
    "AveragePowerTolerance"       REAL,
    "DutyCycle"                   REAL,
    "DutyCycleTolerance"          REAL,
    "PulsePeriod"                 REAL    NOT NULL,
    "PulsePeriodTolerance"        REAL,
    "ReplicaPeriod"               REAL,
    "ReplicaPeriodTolerance"      REAL,
    "PulseWidth"                  REAL    NOT NULL,
    "PulseWidthTolerance"         REAL,
    "PulseSeperation"             REAL,
    "PulseSeperationTolerance"    REAL,
    "RiseTime"                    REAL,
    "RiseTimeTolerance"           REAL,
    "FallTime"                    REAL,
    "FallTimeTolerance"           REAL,
    "AverageTxPower"              REAL,
    "AverageTxPowerTolerance"     REAL,
    "ChirpBandwidth"              REAL,
    "ChirpBandwidthTolerance"     REAL,
    "RepetitionRate"              REAL,
    "RepetitionRateTolerance"     REAL,
    "ReplicaRate"                 REAL,
    "ReplicaRateTolerance"        REAL,
    "FrequencyShift"              REAL,
    "FrequencyShiftTolerance"     REAL,
    "Droop"                       REAL,
    "DroopTolerance"              REAL,
    "Phase"                       REAL,
    "PhaseTolerance"              REAL,
    "Overshoot"                   REAL,
    "OvershootTolerance"          REAL,
    "ChirpRate"                   REAL,
    "ChirpRateTolerance"          REAL,
    "ChirpRateDeviation"          REAL,
    "ChirpRateDeviationTolerance" REAL,
    "Ripple"                      REAL,
    "RippleTolerance"             REAL,
    UNIQUE("ConfigName","ResolutionMode")
);

CREATE TABLE "SpecRx" (
    "ID"                    INTEGER      PRIMARY KEY AUTOINCREMENT,
    "RxName"                TEXT     UNIQUE Not Null,
    "Frequency"             INTEGER NOT NULL,
    "MaxPower"              REAL         NOT NULL,
    "TCSubCarrierFrequency" REAL         NOT NULL,
    "ModulationScheme"      TEXT     NOT NULL,
    "AcquisitionOffset"     REAL,
    "SweepRange"            REAL,
    "SweepRate"             REAL,
    "TCModIndex"            REAL,
    "FrequencyDeviationFM"  REAL,
    "CodeRateInMcps"        REAL
);

CREATE TABLE "SpecRxTMTC" (
    "RxName"                 TEXT PRIMARY KEY,
    "LockStatusMnemonic"     TEXT ,
    "LockStatusValue"        TEXT ,
    "BSLockStatusMnemonic"   TEXT  NOT NULL,
    "BSLockStatusValue"      TEXT  NOT NULL,
    "AGCMnemonic"            TEXT  NOT NULL,
    "LoopStressMnemonic"     TEXT ,
    "CommandCounterMnemonic" TEXT NOT NULL,
    "TestCommandSet"         TEXT NOT NULL,
    "TestCommandReset"       TEXT NOT NULL
);


CREATE TABLE "SpecTp" (
    "TpID"   INTEGER   PRIMARY KEY AUTOINCREMENT,
    "TpName" TEXT  UNIQUE,
    "RxName" TEXT      NOT NULL,
    "TxName" TEXT  NOT NULL
);

CREATE TABLE "SpecTpRanging" (
    "TpName"                                TEXT  NOT NULL,
    "RangingID"                             INTEGER   NOT NULL,
    "RangingName"                           TEXT  NOT NULL,
    "ToneFrequency"                         REAL      NOT NULL,
    "UplinkToneMIOnlyRanging"               REAL      NOT NULL,
    "UplinkToneMISimultaneousCmdAndRanging" REAL,
    "TCMISimultaneousCmdAndRanging"         REAL      NOT NULL,
    "DownlinkMI"                            REAL      NOT NULL,
    "AllowedDownlinkMIDeviation"            REAL      NOT NULL,
    "AvailableForCommanding"                TEXT  NOT NULL,
    PRIMARY KEY ("TpName","RangingID" )
);

CREATE TABLE "SpectrumProfile" (
    "ID"              INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Name"            TEXT  UNIQUE NOT NULL,
    "CenterFrequency" REAL      NOT NULL,
    "Span"            REAL      NOT NULL,
    "RBW"             INTEGER   NOT NULL,
    "VBW"             INTEGER   NOT NULL
);

CREATE TABLE "SpecTx" (
    "TxID"                      INTEGER   PRIMARY KEY AUTOINCREMENT,
    "TxName"                    TEXT  UNIQUE Not NUll,
    "Frequency"                 INTEGER   NOT NULL,
    "Power"                     REAL      NOT NULL,
    "Spurious"                  REAL      NOT NULL,
    "Harmonics"                 REAL      NOT NULL,
    "AllowedFrequencyDeviation" REAL      NOT NULL,
    "AllowedPowerDevaition"     REAL      NOT NULL,
    "ModulationScheme"          TEXT  NOT NULL,
    "IsBurst"                   TEXT  NOT NULL,
    "BurstTime"                 REAL
);

CREATE TABLE "SpecTxHarmonics" (
    "TxName"              TEXT  NOT NULL,
    "HarmonicsID"         INTEGER   NOT NULL,
    "HarmonicType"        TEXT  NOT NULL,
    "HarmonicsName"       TEXT  NOT NULL,
    "Frequency"           REAL      NOT NULL,
    "TotalLossFromTxToSA" REAL      NOT NULL,
    PRIMARY KEY ("TxName","HarmonicsID")
);

CREATE TABLE "SpecTxSubCarriers" (
    "TxName"                   TEXT  NOT NULL,
    "SubCarrierID"             INTEGER   NOT NULL,
    "SubCarrierName"           TEXT  NOT NULL,
    "Frequency"                REAL      NOT NULL,
    "ModIndex"                 REAL,
    "AllowedModIndexDeviation" REAL,
    "PeakFrequencyDeviation"   REAL,
    "AlwaysPresent"            TEXT  NOT NULL,
    PRIMARY KEY ("TxName","SubCarrierID")
);

CREATE TABLE "TestPhases" (
    "ID"           INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Name"         TEXT  UNIQUE NOT NULL,
    "CreationDate" TEXT ,
    "CreationTime" TEXT ,
    "Selected"     INTEGER   NOT NULL
);

CREATE TABLE "Tests" (
    "ID"                       INTEGER   PRIMARY KEY AUTOINCREMENT,
    "ConfigName"               TEXT  NOT NULL,
    "TestType"                 TEXT  NOT NULL,
    "TestCategory"             TEXT  NOT NULL,
    "ULProfileName"            TEXT ,
    "DLProfileName"            TEXT ,
    "PowerProfileName"         TEXT ,
    "FrequencyProfileName"     TEXT ,
    "DownlinkPowerProfileName" TEXT ,
    "PulseProfileName"         TEXT ,
    "TRMProfileName"           TEXT ,
    "TMProfileName"            TEXT
);

CREATE TABLE "TMProfile" (
    "ID"             INTEGER    PRIMARY KEY AUTOINCREMENT,
    "Name"           TEXT   UNIQUE Not Null,
    "PreRequisiteTM" TEXT ,
    "LogTM"          TEXT
);

CREATE TABLE "TRMProfile" (
    "ID"                         INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Name"                       TEXT  UNIQUE,
    "NoOfTRMs"                   INTEGER   NOT NULL,
    "TimePerTRMInSecs"           REAL      NOT NULL,
    "DelayBeforeFirstReadInSecs" REAL      NOT NULL
);

CREATE TABLE "TSMConfigurations" (
    "ID"              INTEGER   PRIMARY KEY AUTOINCREMENT,
    "Name"            TEXT  UNIQUE Not Null,
    "UplinkToSC"      TEXT ,
    "IncludePad"      TEXT ,
    "ExcludePad"      TEXT ,
    "UplinkToSA"      TEXT ,
    "UplinkToPM"      TEXT ,
    "TerminateUplink" TEXT ,
    "DownlinkToSA"    TEXT ,
    "DownlinkToPM"    TEXT ,
    "AttnNumber"      INTEGER,
    "DownlinkToDC"    TEXT ,
    "InputPortName"   TEXT ,
    "SAPortName"      TEXT ,
    "PMPortName"      TEXT ,
    "OutputPortName"  TEXT
);

CREATE TABLE "UpDownConverter" (
    "ID"               INTEGER PRIMARY KEY AUTOINCREMENT,
    "Name"             TEXT    UNIQUE NOT NULL,
    "InputFrequency"   REAL    NOT NULL,
    "OutputFrequency"  REAL    NOT NULL,
    "MaxPowerCable"    REAL    NOT NULL,
    "MinPowerCable"    REAL    NOT NULL,
    "MaxPowerRadiated" REAL,
    "MinPowerRadiated" REAL
);

CREATE TABLE "UplinkLoss" (
    "ID"            INTEGER     PRIMARY KEY AUTOINCREMENT,
    "ConfigName"    TEXT Not Null,
    "TestPhaseName" TEXT Not Null,
    "Profile"       TEXT Not Null,
    UNIQUE ("ConfigName","TestPhaseName")
);
