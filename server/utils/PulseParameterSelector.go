package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"prismServer/logger"
)

func GetAllPpmParameters() []string {
	var tbr []string
	val := reflect.ValueOf(PPMParameters{})
	for i := 0; i < val.Type().NumField(); i++ {
		tbr = append(tbr, val.Type().Field(i).Name)
	}
	return tbr
}

func GetAllVsaParameters() []string {
	var tbr []string
	val := reflect.ValueOf(VSAParameters{})
	for i := 0; i < val.Type().NumField(); i++ {
		tbr = append(tbr, val.Type().Field(i).Name)
	}
	return tbr
}

type PPMParameters struct {
	AverageTxPower    bool
	PulseAveragePower bool
	PeakPower         bool
	DutyCycle         bool
	RiseTime          bool
	FallTime          bool
	PulsePeriod       bool
	PulseWidth        bool
	PulseSeparation   bool
}

type VSAParameters struct {
	AverageTxPower        bool
	PulseAveragePower     bool
	PeakPower             bool
	DutyCycle             bool
	RiseTime              bool
	FallTime              bool
	PulsePeriod           bool
	PulseWidth            bool
	FrequencyDeviation    bool
	PRF                   bool
	PhaseDeviation        bool
	DutyFactor            bool
	Ripple                bool
	Droop                 bool
	PulseToPulsePhaseDiff bool
	PulseToPulseFreqDiff  bool
	RMSFrequencyError     bool
	MaxFrequencyError     bool
	RMSPhaseError         bool
	MaxPhaseError         bool
	ChirpBandwidth        bool
	Overshoot             bool
	ChirpRate             bool
	ChirpRateDeviation    bool
}

type ReportSelectedParameters struct {
	PPM     PPMParameters
	VSA     VSAParameters
	OK      bool
	Message string
}

var SelectedParameters ReportSelectedParameters

func getDefaultPPMParameters() PPMParameters {
	var ppm PPMParameters
	ppm.PulseAveragePower = true
	ppm.AverageTxPower = true
	ppm.PeakPower = true
	ppm.DutyCycle = true
	ppm.RiseTime = true
	ppm.FallTime = true
	ppm.PulsePeriod = true
	ppm.PulseWidth = true
	ppm.PulseSeparation = true
	return ppm
}

func getDefaultVSAParameters() VSAParameters {
	var vsa VSAParameters
	vsa.PulseAveragePower = true
	vsa.AverageTxPower = true
	vsa.PeakPower = true
	vsa.DutyCycle = true
	vsa.RiseTime = true
	vsa.FallTime = true
	vsa.PulsePeriod = true
	vsa.PulseWidth = true
	vsa.PRF = true
	return vsa
}

func ReadSelectionParams() {
	filename := filepath.Join(Config.BaseFolder, ".resources/pulseParametersSelected.json")
	configFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("File ", filename, " doesn't exist")
		fmt.Println("Creating default configuration")
		var sel ReportSelectedParameters
		sel.PPM = getDefaultPPMParameters()
		sel.VSA = getDefaultVSAParameters()
		data, err := json.MarshalIndent(sel, "", " ")
		if err != nil {
			fmt.Println("Unable to create default parameters")
			return
		}
		err = os.WriteFile(filename, data, 0666)
		if err != nil {
			fmt.Println("Unable to create default parameters")
			return
		}
		configFile, _ = os.Open(filename)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&SelectedParameters)
}

