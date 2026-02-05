package server

import (
	"fmt"
	"prismServer/executeTest"
	"prismServer/tne"
	"prismServer/utilities"
	"strconv"
	"strings"
)

func handleMeasureCableLoss(c *client, request actionRequest) (string, bool) {
	deviceProfiles := request.getParam("SelectedDeviceProfile")
	pmChannels := request.getParam("SelectedPMChannel")
	measurementTypes := request.getParam("MeasurementType")
	cableNames := request.getParam("CableName")
	cableLengths := request.getParam("CableLength")
	frequencies := request.getParam("SelectedFrequencies")
	if deviceProfiles == nil || pmChannels == nil || measurementTypes == nil ||
		cableLengths == nil || cableNames == nil || frequencies == nil {
		return "Required Parameters are not set", false
	}
	c.global.CLM = tne.CableLossMeasurement{}
	c.global.CLM.Initialize(pmChannels[0], deviceProfiles[0], frequencies)
	if strings.EqualFold(measurementTypes[0], "Reference") {
		go c.global.CLM.MeasurePMReference()
	} else {
		go c.global.CLM.MeasureCableLoss(cableNames[0], cableLengths[0])
	}
	return "", true
}

func handleMeasureTVACCableLoss(c *client, request actionRequest) (string, bool) {
	deviceProfiles := request.getParam("SelectedDeviceProfile")
	pmChannels := request.getParam("SelectedPMChannel")
	measurementTypes := request.getParam("MeasurementType")
	cableNames := request.getParam("CableName")
	testPhase := request.getParam("TestPhase")
	if deviceProfiles == nil || pmChannels == nil || measurementTypes == nil ||
		testPhase == nil || cableNames == nil {
		return "Required Parameters are not set", false
	}
	c.global.TCLM = utilities.TVACCableLossMeasurement{}
	c.global.TCLM.Initialize(pmChannels[0], deviceProfiles[0], testPhase[0])
	if strings.EqualFold(measurementTypes[0], "Reference") {
		go c.global.TCLM.MeasurePMReference()
	} else if strings.EqualFold(measurementTypes[0], "NewCable") {
		go c.global.TCLM.MeasureTVACReference(cableNames[0], testPhase[0])
	} else {
		go c.global.TCLM.MeasureTVACCableLoss(cableNames[0], testPhase[0])
	}
	return "", true
}

func handleAbortCableLossMeasurement(c *client, request actionRequest) (string, bool) {
	c.global.CLM.Stop()
	return "", true
}

func handleMeasureSGPower(c *client, request actionRequest) (string, bool) {
	deviceProfiles := request.getParam("SelectedDeviceProfile")
	receivers := request.getParam("SelectedReceiver")
	spectrumProfiles := request.getParam("SpectrumProfile")
	maxValues := request.getParam("MaxValue")
	minValues := request.getParam("MinValue")
	stepSizes := request.getParam("StepSize")
	fmt.Println(maxValues, minValues)
	if deviceProfiles == nil || receivers == nil || spectrumProfiles == nil ||
		maxValues == nil || minValues == nil || stepSizes == nil {
		return "Required Parameters are not set", false
	}
	maxPower, err := strconv.ParseFloat(maxValues[0], 64)
	if err != nil {
		return "Max Value not a float", false
	}
	minPower, err := strconv.ParseFloat(minValues[0], 64)
	if err != nil {
		return "Min Value not a float", false
	}
	stepSize, err := strconv.ParseFloat(stepSizes[0], 64)
	if err != nil {
		return "Min Value not a float", false
	}
	c.global.SGPower = tne.SGPowerMeasurement{}
	c.global.SGPower.Initialize(deviceProfiles[0], receivers[0], spectrumProfiles[0], maxPower, minPower, stepSize)
	go c.global.SGPower.StartMeasurement()
	return "", true
}

func handleAbortSGPowerMeasurement(c *client, request actionRequest) (string, bool) {
	c.global.SGPower.Stop()
	return "", true
}

func handleMeasureGTxAttn(c *client, request actionRequest) (string, bool) {
	deviceProfiles := request.getParam("SelectedDeviceProfile")
	receivers := request.getParam("SelectedReceiver")
	spectrumProfiles := request.getParam("SpectrumProfile")
	components := request.getParam("Component")
	maxValues := request.getParam("MaxValue")
	minValues := request.getParam("MinValue")
	stepSizes := request.getParam("StepSize")
	fmt.Println(maxValues, minValues)
	if deviceProfiles == nil || receivers == nil || spectrumProfiles == nil || components == nil ||
		maxValues == nil || minValues == nil || stepSizes == nil {
		return "Required Parameters are not set", false
	}
	maxPower, err := strconv.ParseFloat(maxValues[0], 64)
	if err != nil {
		return "Max Value not a float", false
	}
	minPower, err := strconv.ParseFloat(minValues[0], 64)
	if err != nil {
		return "Min Value not a float", false
	}
	stepSize, err := strconv.ParseFloat(stepSizes[0], 64)
	if err != nil {
		return "Min Value not a float", false
	}
	c.global.GTXAttn = tne.GTxAttnMeasurement{}
	c.global.GTXAttn.Initialize(deviceProfiles[0], receivers[0], spectrumProfiles[0], components[0],
		maxPower, minPower, stepSize)
	go c.global.GTXAttn.StartMeasurement()
	return "", true
}

