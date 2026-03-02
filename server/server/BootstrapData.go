package server

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/resultsDB"
	"prismServer/tne"
	"prismServer/utils"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func getBootstrapData(c *gin.Context) {
	var resp BootstrapData
	tp, _ := database.GetSelectedTestPhase()

	resp.RFUplinkData = getRFUplinkMetadata(tp)
	resp.TestData = getTestMetadata(tp)
	resp.StabilityData = getStabilityMetadata()
	resp.StabilityReportsData = getStabilityReportsMetadata()
	resp.SpectrumDumpData = getSpectrumDumpMetadata()
	resp.MonitorData = getMonitorMetadata()
	resp.TVACCableLossData = getTVACCableLossInitialData()
	resp.CableLossData = getCableLossInitialData()
	resp.DatabaseData = getDatabaseMetadata()
	resp.ReportsData = getResultInitialData()
	resp.TSMInternalLossData = getTSMInternalLossMetadata()
	resp.UCDCData = getUCDCMetadata()
	resp.AttnData = getAttnInitialData()
	resp.GTxData = getGTxMeasurementMetadata()
	resp.SCPIData = getSCPIData()
	resp.PlannerData = getPlannerData()

	c.IndentedJSON(http.StatusOK, resp)
}

func getRFUplinkMetadata(tp string) RFUplinkMetaData {
	var ok bool
	var rfm RFUplinkMetaData
	rfm.TSMs, ok = database.GetTSMList()
	if !ok {
		rfm.TSMs = []string{}
	}

	rfm.AllConfigs, ok = database.GetAllConfigurations()
	if !ok {
		rfm.AllConfigs = []string{}
	}
	rfm.ConfigPathInformation = make(map[string][]ConfigPathInformation)
	for _, config := range rfm.AllConfigs {
		rfm.ConfigPathInformation[config] = []ConfigPathInformation{}
		paths, ok := database.GetTSMPathsForConfig(config)
		if !ok {
			continue
		}
		for _, path := range paths {
			temp := strings.Split(path, ";")
			if strings.EqualFold(strings.TrimSpace(temp[1]), "") {
				continue
			}
			rfm.ConfigPathInformation[config] = append(rfm.ConfigPathInformation[config], ConfigPathInformation{
				Path:     temp[0],
				Mnemonic: temp[1],
			})
		}
	}
	rfm.UplinkConfigs, ok = database.GetConfigsForUplink()
	if !ok {
		rfm.UplinkConfigs = []string{}
	}
	rfm.UplinkConfigInformation = make(map[string]UplinkConfigInformation)
	for _, config := range rfm.UplinkConfigs {
		_, sa, _, sc, ok := database.GetCurrentUplinkLoss(config, tp)
		if !ok {
			continue
		}
		var ul = UplinkConfigInformation{
			SALoss:    sa,
			SCLoss:    sc,
			PowerAtSC: -90,
		}

		rfm.UplinkConfigInformation[config] = ul
	}
	return rfm
}

func getTestMetadata(tp string) AllTests {
	var resp AllTests
	resp.Categories = make([]string, 0)
	resp.Configurations = make(map[string][]string)
	resp.Tests = make(map[string][]TestDescription)
	resp.Losses = make(map[string]string)
	configs, ok := database.GetAllConfigsForTests()
	var configNames = make([]string, 0)
	if !ok {
		resp.OK = false
		resp.Message = "Not able to get Details from Database"
		return resp
	}

	for _, config := range configs {
		temp := strings.Split(config, ";")
		if slices.Index(resp.Categories, temp[0]) == -1 {
			resp.Categories = append(resp.Categories, temp[0])
			resp.Configurations[temp[0]] = make([]string, 0)
		}
		configNames = append(configNames, temp[1])
		resp.Configurations[temp[0]] = append(resp.Configurations[temp[0]], temp[1])
		if strings.EqualFold(temp[0], "rx") {
			_, sa, _, sc, ok := database.GetCurrentUplinkLoss(temp[1], tp)
			if ok {
				resp.Losses[temp[1]] = fmt.Sprintf("SA: %.2f, SC: %.2f", sa, sc)
			} else {
				resp.Losses[temp[1]] = ""
			}
		} else {
			_, sa, pm, ok := database.GetCurrentDownlinkLoss(temp[1], tp)
			if ok {
				resp.Losses[temp[1]] = fmt.Sprintf("SA: %.2f, PM: %.2f", sa, pm)
			} else {
				resp.Losses[temp[1]] = ""
			}
		}
	}

	for _, config := range configNames {
		resp.Tests[config] = make([]TestDescription, 0)
		tests, ok := database.GetTestsForConfig(config)
		if !ok {
			continue
		}
		for _, test := range tests {
			temp := strings.Split(test, ";")
			var t TestDescription
			if len(temp) == 2 {
				t.TestName = temp[0]
				t.TestCategory = temp[1]
			} else {
				t.TestName = temp[0]
				t.TestCategory = ""
			}
			resp.Tests[config] = append(resp.Tests[config], t)
		}
	}
	return resp
}

