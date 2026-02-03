package database

import "context"

func GetTMProfile(name string) (TMProfile, bool) {
	ctx := context.Background()
	tm, err := dbObject.getTMParameters(ctx, name)
	if err != nil {
		return TMProfile{}, false
	}
	return tm, true
}