func handleAbortGTxAttnMeasurement(c *client, request actionRequest) (string, bool) {
	c.global.GTXAttn.Stop()
	return "", true
}

func handleMeasureTSMAttn(c *client, request actionRequest) (string, bool) {
	deviceProfiles := request.getParam("SelectedDeviceProfile")
	receivers := request.getParam("SelectedReceiver")
	spectrumProfiles := request.getParam("SpectrumProfile")
	tsmConfigs := request.getParam("TSMConfig")
	maxValues := request.getParam("MaxValue")
	minValues := request.getParam("MinValue")
	stepSizes := request.getParam("StepSize")
	if deviceProfiles == nil || receivers == nil || spectrumProfiles == nil || tsmConfigs == nil ||
		maxValues == nil || minValues == nil || stepSizes == nil {
		return "Required Parameters are not set", false
	}
	maxPower, err := strconv.ParseFloat(maxValues[0], 64)
	if err != nil {
		return "Max Value not a float", false
	}
	minPower, err := strconv.ParseFloat(minValues[0], 64)
	if err != nil {
		return "Min Value not a float", false
	}
	stepSize, err := strconv.ParseFloat(stepSizes[0], 64)
	if err != nil {
		return "Min Value not a float", false
	}
	c.global.TSMAttn = tne.TSMAttnMeasurement{}
	c.global.TSMAttn.Initialize(deviceProfiles[0], receivers[0], spectrumProfiles[0], tsmConfigs[0],
		maxPower, minPower, stepSize)
	go c.global.TSMAttn.StartMeasurement()
	return "", true
}

func handleAbortTSMAttnMeasurement(c *client, request actionRequest) (string, bool) {
	c.global.TSMAttn.Stop()
	return "", true
}

func handleMeasureTSMInternalPathLoss(c *client, request actionRequest) (string, bool) {
	deviceProfiles := request.getParam("SelectedDeviceProfile")
	pmChannel := request.getParam("PMChannel")
	mode := request.getParam("Mode")
	ports := request.getParam("Ports")

	if deviceProfiles == nil || pmChannel == nil || mode == nil || ports == nil {
		return "Required Parameters are not set", false
	}
	c.global.TSMInternal = tne.TSMInternalLoss{}
	c.global.TSMInternal.Initialize(deviceProfiles[0], pmChannel[0])
	go c.global.TSMInternal.MeasureForConfig(mode[0], ports[0], ports[1])
	return "", true
}

func handleAbortTSMInternalMeasurement(c *client, request actionRequest) (string, bool) {
	c.global.TSMInternal.Stop()
	return "", true
}

func handleSpectrumDump(c *client, request actionRequest) (string, bool) {
	selectedSA := request.getParam("SelectedSA")
	if selectedSA == nil {
		return "Unable to get Selected SA [Reconnect]", false
	}
	return getSpectrumDump(c, selectedSA[0])
}

func handleTraceDump(c *client, request actionRequest) (string, bool) {
	selectedSA := request.getParam("SelectedSA")
	tp := request.getParam("TracePoints")
	if selectedSA == nil || tp == nil {
		return "Unable to get Selected SA [Reconnect]", false
	}
	tracePoints, err := strconv.Atoi(tp[0])
	if err != nil {
		return "Trace Points shoule be an integer", false
	}
	return getTraceDump(c, selectedSA[0], tracePoints)
}

func handleScreenshot(c *client, request actionRequest) (string, bool) {
	selectedVSA := request.getParam("SelectedVSA")
	if selectedVSA == nil {
		return "Unable to get Selected VSA [Reconnect]", false
	}
	mode := request.getParam("Mode")
	if mode == nil {
		return "Unable to get Selected Mode", false
	}
	return getScreenshot(c, selectedVSA[0], mode[0])
}

func handlePMPower(c *client, request actionRequest) (string, bool) {
	selectedPM := request.getParam("SelectedPM")
	if selectedPM == nil {
		return "Unable to get Selected PM [Reconnect]", false
	}
	return getPowerFromPM(c, selectedPM[0])
}

func handleStartStability(c *client, request actionRequest) (string, bool) {
	stability := request.getParam("Parameters")
	if stability == nil {
		return "Unable to get Start Stability", false
	}
	//utilities.StartStability(stability, &c.global.Stability)
	return "", true
}

func handleStopStability(c *client, request actionRequest) (string, bool) {
	c.global.Stability.StopStability()
	return "", true
}

