package driver

import (
	"prismServer/utils"
	"strings"
)

func getDeviceNotAvailable() utils.CommandResponse {
	return utils.CommandResponse{
		Success:      false,
		ErrorMessage: "Device Not Added for Operation",
	}
}

func getSuccessResponse() utils.CommandResponse {
	return utils.CommandResponse{
		Success: true,
		Result:  make(map[string]utils.CommandResult),
	}
}

func getErrorResponse(message string) utils.CommandResponse {
	return utils.CommandResponse{
		Success:      false,
		ErrorMessage: message,
	}
}

func getDriverStatus(driverStatus string) string {
	driverStatus = strings.TrimSpace(driverStatus)
	sts := strings.ToLower(driverStatus)
	temp := strings.Split(sts, "dr")
	required := temp[1 : utils.Config.TSM.NoOfDrivers+1]
	var tbr = make([]string, 0)
	for _, one := range required {
		tempSts := "D" + one
		tempSts = strings.ReplaceAll(tempSts, " on", "A")
		tempSts = strings.ReplaceAll(tempSts, " off", "B")
		tbr = append(tbr, tempSts)
	}
	return strings.Join(tbr, "!")
}