func getStabilityMetadata() StabilityMetadata {
	var stb StabilityMetadata
	stb.InstrumentTypes = []string{"SA", "PM", "PPM", "TM"}
	stb.Instruments = make(map[string][]string)
	stb.Parameters = make(map[string][]string)
	sa, ok := database.GetSAAndVSAList()
	if !ok {
		stb.OK = false
		stb.Message = "SA's not present in Database"
		return stb
	}
	stb.Instruments["SA"] = sa
	pm, ok := database.GetPMAndPPMList()
	if !ok {
		stb.OK = false
		stb.Message = "PM's not present in Database"
		return stb
	}
	stb.Instruments["PM"] = pm
	ppm, ok := database.GetPPMList()
	if ok {
		//PPM is optional
		stb.Instruments["PPM"] = ppm
	}
	stb.Instruments["TM"] = []string{"TM1", "TM2", "ANY"}
	stb.Parameters["SA"] = []string{"Power", "Frequency", "Trace"}
	stb.Parameters["PM"] = []string{"Channel A", "Channel B"}
	stb.Parameters["PPM"] = []string{"Peak Power", "Average Power", "Pulse Width", "Pulse Period"}
	stb.Parameters["TM"] = []string{"Processed", "Raw"}
	stb.PPMChannels = []string{"A", "B"}
	stb.Profiles = make([]SpectrumProfile, 0)
	sps, ok := database.GetAllSpectrumProfiles()
	if !ok {
		stb.OK = false
		stb.Message = "Cannot get Spectrum Profiles from Database"
		return stb
	}
	for _, profile := range sps {
		spec, ok := database.GetSpectrumProfile(profile)
		if !ok {
			continue
		}
		var prof SpectrumProfile
		prof.ProfileName = spec.Name
		prof.CenterFrequency = spec.CenterFrequency
		prof.Span = spec.Span
		prof.RBW = float64(spec.RBW)
		prof.VBW = float64(spec.VBW)
		stb.Profiles = append(stb.Profiles, prof)
	}
	stb.PLConfigs, ok = database.GetPLConfigurations()
	if !ok {
		stb.OK = false
		stb.Message = "PL Configurations not found"
		return stb
	}
	stb.PulseProfiles, ok = database.GetPulseProfileNames()
	if !ok {
		stb.OK = false
		stb.Message = "Pulse Profiles not found"
		return stb
	}
	stb.OK = true
	stb.Message = "Success"
	return stb
}

func getStabilityReportsMetadata() StabilityReportsMetadata {
	rows, err := resultsDB.GetStabilitySessions()
	var resp StabilityReportsMetadata
	resp.ID = make([]int64, 0)
	resp.Date = make([]string, 0)
	resp.Time = make([]string, 0)
	resp.Parameters = make([][]string, 0)
	if err != nil {
		resp.OK = false
		resp.Message = "Error getting stability reports metadata"
		return resp
	}
	for _, row := range rows {
		params, _ := resultsDB.GetStabilityParameters(row.ID)
		if len(params) > 0 {
			resp.ID = append(resp.ID, row.ID)
			resp.Date = append(resp.Date, row.Date)
			resp.Time = append(resp.Time, row.Time)
			resp.Parameters = append(resp.Parameters, params)
		}
	}
	resp.OK = true
	resp.Message = "Success"
	return resp
}

