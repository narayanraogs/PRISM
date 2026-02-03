package remote

import "prismServer/database"

func GetInfo() SoftwareResponse {
	var tbr SoftwareResponse
	tbr.Getters = append(tbr.Getters, getGetters()...)
	tbr.Setters = append(tbr.Setters, getSetters()...)
	tbr.Actions = append(tbr.Actions, getActions()...)
	tbr.Ack = Acknowledgement{
		Status: true,
		Msg:    "",
	}
	return tbr
}

func getGetters() []string {
	var tbr = make([]string, 0)
	tbr = append(tbr, "TestPhase", "Configurations", "UplinkConfigurations", "SA", "PM", "TSM", "SG", "GTx", "PPM", "VSA", "VSG", "DeviceProfile")
	return tbr
}

func getSetters() []Setter {
	var tbr = make([]Setter, 0)
	saNames, _ := database.GetSAAndVSAList()
	vsaNames, _ := database.GetVSAList()
	pmNames, _ := database.GetPMAndPPMList()
	ppmNames, _ := database.GetPPMList()
	devProfiles, _ := database.GetDeviceProfileNames()
	tsmNames, _ := database.GetTSMList()
	configs, _ := database.GetAllConfigurations()
	tbr = append(tbr, Setter{ParamName: "SA", Values: saNames})
	tbr = append(tbr, Setter{ParamName: "VSA", Values: vsaNames})
	tbr = append(tbr, Setter{ParamName: "PM", Values: pmNames})
	tbr = append(tbr, Setter{ParamName: "PPM", Values: ppmNames})
	tbr = append(tbr, Setter{ParamName: "DeviceProfile", Values: devProfiles})
	tbr = append(tbr, Setter{ParamName: "TSM", Values: tsmNames})
	tbr = append(tbr, Setter{ParamName: "Config", Values: configs})
	tbr = append(tbr, Setter{ParamName: "TSMPath", Values: []string{}})
	return tbr
}

func getActions() []Action {
	var tbr = make([]Action, 0)
	tbr = append(tbr, Action{Type: "Screenshot", ParamNames: []string{"SA"}})
	tbr = append(tbr, Action{Type: "RouteTSM", ParamNames: []string{"TSM", "TSMPath"}})
	return tbr
}
