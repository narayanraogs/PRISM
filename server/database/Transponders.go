package database

import "context"

func GetTpSpecs(name string) (SpecTpRanging, bool) {
	ctx := context.Background()
	tp, err := dbObject.getSpecTransponder(ctx, name)
	if err != nil {
		return SpecTpRanging{}, false
	}
	return tp, true
}