func GetSelectedVSAParams() []string {
	var tbr = make([]string, 0)
	if SelectedParameters.VSA.AverageTxPower {
		tbr = append(tbr, "AverageTxPowerVSA")
	}
	if SelectedParameters.VSA.PulseAveragePower {
		tbr = append(tbr, "PulseAveragePowerVSA")
	}
	if SelectedParameters.VSA.PeakPower {
		tbr = append(tbr, "PeakPowerVSA")
	}
	if SelectedParameters.VSA.DutyCycle {
		tbr = append(tbr, "DutyCycle")
	}
	if SelectedParameters.VSA.RiseTime {
		tbr = append(tbr, "RiseTime")
	}
	if SelectedParameters.VSA.FallTime {
		tbr = append(tbr, "FallTime")
	}
	if SelectedParameters.VSA.PulsePeriod {
		tbr = append(tbr, "PulsePeriod")
	}
	if SelectedParameters.VSA.PulseWidth {
		tbr = append(tbr, "PulseWidth")
	}
	if SelectedParameters.VSA.FrequencyDeviation {
		tbr = append(tbr, "FrequencyDeviation")
	}
	if SelectedParameters.VSA.PRF {
		tbr = append(tbr, "RepetitionRate")
	}
	if SelectedParameters.VSA.PhaseDeviation {
		tbr = append(tbr, "PhaseDeviation")
	}
	if SelectedParameters.VSA.DutyFactor {
		tbr = append(tbr, "DutyFactor")
	}
	if SelectedParameters.VSA.Ripple {
		tbr = append(tbr, "Ripple")
	}
	if SelectedParameters.VSA.Droop {
		tbr = append(tbr, "Droop")
	}
	if SelectedParameters.VSA.PulseToPulsePhaseDiff {
		tbr = append(tbr, "PulseToPulsePhaseDiff")
	}
	if SelectedParameters.VSA.PulseToPulseFreqDiff {
		tbr = append(tbr, "PulseToPulseFreqDiff")
	}
	if SelectedParameters.VSA.RMSFrequencyError {
		tbr = append(tbr, "RMSFrequencyError")
	}
	if SelectedParameters.VSA.MaxFrequencyError {
		tbr = append(tbr, "MaxFrequencyError")
	}
	if SelectedParameters.VSA.RMSPhaseError {
		tbr = append(tbr, "RMSPhaseError")
	}
	if SelectedParameters.VSA.MaxPhaseError {
		tbr = append(tbr, "MaxPhaseError")
	}
	if SelectedParameters.VSA.ChirpBandwidth {
		tbr = append(tbr, "ChirpBandwidth")
	}
	if SelectedParameters.VSA.Overshoot {
		tbr = append(tbr, "Overshoot")
	}
	if SelectedParameters.VSA.ChirpRate {
		tbr = append(tbr, "ChirpRate")
	}
	if SelectedParameters.VSA.ChirpRateDeviation {
		tbr = append(tbr, "ChirpRateDeviation")
	}
	return tbr
}

func GetSelectedPPMParams() []string {
	var tbr = make([]string, 0)
	/*if SelectedParameters.PPM.AverageTxPower {
		tbr = append(tbr, "AverageTxPowerPPM")
	}*/
	if SelectedParameters.PPM.PulseAveragePower {
		tbr = append(tbr, "PulseAveragePowerPPM")
	}
	if SelectedParameters.PPM.PeakPower {
		tbr = append(tbr, "PeakPowerPPM")
	}
	if SelectedParameters.PPM.DutyCycle {
		tbr = append(tbr, "DutyCycle")
	}
	if SelectedParameters.PPM.RiseTime {
		tbr = append(tbr, "RiseTime")
	}
	if SelectedParameters.PPM.FallTime {
		tbr = append(tbr, "FallTime")
	}
	if SelectedParameters.PPM.PulsePeriod {
		tbr = append(tbr, "PulsePeriod")
	}
	if SelectedParameters.PPM.PulseWidth {
		tbr = append(tbr, "PulseWidth")
	}
	if SelectedParameters.PPM.PulseSeparation {
		tbr = append(tbr, "PulseSeparation")
	}
	return tbr
}

