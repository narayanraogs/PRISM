package database

import (
	"context"
)

func GetPulseSpecifications(name string) (SpecPL, bool) {
	ctx := context.Background()
	plSpecs, err := dbObject.getFullPulseSpecs(ctx, name)
	if err != nil {
		return SpecPL{}, false
	}
	return plSpecs, true
}
