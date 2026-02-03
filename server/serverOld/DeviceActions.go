package server

import (
	"prismServer/driver"
	"strconv"
)

func handleSetTSMRoute(c *client, request actionRequest) (string, bool) {
	tsmNames := request.getParam("SelectedTSM")
	routes := request.getParam("RouteTSM")
	if tsmNames == nil || routes == nil {
		return "Required Parameters are not set", false
	}
	var tsm driver.TSM
	ok := tsm.LoadDevice(tsmNames[0])
	if !ok {
		return "Selected TSM doesn't exist in database", false
	}
	response := tsm.SetDriverStatus(routes[0])
	if !response.Success {
		return "Error while Routing", false
	}
	return "", true
}

func handleSetTSMAttn(c *client, request actionRequest) (string, bool) {
	tsmNames := request.getParam("SelectedTSM")
	attn := request.getParam("Attn")
	if tsmNames == nil || attn == nil {
		return "Required Parameters are not set", false
	}
	var tsm driver.TSM
	ok := tsm.LoadDevice(tsmNames[0])
	if !ok {
		return "Selected TSM doesn't exist in database", false
	}
	attnNo, err := strconv.Atoi(attn[0])
	if err != nil {
		return "Attn No is not an Integer", false
	}
	attnValue, err := strconv.ParseFloat(attn[1], 64)
	if err != nil {
		return "Attn Value is not a floating point number", false
	}
	response := tsm.SetAttn(attnNo, attnValue)
	if !response.Success {
		return "Error while Routing", false
	}
	return "", true
}

func handleSetSpectrum(c *client, request actionRequest) (string, bool) {
	sas := request.getParam("SelectedSA")
	details := request.getParam("Details")
	if sas == nil || details == nil {
		return "Required Parameters not set", false
	}
	return setSpectrum(c, sas[0], details)
}
