package utils

import (
	"os"
	"path/filepath"
	"sync"
)

type applicationGlobal struct {
	StringMap  sync.Map
	IntegerMap sync.Map
	BooleanMap sync.Map
}

var g applicationGlobal

func GetSatelliteName() string {
	satName, ok := g.StringMap.Load("SatelliteName")
	if !ok {
		return ""
	} else {
		return satName.(string)
	}
}

func SetSatelliteName(value string) {
	g.StringMap.Store("SatelliteName", value)
}

func GetTestPhase() string {
	satName, ok := g.StringMap.Load("CurrentTestPhase")
	if !ok {
		return ""
	} else {
		return satName.(string)
	}
}

func SetTestPhase(value string) {
	g.StringMap.Store("CurrentTestPhase", value)
}

func GetTestResultDirectory() string {
	dir := filepath.Join(Config.BaseFolder, "results", "test", GetTestPhase())
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func GetTNEResultDirectory() string {
	dir := filepath.Join(Config.BaseFolder, "results", "tne", GetTestPhase())
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func GetCSVResultDirectory() string {
	dir := filepath.Join(Config.BaseFolder, "results", "csv", GetTestPhase())
	_ = os.MkdirAll(dir, 0755)
	return dir
}
