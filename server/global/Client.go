package global

import (
	"prismServer/tne"
	"prismServer/utilities"
	"sync"
)

type ClientGlobal struct {
	StringMap      sync.Map
	StringArrayMap sync.Map
	CLM            tne.CableLossMeasurement
	GTXAttn        tne.GTxAttnMeasurement
	SGPower        tne.SGPowerMeasurement
	TSMInternal    tne.TSMInternalLoss
	TCLM           utilities.TVACCableLossMeasurement
	TSMAttn        tne.TSMAttnMeasurement
	Stability      utilities.StabilityPlot
}

func (client *ClientGlobal) SetParameter(param string, value string) {
	client.StringMap.Store(param, value)
}

func (client *ClientGlobal) GetParam(param string) (string, bool) {
	val, ok := client.StringMap.Load(param)
	if !ok {
		return "", false
	}
	return val.(string), true
}

func (client *ClientGlobal) SetParameters(param string, value []string) {
	client.StringMap.Store(param, value)
}

func (client *ClientGlobal) GetParams(param string) ([]string, bool) {
	val, ok := client.StringMap.Load(param)
	if !ok {
		return nil, false
	}
	return val.([]string), true
}
