package server

import (
	"prismServer/utils"
)

func handleSetPpmParameters(c *client, request actionRequest) (string, bool) {
	params := request.getParam("Parameters")
	if params == nil {
		return "Parameters not found in request", false
	}

	utils.SetSelectedPPMParameter(params)
	return "PPM parameters updated successfully", true
}

func handleSetVsaParameters(c *client, request actionRequest) (string, bool) {
	params := request.getParam("Parameters")
	if params == nil {
		return "Parameters not found in request", false
	}

	utils.SetSelectedVSAParameter(params)
	return "VSA parameters updated successfully", true
}