func getSpectrumDumpMetadata() SpectrumDumpMetadata {
	var stb SpectrumDumpMetadata
	stb.SpectrumDumpMode = []string{"Spectrum Dump", "Screenshot"}
	stb.Instruments = make(map[string][]string)
	stb.SpectrumProfiles = make([]SpectrumProfile, 0)
	sa, ok := database.GetSAAndVSAList()
	if !ok {
		stb.OK = false
		stb.Message = "SA's not present in Database"
		return stb
	}
	stb.Instruments["SA"] = sa
	vsa, ok := database.GetVSAList()
	if !ok {
		stb.OK = false
		stb.Message = "VSA's not present in Database"
		return stb
	}
	stb.Instruments["VSA"] = vsa

	sps, ok := database.GetAllSpectrumProfiles()
	if !ok {
		stb.OK = false
		stb.Message = "Cannot get Spectrum Profiles from Database"
		return stb
	}
	for _, profile := range sps {
		spec, ok := database.GetSpectrumProfile(profile)
		if !ok {
			continue
		}
		var prof SpectrumProfile
		prof.ProfileName = spec.Name
		prof.CenterFrequency = spec.CenterFrequency
		prof.Span = spec.Span
		prof.RBW = float64(spec.RBW)
		prof.VBW = float64(spec.VBW)
		stb.SpectrumProfiles = append(stb.SpectrumProfiles, prof)
	}

	stb.ScreenshotProfiles = []string{"Screenshot", "Magniture", "Pulse Magniture",
		"Pulse Frequency", "Pulse Phase", "Spectrogram"}

	stb.OK = true
	stb.Message = "Success"
	return stb
}

func getMonitorMetadata() MonitorMetadata {
	var mmd MonitorMetadata
	mmd.InstrumentTypes = []string{"SA", "VSA", "PM", "PPM"}
	mmd.Instruments = make(map[string][]string)
	sa, ok := database.GetSAAndVSAList()
	if !ok {
		mmd.OK = false
		mmd.Message = "SA's not present in Database"
		return mmd
	}
	mmd.Instruments["SA"] = sa
	pm, ok := database.GetPMAndPPMList()
	if !ok {
		mmd.OK = false
		mmd.Message = "PM's not present in Database"
		return mmd
	}
	mmd.Instruments["PM"] = pm
	ppm, ok := database.GetPPMList()
	if ok {
		//PPM is optional
		mmd.Instruments["PPM"] = ppm
	}
	vsa, ok := database.GetVSAList()
	if ok {
		//VSA is optional
		mmd.Instruments["VSA"] = vsa
	}
	mmd.OK = true
	mmd.Message = "Success"
	return mmd
}

func getTVACCableLossInitialData() TVACCableLossMetadata {
	var resp TVACCableLossMetadata
	resp.Frequencies = make([]float64, 0)
	resp.IsPMZeroed = resultsDB.CheckIfTVACCableLossPMReferenceExists()
	names, ok := database.GetLossMeasurementFrequencyNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get frequencies from LossMeasurementFrequencies"
		return resp
	}
	for _, name := range names {
		frequency, ok := database.GetFrequencyForLossMeasurement(name)
		if !ok {
			resp.OK = false
			resp.Message = "Unable to get frequency for " + name
			return resp
		}
		resp.Frequencies = append(resp.Frequencies, frequency)
	}
	deviceProfile, ok := database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get device profile names"
		return resp
	}
	resp.DeviceProfiles = deviceProfile
	cableNames, ok := resultsDB.GetCableNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get cable names"
		return resp
	}
	resp.ExistingCables = cableNames
	resp.OK = true
	resp.Message = "Success"
	return resp
}

func getCableLossInitialData() CableLossMetadata {
	var resp CableLossMetadata
	resp.Frequencies = make([]string, 0)
	resp.IsPMZeroed = resultsDB.CheckIfCableLossPMReferenceExists()
	names, ok := database.GetLossMeasurementFrequencies()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get frequencies from LossMeasurementFrequencies"
		return resp
	}
	resp.Frequencies = names
	deviceProfile, ok := database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get device profile names"
		return resp
	}
	resp.DeviceProfiles = deviceProfile
	cableNames, ok := resultsDB.GetCableNamesForCableLoss()
	if !ok {
		resp.OK = false
		resp.Message = "Unable to get cable names"
		return resp
	}
	resp.ExistingCables = cableNames
	resp.OK = true
	resp.Message = "Success"
	return resp
}

