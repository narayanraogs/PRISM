package server

import (
	"fmt"
	"net/http"
	"prismServer/database"
	"prismServer/executeTest/measurements"
	"prismServer/resultsDB"
	"prismServer/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type valueHandler func(c *client) (values []string, ok bool)

var valueHandlers map[string]valueHandler

func init() {
	valueHandlers = map[string]valueHandler{
		"SatelliteName":              handleGetSatelliteName,
		"CurrentTestPhase":           handleGetCurrentTestPhase,
		"CableMeasurementCSV":        handleGetCableMeasurementCSV,
		"LossMeasurementFrequencies": handleGetLossMeasurementFrequencies,
		"SA":                         handleGetSA,
		"VSA":                        handleGetVSA,
		"PM":                         handleGetPM,
		"PPM":                        handleGetPPM,
		"SAOnly":                     handleGetSAOnly,
		"SpectrumProfiles":           handleGetSpectrumProfiles,
		"PulseProfiles":              handleGetPulseProfiles,
		"DeviceProfiles":             handleGetDeviceProfiles,
		"Receivers":                  handleGetReceivers,
		"PLConfigs":                  handleGetPLConfigs,
		"TSMConfigurations":          handleGetTSMConfigurations,
		"TestPhases":                 handleGetTestPhases,
		"Configs":                    handleGetConfigs,
		"TSM":                        handleGetTSM,
		"SG":                         handleGetSG,
		"CableNames":                 handleGetCableNames,
		"ConfigTypeAndNames":         handleGetConfigTypeAndNames,
		"ConfigNamesForTests":        handleGetConfigNamesForTests,
		"TestsForConfig":             handleGetTestsForConfig,
		"SpectrumProfile":            handleGetSpectrumProfile,
		"TSMPathForConfig":           handleGetTSMPathForConfig,
		"StabilityPlot":              handleGetStabilityPlot,
		"OfflineReport":              handleGetOfflineReport,
		"ConverterNames":             handleGetConverterNames,
		"ConfigsForUplink":           handleGetConfigsForUplink,
		"UplinkedConfigs":            handleGetUplinkedConfigs,
		"ConfigsForDownlink":         handleConfigsForDownlink,
		"DownlinkLossProfile":        handleDownlinkLossProfile,
		"ConfigsForUplinkLoss":       handleConfigsForUplinkLoss,
		"UplinkLossProfile":          handleUplinkLossProfile,
		"UplinkLossSummary":          handleUplinkLossSummary,
		"SANameFromDeviceProfile":    handleGetSANameInProfile,
		"GTxNameFromDeviceProfile":   handleGetGTxNameInProfile,
		"OfflineTestPhase":           handleGetOfflineTestPhases,
		"Spectrum":                   handleGetSpectrum,
		"AllPpmParameters":           handleGetAllPpmParameters,
		"SelectedPpmParameters":      handleGetSelectedPpmParameters,
		"AllVsaParameters":           handleGetAllVsaParameters,
		"SelectedVsaParameters":      handleGetSelectedVsaParameters,
	}
}

func handleGetAllPpmParameters(c *client) ([]string, bool) {
	return utils.GetAllPpmParameters(), true
}

func handleGetSelectedPpmParameters(c *client) ([]string, bool) {
	return utils.GetSelectedPPMParams(), true
}

func handleGetAllVsaParameters(c *client) ([]string, bool) {
	return utils.GetAllVsaParameters(), true
}

func handleGetSelectedVsaParameters(c *client) ([]string, bool) {
	return utils.GetSelectedVSAParams(), true
}

func getValues(c *gin.Context) {
	var request getRequest
	var response getResponse

	if err := c.BindJSON(&request); err != nil {
		response.OK = false
		response.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, response)
		return
	}

	s := sessions.getServer(request.ID)
	if s == nil {
		response.OK = false
		response.Message = "Client Not registered"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	response = processGetRequest(s, request)
	c.IndentedJSON(http.StatusOK, response)
}

func processGetRequest(c *client, request getRequest) getResponse {
	var response getResponse
	response.Values = make([]parameterValue, 0)
	response.OK = true

	for _, paramName := range request.Parameters {
		handler, ok := valueHandlers[paramName]
		if !ok {
			response.OK = false
			response.Message = fmt.Sprintf("Unknown parameter requested: %s", paramName)
			return response
		}

		values, success := handler(c)
		if !success {
			response.OK = false
			response.Message = fmt.Sprintf("Failed to get values for: %s", paramName)
			continue
		}

		p := parameterValue{
			Name:   paramName,
			Values: values,
		}
		response.Values = append(response.Values, p)
	}

	return response
}

func handleGetSatelliteName(c *client) ([]string, bool) {
	return []string{utils.GetSatelliteName()}, true
}

func handleGetCurrentTestPhase(c *client) ([]string, bool) {
	return []string{utils.GetTestPhase()}, true
}

func handleGetCableMeasurementCSV(c *client) ([]string, bool) {
	//losses, read := resultsDB.GetAllCableLosses()
	var csv strings.Builder
	/*if !read {
		return nil, false
	}

	for _, loss := range losses {
		csv.WriteString(strings.Join(loss, ","))
		csv.WriteString("\n")
	}*/
	return []string{csv.String()}, true
}

func handleGetLossMeasurementFrequencies(c *client) ([]string, bool) {
	return database.GetLossMeasurementFrequencies()
}

func handleGetSA(c *client) ([]string, bool) {
	return database.GetSAAndVSAList()
}

func handleGetVSA(c *client) ([]string, bool) {
	return database.GetVSAList()
}

func handleGetPM(c *client) ([]string, bool) {
	return database.GetPMAndPPMList()
}

func handleGetPPM(c *client) ([]string, bool) {
	return database.GetPPMList()
}

func handleGetSAOnly(c *client) ([]string, bool) {
	return database.GetSAList()
}

func handleGetSpectrumProfiles(c *client) ([]string, bool) {
	return database.GetAllSpectrumProfiles()
}

func handleGetPulseProfiles(c *client) ([]string, bool) {
	return database.GetPulseProfileNames()
}

func handleGetDeviceProfiles(c *client) ([]string, bool) {
	return database.GetDeviceProfileNames()
}

func handleGetReceivers(c *client) ([]string, bool) {
	return database.GetReceiverNames()
}

func handleGetPLConfigs(c *client) ([]string, bool) {
	return database.GetPLConfigurations()
}

func handleGetTSMConfigurations(c *client) ([]string, bool) {
	return database.GetTSMConfigurations()
}

func handleGetTestPhases(c *client) ([]string, bool) {
	return database.GetAllTestPhases()
}

func handleGetConfigs(c *client) ([]string, bool) {
	return database.GetAllConfigurations()
}

func handleGetTSM(c *client) ([]string, bool) {
	return database.GetTSMList()
}

func handleGetSG(c *client) ([]string, bool) {
	return database.GetSGList()
}

func handleGetCableNames(c *client) ([]string, bool) {
	return resultsDB.GetCableNames()
}

func handleGetConfigTypeAndNames(c *client) ([]string, bool) {
	return database.GetAllConfigurationsWithTypes()
}

func handleGetConfigNamesForTests(c *client) ([]string, bool) {
	return database.GetAllConfigsForTests()
}

func handleGetTestsForConfig(c *client) ([]string, bool) {
	config, read := c.global.GetParams("SelectedConfig")
	if !read {
		return nil, false
	}
	return database.GetTestsForConfig(config[0])
}

func handleGetSpectrumProfile(c *client) ([]string, bool) {
	profileName, read := c.global.GetParams("SpectrumProfile")
	if !read {
		return nil, false
	}
	profile, read := database.GetSpectrumProfile(profileName[0])
	if !read {
		return nil, false
	}
	array := []string{
		fmt.Sprintf("%.2f", profile.CenterFrequency),
		fmt.Sprintf("%.2f", profile.Span),
		fmt.Sprintf("%d", profile.RBW),
		fmt.Sprintf("%d", profile.VBW),
	}
	return array, true
}

func handleGetTSMPathForConfig(c *client) ([]string, bool) {
	cfgName, read := c.global.GetParams("SelectedConfiguration")
	if !read {
		return nil, false
	}
	return database.GetTSMPathsForConfig(cfgName[0])
}

func handleGetStabilityPlot(c *client) ([]string, bool) {
	msg, ok := c.global.Stability.Plot()
	if !ok {
		return nil, false
	}
	return []string{msg}, true
}

func handleGetOfflineReport(c *client) ([]string, bool) {
	date, readDate := c.global.GetParams("SelectedDate")
	time, readTime := c.global.GetParams("SelectedTime")
	if !readDate || !readTime {
		return nil, false
	}
	report, err := resultsDB.GetReportPDF(date[0], time[0])
	if err != nil {
		return nil, false
	}
	return []string{report}, true
}

func handleGetConverterNames(c *client) ([]string, bool) {
	conv, err := database.GetAllConverterNames()
	return conv, err != nil
}

func handleGetConfigsForUplink(c *client) ([]string, bool) {
	return database.GetConfigsForUplink()
}

func handleGetUplinkedConfigs(c *client) ([]string, bool) {
	tsms, ok := database.GetTSMList()
	if !ok {
		return nil, false
	}
	uplinks := measurements.GetAllActiveUplinkConfigs(tsms[0])
	return uplinks, true
}

func handleConfigsForDownlink(c *client) ([]string, bool) {
	testPhase, ok := c.global.GetParams("SelectedTestPhase")
	if !ok {
		return testPhase, false
	}
	return database.GetAllConfigsForDownlink(testPhase[0])
}

func handleConfigsForUplinkLoss(c *client) ([]string, bool) {
	testPhase, ok := c.global.GetParams("SelectedTestPhase")
	if !ok {
		return testPhase, false
	}
	return database.GetAllConfigsForUplinkLoss(testPhase[0])
}

func handleDownlinkLossProfile(c *client) ([]string, bool) {
	testPhase, ok := c.global.GetParams("SelectedTestPhase")
	if !ok {
		return nil, false
	}
	cfg, ok := c.global.GetParams("SelectedConfiguration")
	if !ok {
		return nil, false
	}
	loss, ok := database.GetDownlinkLossProfile(cfg[0], testPhase[0])
	if !ok {
		return nil, false
	}
	return []string{loss}, true
}

func handleUplinkLossProfile(c *client) ([]string, bool) {
	testPhase, ok := c.global.GetParams("SelectedTestPhase")
	if !ok {
		return nil, false
	}
	cfg, ok := c.global.GetParams("SelectedConfiguration")
	if !ok {
		return nil, false
	}
	loss, ok := database.GetUplinkLossProfile(cfg[0], testPhase[0])
	if !ok {
		return nil, false
	}
	return []string{loss}, true
}

func handleUplinkLossSummary(c *client) ([]string, bool) {
	tp, _ := database.GetSelectedTestPhase()
	cfg, ok := c.global.GetParams("SelectedConfiguration")
	if !ok {
		return nil, false
	}
	_, sa, _, sc, ok := database.GetCurrentUplinkLoss(cfg[0], tp)
	if !ok {
		return nil, false
	}
	return []string{fmt.Sprintf("%.2f", sa), fmt.Sprintf("%.2f", sc)}, true
}

func handleGetSANameInProfile(c *client) ([]string, bool) {
	deviceProfile, ok := c.global.GetParams("SelectedDeviceProfile")
	if !ok {
		return nil, false
	}
	sa, ok := database.GetSAFromDeviceProfile(deviceProfile[0])
	if !ok {
		return nil, false
	}
	return []string{sa}, true
}

func handleGetGTxNameInProfile(c *client) ([]string, bool) {
	deviceProfile, ok := c.global.GetParams("SelectedDeviceProfile")
	if !ok {
		return nil, false
	}
	gtx, ok := database.GetGTxFromDeviceProfile(deviceProfile[0])
	if !ok {
		return nil, false
	}
	return []string{gtx}, true
}

func handleGetOfflineTestPhases(c *client) ([]string, bool) {
	tp, err := resultsDB.GetOfflineTestPhases()
	if err != nil {
		return nil, false
	}
	return tp, true
}

func handleGetSpectrum(c *client) ([]string, bool) {
	sas, ok := c.global.GetParams("SelectedSA")
	if !ok {
		return nil, false
	}
	return getSpectrum(c, sas[0])
}
