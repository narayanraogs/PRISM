package database

import "context"

func GetAllConverterNames() ([]string, error) {
	ctx := context.Background()
	names, err := dbObject.getAllConverterNames(ctx)
	return names, err
}

func GetConverterDetails(converter string) (UpDownConverter, error) {
	ctx := context.Background()
	details, err := dbObject.getConverterDetails(ctx, converter)
	return details, err
}
