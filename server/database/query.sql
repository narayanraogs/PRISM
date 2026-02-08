-- name: getSelectedTestPhase :one
Select "Name" From "TestPhases"
Where "Selected" = 1;

-- name: getAllLossMeasruementFrequencies :many
Select "Description", "Frequency" from "LossMeasurementFrequencies"
Order by "ID" asc;

-- name: getAllLossMeasruementFrequencyNames :many
Select "Description" from "LossMeasurementFrequencies"
Order by "ID" asc;

-- name: getFrequencyForLossMeasurement :one
Select "Frequency" from "LossMeasurementFrequencies"
Where "Description" = ?;

-- name: getAllTSM :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'TSM'
Order by "ID" asc;

-- name: getAllSAAndVSA :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'SA' or "DeviceType" like 'VSA'
Order by "ID" asc;

-- name: getAllVSA :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'VSA' Order by "ID" asc;

-- name: getAllSA :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'SA' Order by "ID" asc;

-- name: getAllSpectrumProfiles :many
Select "Name" from "SpectrumProfile"
Order by "ID" asc;

-- name: getSpectrumProfile :one
Select * from "SpectrumProfile"
Where "Name" like ?;

-- name: getAllPMAndPPM :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'PM' or "DeviceType" like 'PPM'
Order by "ID" asc;

-- name: getAllPM :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'PM' Order by "ID" asc;

-- name: getAllPPM :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'PPM' Order by "ID" asc;

-- name: getAllSG :many
Select "DeviceName" from "Devices"
Where "DeviceType" like 'SG' Order by "ID" asc;

-- name: getAllPulseProfiles :many
Select "Name" from "PulseProfile";

-- name: getAllPayloadConfigs :many
Select "ConfigName" from "Configurations"
where ConfigType like 'PL';

-- name: getAllConfigs :many
Select "ConfigName" from "Configurations";

-- name: getAllDeviceProfiles :many
Select "DeviceProfileName" from "DeviceProfile";

-- name: getAllRxNames :many
Select "RxName" from "SpecRx";

-- name: getAllTSMConfigurations :many
Select "Name" from "TSMConfigurations";

-- name: getAllTestPhases :many
Select "Name" From "TestPhases"
Order by "Selected" Desc;

-- name: getDeviceDetails :one
Select * from "Devices"
Where "DeviceName" like ?;

-- name: getTSMConfigNameForConfig :one
Select "TSMConfigurationName" from "Configurations"
Where "ConfigName" like ?;

-- name: getAllPathsInTSMConfig :one
Select * from "TSMConfigurations"
Where "Name" like ?;

-- name: getPMFromDeviceProfile :one
Select "PMName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getSGFromDeviceProfile :one
Select "SGName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getSAFromDeviceProfile :one
Select "SAName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getVSAFromDeviceProfile :one
Select "VSAName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getPPMFromDeviceProfile :one
Select "PPMName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getGTxFromDeviceProfile :one
Select "GTxName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getTSMFromDeviceProfile :one
Select "TSMName" from "DeviceProfile"
where "DeviceProfileName" like ?;

-- name: getRxFrequency :one
Select "Frequency" from "SpecRx"
where "RxName" like ?;

-- name: getTxFrequency :one
Select "Frequency" from "SpecTx"
where "TxName" like ?;

-- name: getTxSpecs :one
Select * from "SpecTx"
where "TxName" like ?;

-- name: getPLFrequency :one
Select "CenterFrequency" from "SpecPL"
where "ConfigName" like ? ;

-- name: getAllRxForFrequency :many
Select "RxName" from "SpecRx"
where "Frequency" = ?;

-- name: getConfigurationNamesForTSMConfig :many
Select "ConfigName" from "Configurations"
where "TSMConfigurationName" like ?;

-- name: getConfigurationDetails :one
Select * from "Configurations"
where "ConfigName" like ?;

-- name: deselectTestPhase :exec
Update "TestPhases" set "Selected" = 0;

-- name: selectTestPhase :exec
Update "TestPhases" set "Selected" = 1
where "Name" = ?;

-- name: addNewTestPhase :exec
Insert Into "TestPhases" ("Name", "CreationDate", "CreationTime", "Selected")
values (?,?,?,?);

-- name: getAllDownlinkLossForTestPhase :many
Select * from "DownlinkLoss"
where "TestPhaseName" = ?;