func getDatabaseMetadata() DatabaseMetadata {
	tp, ok := database.GetAllTestPhases()
	if !ok {
		return DatabaseMetadata{OK: false, Message: "Failed to fetch test phases"}
	}
	return DatabaseMetadata{
		TestPhases: tp,
		OK:         true,
	}
}

func getResultInitialData() ReportsResponse {
	var resp ReportsResponse
	results, err := resultsDB.GetAllResults()
	if err != nil {
		resp.OK = false
		resp.Message = "Unable to get results"
		return resp
	}
	for _, result := range results {
		resp.Reports = append(resp.Reports, ReportMetadata{
			Date:         result.Date,
			Time:         result.Time,
			TestType:     result.TestType,
			Config:       result.ConfigName,
			TestCategory: result.TestCategory.String,
			Phase:        result.TestPhase,
			Remarks:      result.Remark.String,
			VSAUsed:      strings.EqualFold(result.TestCategory.String, "vsa"),
			PPMUsed:      strings.EqualFold(result.TestCategory.String, "ppm"),
		})
	}
	resp.AllPPMParams = utils.GetAllPpmParameters()
	resp.SelectedPPMParams = utils.GetSelectedPPMParams()
	resp.AllVSAParams = utils.GetAllVsaParameters()
	resp.SelectedVSAParams = utils.GetSelectedVSAParams()
	resp.OK = true
	resp.Message = "Success"

	return resp
}

func getTSMInternalLossMetadata() TSMInternalLossMetadata {
	var resp TSMInternalLossMetadata
	var ok bool
	resp.DeviceProfile, ok = database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get device profile names"
		return resp
	}
	resp.MeasuredLoss, ok = getTSMLossTable()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get read Loss table from database"
		return resp
	}
	resp.OK = true
	resp.Message = "Success"
	return resp
}

func getGTxMeasurementMetadata() GTxMeasurementMetadata {
	var resp GTxMeasurementMetadata
	profiles, ok := database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get device profiles"
		return resp
	}
	resp.DeviceProfile = profiles
	resp.DeviceMapping = make(map[string]DeviceProfileDetails)
	for _, profile := range profiles {
		var device DeviceProfileDetails
		gtxName, _ := database.GetGTxFromDeviceProfile(profile)
		device.GTxName = gtxName
		saName, _ := database.GetSAFromDeviceProfile(profile)
		device.SAName = saName
		VSAName, _ := database.GetVSAFromDeviceProfile(profile)
		device.VSAName = VSAName
		tsmName, _ := database.GetTSMFromDeviceProfile(profile)
		device.TSMName = tsmName
		pmName, _ := database.GetPMFromDeviceProfile(profile)
		device.PMName = pmName
		ppmName, _ := database.GetPPMFromDeviceProfile(profile)
		device.PPMName = ppmName
		sgName, _ := database.GetSGFromDeviceProfile(profile)
		device.SGName = sgName
		resp.DeviceMapping[profile] = device
	}
	resp.OK = true
	resp.Message = "Success"
	return resp
}