func SetSelectedPPMParameter(parameters []string) {
	var ppm PPMParameters
	if slices.Contains(parameters, "AverageTxPowerPPM") {
		ppm.AverageTxPower = true
	}
	if slices.Contains(parameters, "PulseAveragePowerPPM") {
		ppm.PulseAveragePower = true
	}
	if slices.Contains(parameters, "PeakPowerPPM") {
		ppm.PeakPower = true
	}
	if slices.Contains(parameters, "DutyCycle") {
		ppm.DutyCycle = true
	}
	if slices.Contains(parameters, "RiseTime") {
		ppm.RiseTime = true
	}
	if slices.Contains(parameters, "FallTime") {
		ppm.FallTime = true
	}
	if slices.Contains(parameters, "PulsePeriod") {
		ppm.PulsePeriod = true
	}
	if slices.Contains(parameters, "PulseWidth") {
		ppm.PulseWidth = true
	}
	if slices.Contains(parameters, "PulseSeparation") {
		ppm.PulseSeparation = true
	}
	SelectedParameters.PPM = ppm

	filename := filepath.Join(Config.BaseFolder, ".resources/pulseParametersSelected.json")

	data, err := json.MarshalIndent(SelectedParameters, "", " ")
	if err != nil {
		logger.Log.Error("Unable to create default parameters", "error", err)
		return
	}
	err = os.WriteFile(filename, data, 0666)
	if err != nil {
		logger.Log.Error("Unable to create default parameters", "error", err)
		return
	}

	ReadSelectionParams()
}

func SetSelectedVSAParameter(parameters []string) {
	logger.Log.Debug("SetSelectedVSAParameter called", "parameters", parameters)
	var vsa VSAParameters
	if slices.Contains(parameters, "AverageTxPowerVSA") {
		vsa.AverageTxPower = true
	}
	if slices.Contains(parameters, "PulseAveragePowerVSA") {
		vsa.PulseAveragePower = true
	}
	if slices.Contains(parameters, "PeakPowerVSA") {
		vsa.PeakPower = true
	}
	if slices.Contains(parameters, "DutyCycle") {
		vsa.DutyCycle = true
	}
	if slices.Contains(parameters, "RiseTime") {
		vsa.RiseTime = true
	}
	if slices.Contains(parameters, "FallTime") {
		vsa.FallTime = true
	}
	if slices.Contains(parameters, "PulsePeriod") {
		vsa.PulsePeriod = true
	}
	if slices.Contains(parameters, "PulseWidth") {
		vsa.PulseWidth = true
	}
	if slices.Contains(parameters, "FrequencyDeviation") {
		vsa.FrequencyDeviation = true
	}
	if slices.Contains(parameters, "RepetitionRate") {
		vsa.PRF = true
	}
	if slices.Contains(parameters, "PhaseDeviation") {
		vsa.PhaseDeviation = true
	}
	if slices.Contains(parameters, "DutyFactor") {
		vsa.DutyFactor = true
	}
	if slices.Contains(parameters, "Ripple") {
		vsa.Ripple = true
	}
	if slices.Contains(parameters, "Droop") {
		vsa.Droop = true
	}
	if slices.Contains(parameters, "PulseToPulsePhaseDiff") {
		vsa.PulseToPulsePhaseDiff = true
	}
	if slices.Contains(parameters, "PulseToPulseFreqDiff") {
		vsa.PulseToPulseFreqDiff = true
	}
	if slices.Contains(parameters, "RMSFrequencyError") {
		vsa.RMSFrequencyError = true
	}
	if slices.Contains(parameters, "MaxFrequencyError") {
		vsa.MaxFrequencyError = true
	}
	if slices.Contains(parameters, "RMSPhaseError") {
		vsa.RMSPhaseError = true
	}
	if slices.Contains(parameters, "MaxPhaseError") {
		vsa.MaxPhaseError = true
	}
	if slices.Contains(parameters, "ChirpBandwidth") {
		vsa.ChirpBandwidth = true
	}
	if slices.Contains(parameters, "Overshoot") {
		vsa.Overshoot = true
	}
	if slices.Contains(parameters, "ChirpRate") {
		vsa.ChirpRate = true
	}
	if slices.Contains(parameters, "ChirpRateDeviation") {
		vsa.ChirpRateDeviation = true
	}

	SelectedParameters.VSA = vsa

	filename := filepath.Join(Config.BaseFolder, ".resources/pulseParametersSelected.json")
	data, err := json.MarshalIndent(SelectedParameters, "", " ")
	if err != nil {
		logger.Log.Error("Unable to create default parameters", "error", err)
		return
	}
	err = os.WriteFile(filename, data, 0666)
	if err != nil {
		logger.Log.Error("Unable to create default parameters", "error", err)
		return
	}

	ReadSelectionParams()
}