-- name: getAllUplinkLossForTestPhase :many
Select * from "UplinkLoss"
where "TestPhaseName" = ?;

-- name: insertDownlinkLoss :exec
Insert into "DownlinkLoss" ("ConfigName", "TestPhaseName", "Profile")
values (?,?,?);

-- name: insertUplinkLoss :exec
Insert into "UplinkLoss" ("ConfigName", "TestPhaseName", "Profile")
values (?,?,?);

-- name: getPulseParameters :one
Select * from "PulseProfile"
where "Name" = ?;

-- name: getFullPulseSpecs :one
Select * from "SpecPL" where "ConfigName" = ? and "ResolutionMode" not in ("HR","LR") or "ResolutionMode" is null;

-- name: getTestWithoutCategory :one
Select * from "Tests"
Where "ConfigName" like ? and "TestType" like  ?;

-- name: getTestWithCategory :one
Select * from "Tests"
where "ConfigName" like ? and "TestType" like ? and "TestCategory" like ?;

-- name: getDownlinkLoss :one
Select "Profile" from "DownlinkLoss"
where "ConfigName" like ? and "TestPhaseName" like ?;

-- name: getAllConfigsWithTypes :many
Select "ConfigType", "ConfigName" from "Configurations";

-- name: getAllConfigsForTests :many
Select distinct "Configurations"."ConfigType", "Tests"."ConfigName" from "Tests"
inner join "Configurations" ON "Tests"."ConfigName" = "Configurations"."ConfigName";

-- name: getAllTestForConfig :many
Select "TestType", "TestCategory" from "Tests"
where "ConfigName" like ? and "TestType" not like "%RFUplink%";

-- name: getTMParameters :one
Select * from "TMProfile"
Where "Name" like ?;

-- name: getTxHarmonics :many
Select * from "SpecTxHarmonics"
where "TxName" like ?;

-- name: getDownlinkPowerProfile :one
Select * from "DownlinkPowerProfile"
where "Name" like ?;

-- name: getSpexTxSubCarriersDetails :many
Select * from "SpecTxSubCarriers"
where "TxName" like ?;

-- name: getAllConverterNames :many
Select "Name" from "UpDownConverter";

-- name: getConverterDetails :one
Select * from "UpDownConverter"
where "Name" like ?;

-- name: getTxsubCarriers :one
Select * from "SpecTxSubCarriers"
where "TxName" like ? and "SubCarrierName" like ?;

-- name: getRxDetails :one
Select * from "SpecRx"
where "RxName" = ?;

-- name: getUplinkLoss :one
Select "Profile" from "UplinkLoss"
where "ConfigName" like ? and "TestPhaseName" like ?;

-- name: getPowerLevels :one
Select "PowerLevels" from "PowerProfile"
where "Name" like ?;

-- name: getRxTMTC :one
Select * from "SpecRxTMTC"
where "RxName" like ?;

-- name: getFrequencyProfile :one
Select * from "FrequencyProfile"
where "Name" like ?;

-- name: getCommandsInPowerProfile :one
Select "NoOfCommandsAtThreshold", "NoOfCommandsAtOtherLevels" from "PowerProfile"
where "Name" like ?;

-- name: getConfigsForUplink :many
Select "ConfigName" from "Tests"
Where "TestType" like "RFUplink";

-- name: getTRMParameters :one
Select * from "TRMProfile" where "Name" = ?;

-- name: getAllConfigsForDownlinkLoss :many
Select "ConfigName" from "DownlinkLoss"
where "TestPhaseName" like ?;

-- name: updateDownlinkLoss :exec
Update "DownlinkLoss" set "Profile"  = ?
where "ConfigName" like ? and "TestPhaseName" like ?;

-- name: getAllConfigsForUplinkLoss :many
Select "ConfigName" from "UplinkLoss"
where "TestPhaseName" like ?;

-- name: updateUplinkLoss :exec
Update "UplinkLoss" set "Profile"  = ?
where "ConfigName" like ? and "TestPhaseName" like ?;

-- name: getSpecTransponder :one
Select * from "SpecTpRanging" where "TpName" like ?;

-- name: getPulseSpecHRMode :one
Select * from "SpecPL" where "ConfigName" = ? and "ResolutionMode" = "HR";

-- name: getPulseSpecLRMode :one
Select * from "SpecPL" where "ConfigName" = ? and "ResolutionMode" = "LR";
