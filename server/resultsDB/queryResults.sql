-- name: getStabilityID :one
Select "ID" From "Stability"
Where "Date" = ? and "Time" = ?;

-- name: checkIfCableLossPMReferenceExists :one
Select Count("CableID") From "CableLosses"
Where "CableName" = "PM";

-- name: updateCableLossPMReference :exec
Update "CableLosses" Set "Date" = ?, "Time" = ?, "Loss" = ?
Where "CableName" = "PM";

-- name: insertCableLossEntry :exec
Insert into "CableLosses" ("Date", "Time", "CableName", "CableLength", "Loss")
Values (?,?,?,?,?);

-- name: getPMReference :one
Select "Loss" from "CableLosses"
where "CableName" like "PM";

-- name: getCableLosses :many
Select * from "CableLosses";

-- name: getCableNames :many
Select distinct "CableName" from "TVACCableLosses";

-- name: getCableNamesForCableLoss :many
Select distinct "CableName" from "CableLosses";

-- name: checkIfTVACCableLossPMReferenceExists :one
Select Count("CableID") From "TVACCableLosses"
Where "CableName" = "PM";

-- name: checkIfTVACCableAsReferenceExists :one
Select Count("CableID") From "TVACCableLosses"
Where "Reference" = "1";

-- name: updateTVACCableLossPMReference :exec
Update "TVACCableLosses" Set "Date" = ?, "Time" = ?, "Reference" = ?, "Loss" = ?
Where "CableName" = "PM";

-- name: insertTVACCableLossEntry :exec
Insert into "TVACCableLosses" ("Date", "Time", "CableName", "TestPhase", "Reference", "Loss")
Values (?,?,?,?,?,?);

-- name: getTVACPMReference :one
Select "Loss" from "TVACCableLosses"
where "CableName" like "PM";

-- name: getTVACCableLosses :many
Select * from "TVACCableLosses";

-- name: getTVACReferenceCableLosses :one
Select "Loss" from "TVACCableLosses" where "CableName" = ? and "Reference" = 1;

-- name: insertTSMInternalLoss :exec
Insert into "TSMInternalLoss" ("LossID", "InputPort", "OutputPort", "PathMnemonic", "MeasuredLosses", "MeasurementCompleted")
values (?,?,?,?,?,'No');

-- name: updateTSMInternalLoss :exec
Update "TSMInternalLoss"  set "MeasuredLosses" = ?, "MeasurementCompleted" = 'Yes'
where "InputPort" = ? and "OutputPort" = ?;

-- name: clearTSMInternalLoss :exec
Delete from "TSMInternalLoss";

-- name: getAllTSMInternalLoss :many
Select * from "TSMInternalLoss"
Order by "InputPort" ASC;

-- name: getMeasuredTSMInternalLoss :one
Select "PathMnemonic","MeasuredLosses" from "TSMInternalLoss"
where "InputPort" = ? and "OutputPort" = ?;

-- name: getPMOffsetForTSMInternalLoss :one
Select "MeasuredLosses" from "TSMInternalLoss"
where "PathMnemonic" like "PM-Measurement";

-- name: getCableLossForTSMInternalLoss :one
Select "MeasuredLosses" from "TSMInternalLoss"
where "PathMnemonic" like "Cable-Measurement";

-- name: updatePMOffsetTSMInternalLoss :exec
Update "TSMInternalLoss"  set "MeasuredLosses" = ?, "MeasurementCompleted" = 'Yes'
where "PathMnemonic" like "PM-Measurement";

-- name: updateCableLossTSMInternalLoss :exec
Update "TSMInternalLoss"  set "MeasuredLosses" = ?, "MeasurementCompleted" = 'Yes'
where "PathMnemonic" like "Cable-Measurement";

-- name: insertStability :one
Insert into "Stability" ("Date", "Time")
Values (?, ?) Returning "ID";

-- name: getStabilitySessions :many
Select "ID", "Date", "Time" from "Stability"
Order by "ID" DESC;

-- name: insertStabilityPoints :exec
Insert into "StabilityValues" ("StabilityID", "TimeStampInteger", "TimeStamp", "Description", "Value")
Values (?, ?, ?, ?, ?);

-- name: getStabilityPoints :many
Select * from "StabilityValues"
where "StabilityID" = ? and "Description" = ?;

-- name: getStabilityParameters :many
Select Distinct "Description" from "StabilityValues"
where "StabilityID" = ?;

-- name: insertResults :exec
Insert into "Results" ("SatName", "TestPhase", "TestType", "TestCategory", "ConfigName", "Date", "Time", "Remark", "Report", "FilePath", "CSVFilePath")
Values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: getResults :many
Select * from "Results"
where "TestPhase" like ? and "TestType" like ? and "TestCategory" like ?
and  "ConfigName" like ? and "Date" like ?
Order by "TestID" Desc;

-- name: getSingleResult :one
Select * FROM "Results"
WHERE "Date" = ? AND "Time" = ?;

-- name: updateResult :exec
UPDATE "Results"
SET "Report" = ?, "FilePath" = ?
WHERE "Date" = ? AND "Time" = ?;

-- name: getOfflineTestPhases :many
Select DISTINCT "TestPhase" from "Results"
Order by "TestID" desc;

-- name: insertUpDownConverterResult :exec
Insert into "UpDownConverter" ("Name", "TestType", "Date", "Time", "Results")
Values (?, ?, ?, ?, ?);

-- name: getUpDownConverterResult :one
Select * from "UpDownConverter"
where "TestType" like ? and "Name" like ?
Order by "ID" Desc Limit 1;

-- name: getUpDownConverterResultWithDateAndTime :one
Select * from "UpDownConverter"
where "Name" like ? and "Date" like ? and "Time" like ?;

-- name: getAllResultsForConverter :many
Select * from "UpDownConverter" where "Name" like ?
Order by "ID" Desc;


-- name: getDistinctConvertersFromResultDB :many
SELECT DISTINCT "Name" FROM "UpDownConverter";