func getUCDCMetadata() UCDCMetadata {
	var resp UCDCMetadata
	profiles, ok := database.GetDeviceProfileNames()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get device profiles"
		return resp
	}
	resp.DeviceProfiles = profiles
	resp.DeviceMapping = make(map[string]DeviceProfileDetails)
	for _, profile := range profiles {
		var device DeviceProfileDetails
		gtxName, _ := database.GetGTxFromDeviceProfile(profile)
		device.GTxName = gtxName
		saName, _ := database.GetSAFromDeviceProfile(profile)
		device.SAName = saName
		VSAName, _ := database.GetVSAFromDeviceProfile(profile)
		device.VSAName = VSAName
		tsmName, _ := database.GetTSMFromDeviceProfile(profile)
		device.TSMName = tsmName
		pmName, _ := database.GetPMFromDeviceProfile(profile)
		device.PMName = pmName
		ppmName, _ := database.GetPPMFromDeviceProfile(profile)
		device.PPMName = ppmName
		sgName, _ := database.GetSGFromDeviceProfile(profile)
		device.SGName = sgName
		resp.DeviceMapping[profile] = device
	}
	converters, err := database.GetAllConverterNames()
	if err != nil {
		resp.OK = false
		resp.Message = "Failed to get converters"
		return resp
	}
	resp.Converters = converters
	resp.ConverterDetails = make(map[string]UCDCDetails)
	for _, converter := range converters {
		var converterDetails UCDCDetails
		details, err := database.GetConverterDetails(converter)
		if err != nil {
			resp.OK = false
			resp.Message = "Failed to get converter details"
			return resp
		}
		converterDetails.InputFrequency = details.InputFrequency
		converterDetails.OutputFrequency = details.OutputFrequency
		converterDetails.LOFrequency = math.Abs(details.InputFrequency - details.OutputFrequency)
		resp.ConverterDetails[converter] = converterDetails
	}
	sg, ok := database.GetSGList()
	if !ok {
		resp.OK = false
		resp.Message = "Failed to get SG list"
		return resp
	}
	resp.SignalGenerators = sg
	resp.AvailableTests = []UCDCTestMetadata{
		{Code: tne.UCDCGainInternalCable, DisplayName: "Gain Measurement - Internal LO (Cable)", Category: "Output Port"},
		{Code: tne.UCDCGainInternalRadiated, DisplayName: "Gain Measurement - Internal LO (Radiated)", Category: "Output Port"},
		{Code: tne.UCDCFreqMeas, DisplayName: "Frequency Measurement", Category: "Output Port"},
		{Code: tne.UCDCHarmonicMeas, DisplayName: "Harmonics Measurement", Category: "Output Port"},
		{Code: tne.UCDCSpuriousInBand, DisplayName: "Spurious - In-Band", Category: "Output Port"},
		{Code: tne.UCDCSpuriousOutBand, DisplayName: "Spurious - Out of Band", Category: "Output Port"},
		{Code: tne.UCDCLOLeakage, DisplayName: "LO Leakage", Category: "Output Port"},
		{Code: tne.UCDCInputLeakage, DisplayName: "Input Leakage", Category: "Output Port"},
		{Code: tne.UCDCGainExternalCable, DisplayName: "Gain Measurement - External LO (Cable)", Category: "Output Port"},
		{Code: tne.UCDCGainExternalRadiated, DisplayName: "Gain Measurement - External LO (Radiated)", Category: "Output Port"},
		{Code: tne.UCDCExtLOPowerMatch, DisplayName: "External LO Power Matching", Category: "LO Monitor"},

		{Code: tne.UCDCLOLeakage, DisplayName: "LO Leakage", Category: "Input Port"},

		{Code: tne.UCDCOutputMonPower, DisplayName: "Power Measurement", Category: "Output Monitor"},

		{Code: tne.UCDCInputMonPower, DisplayName: "Power Measurement", Category: "Input Monitor"},

		{Code: tne.UCDCLOMonPower, DisplayName: "Power & Frequency Measurement", Category: "LO Monitor"},
		{Code: tne.UCDCLOMonPhaseNoise, DisplayName: "Phase Noise Measurement", Category: "LO Monitor"},
	}
	resp.OK = true
	resp.Message = "Successfully got UCDC metadata"
	return resp
}

