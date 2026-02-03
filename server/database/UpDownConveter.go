package database

import "context"

func GetAllConverterNames() ([]string,bool){
	ctx := context.Background()
	names, err := dbObject.getAllConverterNames(ctx)
	return names, err == nil
}