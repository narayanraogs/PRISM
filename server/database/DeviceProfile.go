package database

import "context"

func GetDeviceProfileNames() ([]string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getAllDeviceProfiles(ctx)
	if err != nil {
		return nil, false
	}
	return dev, true
}

func GetPMFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getPMFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}

func GetSGFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getSGFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}

func GetVSAFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getVSAFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}

func GetPPMFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getPPMFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}

func GetTSMFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getTSMFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}

func GetSAFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getSAFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}

func GetGTxFromDeviceProfile(deviceProfile string) (string, bool) {
	ctx := context.Background()
	dev, err := dbObject.getGTxFromDeviceProfile(ctx, deviceProfile)
	if err != nil {
		return "", false
	}
	return dev.String, dev.Valid
}