func getAttnInitialData() AttnMetaData {
	var attnMetaData AttnMetaData
	deviceProfiles, ok := database.GetDeviceProfileNames()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get device profiles"
		return attnMetaData
	}
	attnMetaData.DeviceProfile = deviceProfiles
	rxNames, ok := database.GetReceiverNames()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get receiver names"
		return attnMetaData
	}
	attnMetaData.Receiver = rxNames
	spectrumProfileNames, ok := database.GetAllSpectrumProfiles()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get spectrum profile names"
		return attnMetaData
	}
	attnMetaData.SprectrumProfile = spectrumProfileNames
	tsmConfigNames, ok := database.GetTSMConfigurations()
	if !ok {
		attnMetaData.OK = false
		attnMetaData.Message = "Unable to get TSM config names"
		return attnMetaData
	}
	attnMetaData.TSMConfig = tsmConfigNames
	attnMetaData.GTxComponents = []string{"IFM-1", "IFM-2"}
	attnMetaData.AttnRanges = map[string]AttnRange{
		"TSM": {Max: 63.5, Min: 0, StepSize: 0.1},
		"GTx": {Max: 0, Min: -50, StepSize: 0.5},
		"SG":  {Max: 0, Min: -80, StepSize: 1},
	}
	attnMetaData.OK = true
	attnMetaData.Message = "Successfully retrieved metadata"
	return attnMetaData
}

func getSCPIData() SCPIDetails {
	var scpiDetails SCPIDetails
	var instruments = make([]string, 0)
	sa, ok := database.GetSAAndVSAList()
	if !ok {
		scpiDetails.OK = false
		scpiDetails.Message = "Unable to get SA and VSA list"
		return scpiDetails
	}
	instruments = append(instruments, sa...)
	pm, ok := database.GetPMAndPPMList()
	if !ok {
		scpiDetails.OK = false
		scpiDetails.Message = "Unable to get PM and PPM list"
		return scpiDetails
	}
	instruments = append(instruments, pm...)
	sg, ok := database.GetSGList()
	if !ok {
		scpiDetails.OK = false
		scpiDetails.Message = "Unable to get SG list"
		return scpiDetails
	}
	instruments = append(instruments, sg...)
	scpiDetails.Instruments = instruments
	scpiDetails.InstrumentDetails = make(map[string]InstrumentDetails)
	for _, instrument := range instruments {
		var details InstrumentDetails
		dev, ok := database.GetDeviceDetails(instrument)
		if !ok {
			scpiDetails.OK = false
			scpiDetails.Message = "Unable to get IP address and Port No for " + instrument
			return scpiDetails
		}
		details.IPAddress = dev.IPAddress
		details.PortNo = int(dev.ControlPort)
		scpiDetails.InstrumentDetails[instrument] = details
	}
	scpiDetails.Commands = make(map[string][]CommandDetails)
	for _, instrument := range sa {
		var sa driver.SA
		sa.LoadDevice(instrument)
		commands := sa.GetCommandDatabase()
		for _, command := range commands {
			scpiDetails.Commands[instrument] = append(scpiDetails.Commands[instrument], CommandDetails{
				Command:  command.Command,
				Mnemonic: command.Mnemonic,
				Argument: command.Argument,
				Write:    !command.Read,
			})
		}
	}
	for _, instrument := range pm {
		var pm driver.PM
		pm.LoadDevice(instrument)
		commands := pm.GetCommandDatabase()
		for _, command := range commands {
			scpiDetails.Commands[instrument] = append(scpiDetails.Commands[instrument], CommandDetails{
				Command:  command.Command,
				Mnemonic: command.Mnemonic,
				Argument: command.Argument,
				Write:    !command.Read,
			})
		}
	}
	for _, instrument := range sg {
		var sg driver.SG
		sg.LoadDevice(instrument)
		commands := sg.GetCommandDatabase()
		for _, command := range commands {
			scpiDetails.Commands[instrument] = append(scpiDetails.Commands[instrument], CommandDetails{
				Command:  command.Command,
				Mnemonic: command.Mnemonic,
				Argument: command.Argument,
				Write:    !command.Read,
			})
		}
	}
	scpiDetails.OK = true
	scpiDetails.Message = "Successfully retrieved metadata"
	return scpiDetails
}

func getPlannerData() string {
	plannerPath := filepath.Join(utils.Config.BaseFolder, ".resources", "plannerState.json")
	data, err := os.ReadFile(plannerPath)
	if err != nil {
		return ""
	}
	return string(data)
}

func savePlannerData(c *gin.Context) {
	var req struct {
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "Invalid request"})
		return
	}
	plannerPath := filepath.Join(utils.Config.BaseFolder, ".resources", "plannerState.json")
	err := os.WriteFile(plannerPath, []byte(req.Data), 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": "Failed to save data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Saved successfully"})
}
