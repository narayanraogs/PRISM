package database

import "context"

func GetTSMConfigurations() ([]string, bool) {
	ctx := context.Background()
	tsm, err := dbObject.getAllTSMConfigurations(ctx)
	if err != nil {
		return nil, false
	}
	return tsm, true
}

func GetTSMPathsForConfig(name string) ([]string, bool) {
	ctx := context.Background()
	tsmCfg, err := dbObject.getTSMConfigNameForConfig(ctx, name)
	if err != nil {
		return nil, false
	}
	tsm, err := dbObject.getAllPathsInTSMConfig(ctx, tsmCfg)
	if err != nil {
		return nil, false
	}
	var paths = make([]string, 0)
	paths = append(paths, "UplinkToSC;"+tsm.UplinkToSC.String)
	paths = append(paths, "IncludePad;"+tsm.IncludePad.String)
	paths = append(paths, "ExcludePad;"+tsm.ExcludePad.String)
	paths = append(paths, "UplinkToSA;"+tsm.UplinkToSA.String)
	paths = append(paths, "UplinkToPM;"+tsm.UplinkToPM.String)
	paths = append(paths, "TerminateUplink;"+tsm.TerminateUplink.String)
	paths = append(paths, "DownlinkToSA;"+tsm.DownlinkToSA.String)
	paths = append(paths, "DownlinkToPM;"+tsm.DownlinkToPM.String)
	paths = append(paths, "DownlinkToDC;"+tsm.DownlinkToDC.String)
	return paths, true
}

func GetTSMPathDetails(name string) (TSMConfiguration, bool) {
	ctx := context.Background()
	tsm, err := dbObject.getAllPathsInTSMConfig(ctx, name)
	if err != nil {
		return tsm, false
	}
	return tsm, true
}