func handleStartTests(c *client, request actionRequest) (string, bool) {
	configs := request.getParam("Configs")
	if configs == nil {
		return "Unable to get Selected Configurations", false
	}
	tests := request.getParam("Tests")
	if tests == nil {
		return "Unable to get Selected Tests", false
	}
	remarks := request.getParam("Remarks")
	if remarks == nil {
		return "Unable to get Remarks", false
	}
	extraParameters := request.getParam("ExtraParameters")
	var paramMap = make(map[string]interface{})
	if extraParameters != nil {
		for _, p := range extraParameters {
			temp := strings.Split(p, ";")

			switch temp[0] {
			case "NominalPower":
				power, _ := strconv.ParseFloat(temp[1], 64)
				paramMap["NominalPower"] = power
			}
		}
	}
	var testTypes = make([]string, 0)
	var testCategories = make([]string, 0)
	for _, t := range tests {
		temp := strings.Split(t, ";")
		testTypes = append(testTypes, temp[0])
		if len(temp) == 2 {
			testCategories = append(testCategories, temp[1])
		} else {
			testCategories = append(testCategories, "")
		}
	}
	commChannel := make(chan executeTest.TestProgressResponse, 100)
	inputChannel := make(chan string, 10)

	c.orchestrator = executeTest.NewOrchestrator(configs, testTypes, testCategories, remarks, paramMap, commChannel, inputChannel)
	return "", true
}

func handleAbortTest(c *client, request actionRequest) (string, bool) {
	if c.orchestrator != nil {
		c.orchestrator.Abort()
	}
	return "", true
}

func handleGTxMeasurement(c *client, request actionRequest) (string, bool) {
	c.gtxMeasurement = tne.NewGTxGroundTransmitterMeasurement()
	information := request.getParam("Information")
	modulation := request.getParam("Modulation")
	powerSpectrum := request.getParam("PowerSpectrum")
	freqSpectrum := request.getParam("FrequencySpectrum")
	inBandSpectrum := request.getParam("InBandSpectrum")
	outBandSpectrum := request.getParam("OutBandSpectrum")
	if information == nil || modulation == nil || powerSpectrum == nil ||
		freqSpectrum == nil || inBandSpectrum == nil || outBandSpectrum == nil {
		return "Required Parameters are not set", false
	}
	if len(information) != 4 {
		return "Required Information not set", false
	}
	intFreq, err := strconv.ParseFloat(information[2], 64)
	if err != nil {
		return "Intermediate Frequency is not float", false
	}
	outLoss, err := strconv.ParseFloat(information[3], 64)
	if err != nil {
		return "Output Cable Loss is not float", false
	}
	if len(modulation) != 4 {
		return "Required Modulation Parameters not set", false
	}
	subCar, err := strconv.ParseFloat(modulation[1], 64)
	if err != nil {
		return "Sub Carrier Frequency is not float", false
	}
	freqDev, err := strconv.ParseFloat(modulation[2], 64)
	if err != nil {
		return "Frequency Deviation is not float", false
	}
	mi, err := strconv.ParseFloat(modulation[3], 64)
	if err != nil {
		return "Modulation Index is not float", false
	}
	if len(powerSpectrum) != 3 || len(freqSpectrum) != 3 ||
		len(inBandSpectrum) != 3 || len(outBandSpectrum) != 3 {
		return "Required Spectrum Parameters not set", false
	}
	powerSpan, _ := strconv.ParseFloat(powerSpectrum[0], 64)
	powerRBW, _ := strconv.ParseFloat(powerSpectrum[1], 64)
	powerVBW, _ := strconv.ParseFloat(powerSpectrum[2], 64)
	freqSpan, _ := strconv.ParseFloat(freqSpectrum[0], 64)
	freqRBW, _ := strconv.ParseFloat(freqSpectrum[1], 64)
	freqVBW, _ := strconv.ParseFloat(freqSpectrum[2], 64)
	inSpan, _ := strconv.ParseFloat(inBandSpectrum[0], 64)
	inRBW, _ := strconv.ParseFloat(inBandSpectrum[1], 64)
	inVBW, _ := strconv.ParseFloat(inBandSpectrum[2], 64)
	outSpan, _ := strconv.ParseFloat(outBandSpectrum[0], 64)
	outRBW, _ := strconv.ParseFloat(outBandSpectrum[1], 64)
	outVBW, _ := strconv.ParseFloat(outBandSpectrum[2], 64)
	ok := c.gtxMeasurement.SetDevices(information[0], information[1], intFreq, outLoss)
	if !ok {
		return "Unable to load Device Profile", false
	}

	c.gtxMeasurement.SetModulationParameters(modulation[0], subCar, freqDev, mi)
	c.gtxMeasurement.SetPowerSpectrum(powerSpan, powerRBW, powerVBW)
	c.gtxMeasurement.SetFrequencySpectrum(freqSpan, freqRBW, freqVBW)
	c.gtxMeasurement.SetInBandSpectrum(inSpan, inRBW, inVBW)
	c.gtxMeasurement.SetOutBandSpectrum(outSpan, outRBW, outVBW)
	return "", true
}
