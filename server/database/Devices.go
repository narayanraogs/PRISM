package database

import "context"

func GetSAAndVSAList() ([]string, bool) {
	ctx := context.Background()
	vsa, err := dbObject.getAllSAAndVSA(ctx)
	if err != nil {
		return nil, false
	}
	return vsa, true
}

func GetVSAList() ([]string, bool) {
	ctx := context.Background()
	vsa, err := dbObject.getAllVSA(ctx)
	if err != nil {
		return nil, false
	}
	return vsa, true
}

func GetSAList() ([]string, bool) {
	ctx := context.Background()
	vsa, err := dbObject.getAllSA(ctx)
	if err != nil {
		return nil, false
	}
	return vsa, true
}

func GetPMAndPPMList() ([]string, bool) {
	ctx := context.Background()
	pm, err := dbObject.getAllPMAndPPM(ctx)
	if err != nil {
		return nil, false
	}
	return pm, true
}

func GetPPMList() ([]string, bool) {
	ctx := context.Background()
	pm, err := dbObject.getAllPPM(ctx)
	if err != nil {
		return nil, false
	}
	return pm, true
}

func GetPMList() ([]string, bool) {
	ctx := context.Background()
	pm, err := dbObject.getAllPM(ctx)
	if err != nil {
		return nil, false
	}
	return pm, true
}

func GetTSMList() ([]string, bool) {
	ctx := context.Background()
	tsm, err := dbObject.getAllTSM(ctx)
	if err != nil {
		return nil, false
	}
	return tsm, true
}

func GetSGList() ([]string, bool) {
	ctx := context.Background()
	tsm, err := dbObject.getAllSG(ctx)
	if err != nil {
		return nil, false
	}
	return tsm, true
}

func GetDeviceDetails(name string) (Device, bool) {
	ctx := context.Background()
	details, err := dbObject.getDeviceDetails(ctx, name)
	if err != nil {
		return Device{}, false
	}
	return details, true
}
