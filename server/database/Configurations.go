package database

import "context"

func GetPLConfigurations() ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getAllPayloadConfigs(ctx)
	if err != nil {
		return nil, false
	}
	return cfgs, true
}

func GetAllConfigurations() ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getAllConfigs(ctx)
	if err != nil {
		return nil, false
	}
	return cfgs, true
}

func GetAllConfigurationsWithTypes() ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getAllConfigsWithTypes(ctx)
	if err != nil {
		return nil, false
	}
	var tbr = make([]string, 0)
	for _, row := range cfgs {
		cfg := row.ConfigType + ";" + row.ConfigName
		tbr = append(tbr, cfg)
	}
	return tbr, true
}

func GetAllConfigsForTests() ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getAllConfigsForTests(ctx)
	if err != nil {
		return nil, false
	}
	var tbr = make([]string, 0)
	for _, row := range cfgs {
		cfg := row.ConfigType + ";" + row.ConfigName
		tbr = append(tbr, cfg)
	}
	return tbr, true
}

func GetConfigNamesForTSMConfig(tsmConfig string) ([]string, bool) {
	ctx := context.Background()
	cfgs, err := dbObject.getConfigurationNamesForTSMConfig(ctx, tsmConfig)
	if err != nil {
		return nil, false
	}
	return cfgs, true
}

func GetConfigurationDetails(configName string) (Configuration, bool) {
	ctx := context.Background()
	cfg, err := dbObject.getConfigurationDetails(ctx, configName)
	if err != nil {
		return Configuration{}, false
	}
	return cfg, true
}

func GetTSMUplinkPathForAllConfigs() (map[string]string, bool) {
	ctx := context.Background()
	var uplinkPaths = make(map[string]string)
	cfgs, err := dbObject.getAllConfigs(ctx)
	if err != nil {
		return nil, false
	}
	for _, configName := range cfgs {
		cfg, err := dbObject.getConfigurationDetails(ctx, configName)
		if err != nil {
			continue
		}
		tsm, err := dbObject.getAllPathsInTSMConfig(ctx, cfg.TSMConfigurationName)
		if err != nil {
			continue
		}
		uplinkPaths[configName] = tsm.UplinkToSC.String
	}
	return uplinkPaths, true
}
