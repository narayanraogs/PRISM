package tne

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/logger"
	"prismServer/reports"
	"prismServer/resultsDB"
	"prismServer/utils"
	"slices"
	"strings"
	"time"
)

const (
	UCDCGainInternalCable    = "UCDC_GAIN_INT_CABLE"
	UCDCGainInternalRadiated = "UCDC_GAIN_INT_RAD"
	UCDCFreqMeas             = "UCDC_FREQ_MEAS"
	UCDCHarmonicMeas         = "UCDC_HARM_MEAS"
	UCDCSpuriousInBand       = "UCDC_SPUR_IN_BAND"
	UCDCSpuriousOutBand      = "UCDC_SPUR_OUT_BAND"
	UCDCLOLeakage            = "UCDC_LO_LEAKAGE"
	UCDCInputLeakage         = "UCDC_INPUT_LEAKAGE"
	UCDCGainExternalCable    = "UCDC_GAIN_EXT_CABLE"
	UCDCGainExternalRadiated = "UCDC_GAIN_EXT_RAD"
	UCDCOutputMonPower       = "UCDC_OUT_MON_POWER"
	UCDCInputMonPower        = "UCDC_IN_MON_POWER"
	UCDCLOMonPower           = "UCDC_LO_MON_POWER"
	UCDCLOMonPhaseNoise      = "UCDC_LO_MON_PN"
	UCDCExtLOPowerMatch      = "UCDC_EXT_LO_MATCH"
)

type UpDownConverterMeasurement struct {
	deviceProfile      string
	externalSGName     string
	converterName      string
	converter          database.UpDownConverter
	inputFrequency     float64
	outputFrequency    float64
	stageFactor        int64
	inputPower         float64
	inputCableLoss     float64
	outputCableLoss    []float64
	loCableLoss        float64
	sa                 driver.SA
	sg                 driver.SG
	sgExt              driver.SG
	statusMonitor      chan RTStatus
	measurementMonitor chan ConvertorResults
	powerSpectrum      spectrumSettings
	frequencySpectrum  spectrumSettings
	inBandSpectrum     spectrumSettings
	outBandSpectrum    spectrumSettings
	stop               bool
}

// Internal Helpers

func (udc *UpDownConverterMeasurement) notify(msg string) {
	if udc.statusMonitor != nil {
		udc.statusMonitor <- RTStatus{Message: msg, Success: true}
	}
}

func (udc *UpDownConverterMeasurement) finish(msg string, success bool) {
	if udc.statusMonitor != nil {
		udc.statusMonitor <- RTStatus{
			Message:   msg,
			Success:   success,
			Error:     !success,
			Completed: true,
		}
		close(udc.statusMonitor)
	}
	if udc.measurementMonitor != nil {
		close(udc.measurementMonitor)
	}
}

func (udc *UpDownConverterMeasurement) setError(msg string) {
	logger.Log.Error("Up/Down Converter Measurement Error", "error", msg)
	udc.finish(msg, false)
}

func (udc *UpDownConverterMeasurement) check(resp utils.CommandResponse, errMsg string) bool {
	if !resp.Success {
		udc.setError(fmt.Sprintf("%s: %s", errMsg, resp.ErrorMessage))
		return false
	}
	if udc.stop {
		udc.setError("Measurement aborted by user")
		return false
	}
	return true
}

func parseFrequenciesFromString(text string) (float64, float64, bool) {
	idx := strings.Index(text, "(In:")
	if idx == -1 {
		return 0, 0, false
	}
	sub := text[idx:]
	var inMHz, outMHz float64
	n, _ := fmt.Sscanf(sub, "(In: %f MHz, Out: %f MHz)", &inMHz, &outMHz)
	if n == 2 && inMHz > 0 && outMHz > 0 {
		return inMHz * 1e6, outMHz * 1e6, true
	}
	return 0, 0, false
}

func (udc *UpDownConverterMeasurement) save(result ConvertorResults, testType string) bool {
	result.InputFrequency = udc.inputFrequency
	result.OutputFrequency = udc.outputFrequency
	result.LOFrequency = udc.loFrequency()
	result.StageFactor = udc.stageFactor
	result.FrequencyMap = udc.converter.FrequencyMap

	saveTestType := testType
	if udc.inputFrequency > 0 && udc.outputFrequency > 0 {
		saveTestType = fmt.Sprintf("%s (In: %.2f MHz, Out: %.2f MHz)", testType, udc.inputFrequency/1e6, udc.outputFrequency/1e6)
	}
	result.TestName = saveTestType
	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		udc.setError("Failed to process results")
		return false
	}
	if err := resultsDB.InsertUpDownConverterResult(udc.converterName, saveTestType, string(data)); err != nil {
		udc.setError("Failed to save results to database")
		return false
	}
	return true
}

func (udc *UpDownConverterMeasurement) GeneratePDF(name string, dates, times []string, targetFreqs ...float64) (string, bool) {
	var report reports.Report
	testPhase, _ := database.GetSelectedTestPhase()
	report.SetHeader(name, "Up/Down Converter Results", "", testPhase)
	udc.converterName = name

	var targetIn, targetOut float64
	if len(targetFreqs) >= 2 && targetFreqs[0] > 0 && targetFreqs[1] > 0 {
		targetIn = targetFreqs[0]
		targetOut = targetFreqs[1]
	}

	// 1. Fetch raw sessions from results database
	type sessionItem struct {
		result ConvertorResults
		date   string
		time   string
	}
	var rawSessions []sessionItem
	for i := 0; i < len(dates); i++ {
		res, err := udc.getResult(dates[i], times[i])
		if err != nil {
			return "", false
		}
		rawSessions = append(rawSessions, sessionItem{result: res, date: dates[i], time: times[i]})
	}

	// Fallback converter spec from database
	details, err := database.GetConverterDetails(strings.TrimSpace(name))
	var dbPairs []database.FrequencyPair
	var sf int64 = 1
	if err == nil {
		dbPairs, _ = details.ParseFrequencyMap()
		sf = details.StageFactor
		if sf <= 0 {
			sf = 1
		}
	}

	// Prefer StageFactor from measurement result if tagged
	for _, s := range rawSessions {
		if s.result.StageFactor > 0 {
			sf = s.result.StageFactor
			break
		}
	}

	// 2. Normalize and backfill frequencies from TestName if needed
	var measuredSessions []sessionItem
	for _, s := range rawSessions {
		res := s.result
		if res.InputFrequency <= 0 {
			if inFreq, outFreq, ok := parseFrequenciesFromString(res.TestName); ok {
				res.InputFrequency = inFreq
				res.OutputFrequency = outFreq
			}
		}
		measuredSessions = append(measuredSessions, sessionItem{result: res, date: s.date, time: s.time})
	}

	// Filter sessions by target frequency set if explicitly selected
	var filteredSessions []sessionItem
	if targetIn > 0 && targetOut > 0 {
		for _, s := range measuredSessions {
			res := s.result
			if res.InputFrequency > 0 && (math.Abs(res.InputFrequency-targetIn) > 1e3 || math.Abs(res.OutputFrequency-targetOut) > 1e3) {
				continue
			}
			res.InputFrequency = targetIn
			res.OutputFrequency = targetOut
			res.LOFrequency = math.Abs(targetIn-targetOut) / float64(sf)
			filteredSessions = append(filteredSessions, sessionItem{result: res, date: s.date, time: s.time})
		}
	} else {
		filteredSessions = measuredSessions
	}

	// 3. Determine the actual frequency sets that were measured in these sessions
	type freqPair struct {
		Input  float64
		Output float64
		LO     float64
	}
	var activePairs []freqPair
	seenPair := make(map[string]bool)
	for _, s := range filteredSessions {
		res := s.result
		if res.InputFrequency > 0 && res.OutputFrequency > 0 {
			key := fmt.Sprintf("%.2f_%.2f", res.InputFrequency/1e6, res.OutputFrequency/1e6)
			if !seenPair[key] {
				seenPair[key] = true
				lo := res.LOFrequency
				if lo <= 0 {
					lo = math.Abs(res.InputFrequency-res.OutputFrequency) / float64(sf)
				}
				activePairs = append(activePairs, freqPair{
					Input:  res.InputFrequency,
					Output: res.OutputFrequency,
					LO:     lo,
				})
			}
		}
	}

	// If no frequencies were tagged in the results (legacy data without frequency tags):
	if len(activePairs) == 0 {
		if targetIn > 0 && targetOut > 0 {
			activePairs = append(activePairs, freqPair{
				Input:  targetIn,
				Output: targetOut,
				LO:     math.Abs(targetIn-targetOut) / float64(sf),
			})
		} else if len(dbPairs) > 0 {
			// Even if db table was later edited to have multiple pairs, this measurement was only for 1 pair!
			p0 := dbPairs[0]
			lo := math.Abs(float64(p0.Input-p0.Output)) / float64(sf)
			activePairs = append(activePairs, freqPair{
				Input:  float64(p0.Input),
				Output: float64(p0.Output),
				LO:     lo,
			})
		}
	}

	// Ensure all filtered sessions have frequency attributes populated
	if len(activePairs) == 1 {
		p0 := activePairs[0]
		for i := range filteredSessions {
			if filteredSessions[i].result.InputFrequency <= 0 {
				filteredSessions[i].result.InputFrequency = p0.Input
				filteredSessions[i].result.OutputFrequency = p0.Output
				filteredSessions[i].result.LOFrequency = p0.LO
			}
		}
	}

	// 4. Build Test Information block ONLY for activePairs (the frequencies of these measurements!)
	report.TestInformation = append(report.TestInformation, reports.ReportInfo{
		Parameter: "Converter Name",
		Value:     name,
	})
	report.TestInformation = append(report.TestInformation, reports.ReportInfo{
		Parameter: "Stage Factor",
		Value:     fmt.Sprintf("%d", sf),
	})

	if len(activePairs) == 1 {
		p := activePairs[0]
		report.TestInformation = append(report.TestInformation, reports.ReportInfo{
			Parameter: "Input Frequency",
			Value:     fmt.Sprintf("%.2f MHz", p.Input/1e6),
		})
		report.TestInformation = append(report.TestInformation, reports.ReportInfo{
			Parameter: "Output Frequency",
			Value:     fmt.Sprintf("%.2f MHz", p.Output/1e6),
		})
		report.TestInformation = append(report.TestInformation, reports.ReportInfo{
			Parameter: "LO Frequency",
			Value:     fmt.Sprintf("%.2f MHz", p.LO/1e6),
		})
		report.TestInformation = append(report.TestInformation, reports.ReportInfo{
			Parameter: "",
			Value:     "",
		})
	} else if len(activePairs) > 1 {
		for idx, p := range activePairs {
			report.TestInformation = append(report.TestInformation, reports.ReportInfo{
				Parameter: fmt.Sprintf("Stage %d Frequencies", idx+1),
				Value:     fmt.Sprintf("In: %.1f MHz, Out: %.1f MHz", p.Input/1e6, p.Output/1e6),
			})
			report.TestInformation = append(report.TestInformation, reports.ReportInfo{
				Parameter: fmt.Sprintf("Stage %d LO", idx+1),
				Value:     fmt.Sprintf("%.1f MHz", p.LO/1e6),
			})
		}
	}

	order := make([]string, 0)
	seen := make(map[string]bool)

	for i := 0; i < len(filteredSessions); i++ {
		result := filteredSessions[i].result

		var baseReportName string
		switch result.TestCode {
		case UCDCGainInternalCable:
			baseReportName = "Output Port - Gain Measurement - Internal LO - Cable Mode"
		case UCDCGainInternalRadiated:
			baseReportName = "Output Port - Gain Measurement - Internal LO - Radiated Mode"
		case UCDCFreqMeas:
			baseReportName = "Output Port - Frequency Measurement"
		case UCDCHarmonicMeas:
			baseReportName = "Output Port - Harmonics Measurement"
		case UCDCSpuriousInBand:
			baseReportName = "Output Port - Spurious - In Band"
		case UCDCSpuriousOutBand:
			baseReportName = "Output Port - Spurious - Out of Band"
		case UCDCLOLeakage:
			baseReportName = "Output Port - LO Leakage"
		case UCDCInputLeakage:
			baseReportName = "Output Port - Input Leakage"
		case UCDCGainExternalCable:
			baseReportName = "Output Port - Gain Measurement - External LO - Cable Mode"
		case UCDCGainExternalRadiated:
			baseReportName = "Output Port - Gain Measurement - External LO - Radiated Mode"
		case UCDCOutputMonPower:
			baseReportName = "Output Monitor Port - Power Measurement"
		case UCDCInputMonPower:
			baseReportName = "Input Monitor Port - Power Measurement"
		case UCDCLOMonPower:
			baseReportName = "LO Monitor Port - Power Measurement"
		case UCDCLOMonPhaseNoise:
			baseReportName = "LO Monitor Port - Phase Noise Measurement"
		case UCDCExtLOPowerMatch:
			baseReportName = "LO Monitor Port - External LO Power Match"
		}

		reportName := baseReportName
		if result.InputFrequency > 0 && result.OutputFrequency > 0 {
			reportName = fmt.Sprintf("%s (In: %.2f MHz, Out: %.2f MHz)", baseReportName, result.InputFrequency/1e6, result.OutputFrequency/1e6)
		}

		if !seen[reportName] {
			seen[reportName] = true
			order = append(order, reportName)
		}

		switch result.TestCode {
		case UCDCGainInternalCable, UCDCGainInternalRadiated, UCDCGainExternalCable, UCDCGainExternalRadiated:
			table := result.GainResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCFreqMeas:
			table := result.FrequencyResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCHarmonicMeas:
			table := result.HarmonicResultValue.getResultTable()
			for hIdx := 0; hIdx < len(result.HarmonicResultValue.HarmonicNo); hIdx++ {
				if len(result.SpectrumDump) > hIdx {
					report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[hIdx], Caption: fmt.Sprintf("%s - %d Harmonic", reportName, result.HarmonicResultValue.HarmonicNo[hIdx])})
				}
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCSpuriousInBand, UCDCSpuriousOutBand:
			table := result.SpuriousResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCLOLeakage, UCDCInputLeakage, UCDCOutputMonPower, UCDCInputMonPower:
			table := result.PowerOrLeakageResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCLOMonPower:
			table := result.PowerOrLeakageResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			}
			report.SetResults(reportName, table.Header, table.Data)

			reportName2 := "LO Monitor Port - Frequency Measurement"
			if result.InputFrequency > 0 && result.OutputFrequency > 0 {
				reportName2 = fmt.Sprintf("%s (In: %.2f MHz, Out: %.2f MHz)", reportName2, result.InputFrequency/1e6, result.OutputFrequency/1e6)
			}
			if !seen[reportName2] {
				seen[reportName2] = true
				order = append(order, reportName2)
			}
			table2 := result.FrequencyResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName2})
			}
			report.SetResults(reportName2, table2.Header, table2.Data)
		case UCDCLOMonPhaseNoise:
			table := result.PhaseNoiseResultValue.getResultTable()
			if len(result.SpectrumDump) > 0 {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCExtLOPowerMatch:
			table := result.PowerMatchingResultValue.getResultTable()
			report.SetResults(reportName, table.Header, table.Data)
		}
	}
	report.SetOrder(order)
	pdf, err := reports.GenerateResult(report, true, false, false, true, true)
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(pdf)
	if err != nil {
		return "", false
	}
	return base64.StdEncoding.EncodeToString(data), true
}

func (udc *UpDownConverterMeasurement) setupBasic(inputFreq, outputFreq float64) bool {
	if !udc.check(udc.sa.SetAlignmentOff(), "SA: alignment off") {
		return false
	}
	if !udc.check(udc.sg.SetFrequency(inputFreq), "SG: frequency set") {
		return false
	}
	if !udc.check(udc.sg.SetModOff(), "SG: mod off") {
		return false
	}
	return true
}

func (udc *UpDownConverterMeasurement) GetStatusMonitor() (chan RTStatus, chan ConvertorResults) {
	udc.statusMonitor = make(chan RTStatus, 20)
	udc.measurementMonitor = make(chan ConvertorResults, 20)
	return udc.statusMonitor, udc.measurementMonitor
}

func (udc *UpDownConverterMeasurement) Init(deviceProfile, externalSGName, converterName string) bool {
	udc.deviceProfile = deviceProfile
	udc.externalSGName = externalSGName
	udc.converterName = converterName

	ucdc, err := database.GetConverterDetails(converterName)
	if err != nil {
		udc.setError("Unable to read converter from database")
		return false
	}
	udc.converter = ucdc
	udc.inputFrequency = ucdc.PrimaryInputFreq()
	udc.outputFrequency = ucdc.PrimaryOutputFreq()
	udc.stageFactor = ucdc.StageFactor
	if udc.stageFactor <= 0 {
		udc.stageFactor = 1
	}

	saName, okSa := database.GetSAFromDeviceProfile(deviceProfile)
	sgName, okSg := database.GetSGFromDeviceProfile(deviceProfile)
	if !okSa || !okSg {
		udc.setError("Unable to resolve devices from profile")
		return false
	}

	if strings.EqualFold(sgName, externalSGName) {
		udc.setError("External LO SG and Internal LO SG cannot be the same")
		return false
	}

	if !udc.sa.LoadDevice(saName) || !udc.sg.LoadDevice(sgName) || !udc.sgExt.LoadDevice(externalSGName) {
		udc.setError("Failed to load device drivers")
		return false
	}
	udc.stop = false
	return true
}

func (udc *UpDownConverterMeasurement) SetFrequencyPair(inputFreq, outputFreq float64) {
	udc.inputFrequency = inputFreq
	udc.outputFrequency = outputFreq
}

func (udc *UpDownConverterMeasurement) loFrequency() float64 {
	sf := float64(udc.stageFactor)
	if sf <= 0 {
		sf = 1
	}
	return math.Abs(udc.inputFrequency - udc.outputFrequency) / sf
}

func (udc *UpDownConverterMeasurement) GetFrequencyPairs() []database.FrequencyPair {
	pairs, err := udc.converter.ParseFrequencyMap()
	if err != nil || len(pairs) == 0 {
		return []database.FrequencyPair{
			{Input: int64(udc.converter.PrimaryInputFreq()), Output: int64(udc.converter.PrimaryOutputFreq())},
		}
	}
	return pairs
}

func (udc *UpDownConverterMeasurement) IsStopped() bool {
	return udc.stop
}

func (udc *UpDownConverterMeasurement) GetConverter() database.UpDownConverter {
	return udc.converter
}

func (udc *UpDownConverterMeasurement) populateFreqs(r *ConvertorResults) {
	r.InputFrequency = udc.inputFrequency
	r.OutputFrequency = udc.outputFrequency
	r.LOFrequency = udc.loFrequency()
}

func (udc *UpDownConverterMeasurement) SetInputCableLoss(loss, power float64) {
	udc.inputPower = power
	udc.inputCableLoss = math.Abs(loss)
}

func (udc *UpDownConverterMeasurement) SetOutputCableLoss(loss []float64) {
	udc.outputCableLoss = make([]float64, len(loss))
	for i, v := range loss {
		udc.outputCableLoss[i] = math.Abs(v)
	}
}

func (udc *UpDownConverterMeasurement) SetLOCableLoss(loss float64) {
	udc.loCableLoss = math.Abs(loss)
}

func (udc *UpDownConverterMeasurement) SetPowerSpectrum(span, rbw, vbw float64) {
	udc.powerSpectrum = spectrumSettings{span, rbw, vbw}
}

func (udc *UpDownConverterMeasurement) SetFrequencySpectrum(span, rbw, vbw float64) {
	udc.frequencySpectrum = spectrumSettings{span, rbw, vbw}
}

func (udc *UpDownConverterMeasurement) SetInBandSpectrum(span, rbw, vbw float64) {
	udc.inBandSpectrum = spectrumSettings{span, rbw, vbw}
}

func (udc *UpDownConverterMeasurement) SetOutBandSpectrum(span, rbw, vbw float64) {
	udc.outBandSpectrum = spectrumSettings{span, rbw, vbw}
}

func (udc *UpDownConverterMeasurement) Stop() {
	udc.stop = true
}

func (udc *UpDownConverterMeasurement) OutputGainMeasurement(stepSize float64, cable bool) {
	logger.Log.Info("Starting UDC Gain Measurement", "cable", cable)
	udc.notify("Gain Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}

	maxPower, minPower := udc.converter.MaxPowerCable, udc.converter.MinPowerCable
	testName, testCode := "Output Port - Gain Measurement - Internal LO - Cable", UCDCGainInternalCable
	if !cable {
		maxPower, minPower = udc.converter.MaxPowerRadiated.Float64, udc.converter.MinPowerRadiated.Float64
		testName, testCode = "Output Port - Gain Measurement - Internal LO - Radiated", UCDCGainInternalRadiated
	}

	result := ConvertorResults{TestName: testName, TestCode: testCode, GainResults: true}
	udc.populateFreqs(&result)

	if !udc.check(udc.sa.SetSpectrum(udc.outputFrequency, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sg.SetPower(minPower-udc.inputCableLoss), "SG: set start power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: reference carrier") {
		return
	}
	time.Sleep(time.Second)

	for p := minPower; p <= maxPower; p += stepSize {
		if !udc.check(udc.sg.SetPower(p+udc.inputCableLoss), "SG: iterate power") {
			return
		}
		time.Sleep(time.Second)
		if !udc.check(udc.sa.SetReferenceNominal(), "SA: reference carrier loop") {
			return
		}
		time.Sleep(time.Second)

		if p == minPower {
			resp := udc.sa.GetSpectrumDump()
			if !udc.check(resp, "SA: Spectrum dump") {
				return
			}
			result.SpectrumDump = []string{resp.Result["SpectrumDump"].String}
		}

		resp := udc.sa.GetMaxMarkerValue()
		if !udc.check(resp, "SA: read power") {
			return
		}

		outPower := resp.Result["MarkerY"].Value + udc.outputCableLoss[1]
		gain := outPower - p

		result.GainResultValue.SetPower = append(result.GainResultValue.SetPower, p)
		result.GainResultValue.OutputPower = append(result.GainResultValue.OutputPower, outPower)
		result.GainResultValue.Gain = append(result.GainResultValue.Gain, gain)
		result.GainResultValue.AverageGain = mean(result.GainResultValue.Gain)

		logger.Log.Info("UDC Gain Measurement Point", "setPower", p, "outPower", outPower, "gain", gain)
		udc.measurementMonitor <- result
		udc.notify(fmt.Sprintf("Completed measurement for %.3f dBm", p))
	}

	if udc.save(result, testName) {
		logger.Log.Info("Completed UDC Gain Measurement", "success", true)
		udc.finish("Gain Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputFrequencyMeasurement() {
	logger.Log.Info("Starting UDC Frequency Measurement")
	udc.notify("Frequency Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}

	if !udc.check(udc.sa.SetSpectrum(udc.outputFrequency, udc.frequencySpectrum.span, udc.frequencySpectrum.rbw, udc.frequencySpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sg.SetPower(udc.inputPower+udc.inputCableLoss), "SG: set power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	udc.sa.WaitForSweeps(2)

	if !udc.check(udc.sa.PeakSearch(true, 1), "SA: peak search") {
		return
	}
	udc.sa.WaitForSweeps(5)

	resp := udc.sa.GetFrequencyInCounterMode(1)
	if !udc.check(resp, "SA: frequency counter") {
		return
	}

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(resp, "SA: Spectrum dump") {
		return
	}

	freq := resp.Result["Frequency"].Value
	deviation := math.Abs(freq - udc.outputFrequency)

	result := ConvertorResults{
		TestName: "Output Port - Frequency Measurement", TestCode: UCDCFreqMeas, FrequencyResults: true,
		FrequencyResultValue: FrequencyResults{ExpectedFrequency: udc.outputFrequency, MeasuredFrequency: freq, Deviation: deviation},
	}
	udc.populateFreqs(&result)
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	logger.Log.Info("Completed UDC Frequency Measurement", "measuredFreq", freq, "deviation", deviation)
	udc.measurementMonitor <- result
	if udc.save(result, result.TestName) {
		udc.finish("Frequency Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputHarmonicsMeasurement() {
	logger.Log.Info("Starting UDC Harmonics Measurement")
	udc.notify("Harmonics Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}

	if !udc.check(udc.sa.SetSpectrum(udc.outputFrequency, udc.frequencySpectrum.span, udc.frequencySpectrum.rbw, udc.frequencySpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sg.SetPower(udc.inputPower+udc.inputCableLoss), "SG: set power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	resp := udc.sa.PeakSearch(true, 1)
	if !udc.check(resp, "SA: peak search") {
		return
	}
	mainPower := resp.Result["MarkerY"].Value

	result := ConvertorResults{TestName: "Output Port - Harmonics Measurement", TestCode: UCDCHarmonicMeas, HarmonicsResults: true}
	udc.populateFreqs(&result)

	for i := 2; i < 5; i++ {
		hFreq := udc.outputFrequency * float64(i)
		if !udc.check(udc.sa.SetSpectrum(hFreq, udc.frequencySpectrum.span, udc.frequencySpectrum.rbw, udc.frequencySpectrum.vbw), "SA: set harmonic spectrum") {
			return
		}

		presResp := udc.sa.CheckIfCarrierIsPresent()
		if !udc.check(presResp, "SA: check carrier") {
			return
		}

		noiseFloor := presResp.Result["MinValue"].Value
		isPresent := presResp.Result["Carrier"].Bool
		levelStr := "NIL"

		if isPresent {
			if !udc.check(udc.sa.PeakSearch(true, 1), "SA: peak search harmonic") {
				return
			}
			udc.sa.WaitForSweeps(5)
			hPower := udc.sa.GetMarkerValue(1).Result["MarkerY"].Value
			hMarkerFreq := udc.sa.GetMarkerValue(1).Result["MarkerX"].Value

			if math.Abs(hMarkerFreq-hFreq) < (hFreq * 4e-6) { // 4ppm tolerance
				levelStr = fmt.Sprintf("%.6f", mainPower-hPower)
			}
		}

		respSpectrum := udc.sa.GetSpectrumDump()
		if !udc.check(respSpectrum, "SA: Spectrum dump") {
			return
		}

		result.HarmonicResultValue.HarmonicNo = append(result.HarmonicResultValue.HarmonicNo, i)
		result.HarmonicResultValue.HarmonicFrequency = append(result.HarmonicResultValue.HarmonicFrequency, fmt.Sprintf("%.6f", hFreq))
		result.HarmonicResultValue.CarrierLevel = append(result.HarmonicResultValue.CarrierLevel, levelStr)
		result.HarmonicResultValue.NoiseFloor = append(result.HarmonicResultValue.NoiseFloor, noiseFloor)
		result.SpectrumDump = append(result.SpectrumDump, respSpectrum.Result["SpectrumDump"].String)

		logger.Log.Info("UDC Harmonics Measurement Point", "harmonicFreq", hFreq, "carrierLevel", levelStr, "noiseFloor", noiseFloor)
		udc.measurementMonitor <- result
		udc.notify(fmt.Sprintf("Harmonics measurement completed for %.2f MHz", hFreq/1e6))
	}

	if udc.save(result, result.TestName) {
		logger.Log.Info("Completed UDC Harmonics Measurement", "success", true)
		udc.finish("Harmonics Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputSpuriousMeasurement(inBand bool) {
	logger.Log.Info("Starting UDC Spurious Measurement", "inBand", inBand)
	udc.notify("Spurious Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}
	if !udc.check(udc.sg.SetPower(udc.inputPower+udc.inputCableLoss), "SG: set power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}

	spec := udc.outBandSpectrum
	testName, testCode := "Output Port - Spurious Measurement - Out of Band", UCDCSpuriousOutBand
	if inBand {
		spec = udc.inBandSpectrum
		testName, testCode = "Output Port - Spurious Measurement - In Band", UCDCSpuriousInBand
	}

	result := ConvertorResults{TestName: testName, TestCode: testCode, SpuriousResults: true}
	udc.populateFreqs(&result)

	if !udc.check(udc.sa.SetSpectrum(udc.outputFrequency, spec.span, spec.rbw, spec.vbw), "SA: set spectrum") {
		return
	}
	resp := udc.sa.SetReferenceNominal()
	if !udc.check(resp, "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	noiseFloor := resp.Result["MinValue"].Value
	carrierPower := resp.Result["MaxValue"].Value
	_ = noiseFloor

	resp = udc.sa.GetAllPeaksAbove(10, 1)
	if !udc.check(resp, "SA: get all peaks") {
		return
	}
	respScreendump := udc.sa.GetSpectrumDump()
	if !udc.check(respScreendump, "SA: Taking Spectrum dump") {
		return
	}

	frequencies := resp.Result["Frequencies"].Values
	peaks := resp.Result["Peaks"].Values

	expectedCarrier := udc.outputFrequency
	allowedDeviation := math.Max(spec.rbw, spec.span/500)

	// Identify carrier peak(s) within allowed frequency tolerance
	carrierIdx := -1
	removeAtIndex := make([]int, 0)
	for i := 0; i < len(frequencies); i++ {
		if math.Abs(frequencies[i]-expectedCarrier) <= allowedDeviation {
			removeAtIndex = append(removeAtIndex, i)
			if carrierIdx == -1 || peaks[i] > peaks[carrierIdx] {
				carrierIdx = i
			}
		}
	}

	if carrierIdx != -1 {
		carrierPower = peaks[carrierIdx]
	}

	// Filter out expected carrier peak(s) to isolate spurious components
	var spurIndices []int
	for i := 0; i < len(frequencies); i++ {
		if !slices.Contains(removeAtIndex, i) {
			spurIndices = append(spurIndices, i)
		}
	}

	if len(spurIndices) == 0 {
		result.SpuriousResultValue.Frequency = []string{"NIL"}
		result.SpuriousResultValue.MeasuredPowerdBm = []string{"NIL"}
		result.SpuriousResultValue.SpuriousLeveldBC = []string{"NIL"}
	} else {
		result.SpuriousResultValue.Frequency = make([]string, len(spurIndices))
		result.SpuriousResultValue.MeasuredPowerdBm = make([]string, len(spurIndices))
		result.SpuriousResultValue.SpuriousLeveldBC = make([]string, len(spurIndices))
		for j, i := range spurIndices {
			levelDbc := peaks[i] - carrierPower
			logger.Log.Info("UDC Spurious Measurement Point", "freq", frequencies[i], "power", peaks[i], "levelDbc", levelDbc)
			result.SpuriousResultValue.Frequency[j] = fmt.Sprintf("%.6f", frequencies[i])
			result.SpuriousResultValue.MeasuredPowerdBm[j] = fmt.Sprintf("%.2f", peaks[i])
			result.SpuriousResultValue.SpuriousLeveldBC[j] = fmt.Sprintf("%.2f", levelDbc)
		}
	}

	result.SpectrumDump = []string{respScreendump.Result["SpectrumDump"].String}
	udc.measurementMonitor <- result

	if udc.save(result, testName) {
		logger.Log.Info("Completed UDC Spurious Measurement", "success", true)
		udc.finish("Spurious Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) LOLeakageMeasurement() {
	logger.Log.Info("Starting UDC LO Leakage Measurement")
	udc.notify("LO Leakage Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.check(udc.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !udc.check(udc.sg.SetModOff(), "SG: mod off") {
		return
	}
	if !udc.check(udc.sg.SetRFOff(), "SG: RF off") {
		return
	}

	targetFreq := udc.loFrequency()
	if !udc.check(udc.sa.SetSpectrum(targetFreq, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
		return
	}

	resp := udc.sa.CheckIfCarrierIsPresent()
	if !udc.check(resp, "SA: check carrier") {
		return
	}
	time.Sleep(time.Second)

	leakage := resp.Result["MaxValue"].Value + udc.outputCableLoss[2]
	result := ConvertorResults{
		TestName: "Output Port - LO Leakage Measurement", TestCode: UCDCLOLeakage, PowerOrLeakageResults: true,
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: targetFreq, Power: leakage},
	}
	udc.populateFreqs(&result)

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	logger.Log.Info("Completed UDC LO Leakage Measurement", "leakage", leakage)
	udc.measurementMonitor <- result
	if udc.save(result, result.TestName) {
		udc.finish("LO Leakage Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputInputLeakageMeasurement() {
	logger.Log.Info("Starting UDC Input Leakage Measurement")
	udc.notify("Input Leakage Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}
	if !udc.check(udc.sa.SetSpectrum(udc.inputFrequency, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sg.SetPower(udc.inputPower+udc.inputCableLoss), "SG: set power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}

	resp := udc.sa.CheckIfCarrierIsPresent()
	if !udc.check(resp, "SA: search carrier") {
		return
	}
	if !udc.check(udc.sa.PeakSearch(true, 1), "SA: peak search") {
		return
	}

	power := udc.sa.GetMarkerValue(1).Result["MarkerY"].Value + udc.outputCableLoss[1]
	result := ConvertorResults{
		TestName: "Output Port - Input Leakage Measurement", TestCode: UCDCInputLeakage, PowerOrLeakageResults: true,
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: udc.inputFrequency, Power: power},
	}
	udc.populateFreqs(&result)
	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	logger.Log.Info("Completed UDC Input Leakage Measurement", "leakage", power)
	udc.measurementMonitor <- result
	if udc.save(result, result.TestName) {
		udc.finish("Input Leakage Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputExtLOGainMeasurement(stepSize float64, cable bool) {
	logger.Log.Info("Starting UDC External LO Gain Measurement", "cable", cable)
	udc.notify("External LO Gain Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()
	defer udc.sgExt.SetRFOff()

	loPower, err := udc.getExtLOPower()
	if err != nil {
		udc.setError("External LO Power Matching required before gain measurement")
		return
	}

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}
	loFreq := udc.loFrequency()
	if !udc.check(udc.sgExt.SetFrequency(loFreq), "SG Ext: frequency set") {
		return
	}
	if !udc.check(udc.sgExt.SetPower(loPower+udc.loCableLoss), "SG Ext: power set") {
		return
	}
	if !udc.check(udc.sgExt.SetModOff(), "SG Ext: mod off") {
		return
	}
	if !udc.check(udc.sgExt.SetRFOn(), "SG Ext: RF on") {
		return
	}

	maxPower, minPower := udc.converter.MaxPowerCable, udc.converter.MinPowerCable
	testName, testCode := "Output Port - Gain Measurement - External LO - Cable", UCDCGainExternalCable
	if !cable {
		maxPower, minPower = udc.converter.MaxPowerRadiated.Float64, udc.converter.MinPowerRadiated.Float64
		testName, testCode = "Output Port - Gain Measurement - External LO - Radiated", UCDCGainExternalRadiated
	}

	result := ConvertorResults{TestName: testName, TestCode: testCode, GainResults: true}
	udc.populateFreqs(&result)
	if !udc.check(udc.sa.SetSpectrum(udc.outputFrequency, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sg.SetPower(minPower-udc.inputCableLoss), "SG: set start power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	for p := minPower; p <= maxPower; p += stepSize {
		if !udc.check(udc.sg.SetPower(p+udc.inputCableLoss), "SG: iterate power") {
			return
		}
		time.Sleep(time.Second)
		if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier loop") {
			return
		}
		time.Sleep(time.Second)

		resp := udc.sa.GetMaxMarkerValue()
		if !udc.check(resp, "SA: read power") {
			return
		}

		if p == minPower {
			respSpectrum := udc.sa.GetSpectrumDump()
			if !udc.check(respSpectrum, "SA: Spectrum dump") {
				return
			}
			result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}
		}

		outPower := resp.Result["MarkerY"].Value + udc.outputCableLoss[1]
		result.GainResultValue.SetPower = append(result.GainResultValue.SetPower, p)
		result.GainResultValue.OutputPower = append(result.GainResultValue.OutputPower, outPower)
		result.GainResultValue.Gain = append(result.GainResultValue.Gain, outPower-p)
		result.GainResultValue.AverageGain = mean(result.GainResultValue.Gain)

		logger.Log.Info("UDC External LO Gain Measurement Point", "setPower", p, "outPower", outPower, "gain", outPower-p)
		udc.measurementMonitor <- result
		udc.notify(fmt.Sprintf("Completed for %.3f dBm", p))
	}

	if udc.save(result, testName) {
		logger.Log.Info("Completed UDC External LO Gain Measurement", "success", true)
		udc.finish("External LO Gain Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) MonitorPowerMeasurement(isOutput bool) {
	testName, testCode := "Input Monitor Port - Power Measurement", UCDCInputMonPower
	targetFreq := udc.inputFrequency
	if isOutput {
		testName, testCode = "Output Monitor Port - Power Measurement", UCDCOutputMonPower
		targetFreq = udc.outputFrequency
	}
	logger.Log.Info("Starting UDC Monitor Power Measurement", "isOutput", isOutput)
	udc.notify(testName + " Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.inputFrequency, udc.outputFrequency) {
		return
	}
	if !udc.check(udc.sa.SetSpectrum(targetFreq, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sg.SetPower(udc.inputPower+udc.inputCableLoss), "SG: set power") {
		return
	}
	if !udc.check(udc.sg.SetRFOn(), "SG: RF on") {
		return
	}

	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)
	if !udc.check(udc.sa.PeakSearch(true, 1), "SA: peak search") {
		return
	}
	udc.sa.WaitForSweeps(5)
	power := 0.0
	if isOutput {
		power = udc.sa.GetMarkerValue(1).Result["MarkerY"].Value + udc.outputCableLoss[1]
	} else {
		power = udc.sa.GetMarkerValue(1).Result["MarkerY"].Value + udc.outputCableLoss[0]
	}
	result := ConvertorResults{
		TestName: testName, TestCode: testCode, PowerOrLeakageResults: true,
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: targetFreq, Power: power},
	}
	udc.populateFreqs(&result)

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	logger.Log.Info("Completed UDC Monitor Power Measurement", "power", power)
	udc.measurementMonitor <- result
	if udc.save(result, testName) {
		udc.finish(testName+" Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) LOMonFreqPowerMeasurement() {
	logger.Log.Info("Starting UDC LO Monitor Freq Power Measurement")
	udc.notify("LO MON Port Frequency & Power Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.check(udc.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !udc.check(udc.sg.SetModOff(), "SG: mod off") {
		return
	}

	targetlo := udc.loFrequency()
	if !udc.check(udc.sa.SetSpectrum(targetlo, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	if !udc.check(udc.sa.PeakSearch(true, 1), "SA: peak search") {
		return
	}
	udc.sa.WaitForSweeps(5)
	power := udc.sa.GetMarkerValue(1).Result["MarkerY"].Value
	power = power + udc.outputCableLoss[2]

	resp := udc.sa.GetFrequencyInCounterMode(1)
	if !udc.check(resp, "SA: counter mode") {
		return
	}
	freq := resp.Result["Frequency"].Value

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}

	result := ConvertorResults{
		TestName: "LO MON Port Frequency & Power Measurement", TestCode: UCDCLOMonPower, PowerOrLeakageResults: true, FrequencyResults: true,
		FrequencyResultValue:      FrequencyResults{ExpectedFrequency: targetlo, MeasuredFrequency: freq, Deviation: math.Abs(freq - targetlo)},
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: freq, Power: power},
	}
	udc.populateFreqs(&result)
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}
	logger.Log.Info("Completed UDC LO Monitor Freq Power Measurement", "power", power, "freq", freq)
	udc.measurementMonitor <- result

	if udc.save(result, "LO MON Port - Frequency & Power Measurement") {
		udc.finish("LO Mon Power & Frequency Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) LOMonPhaseNoiseMeasurement() {
	logger.Log.Info("Starting UDC LO Monitor Phase Noise Measurement")
	udc.notify("LO Mon Port Phase Noise Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()
	defer udc.sa.SetSAMode()

	if !udc.check(udc.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !udc.check(udc.sg.SetModOff(), "SG: mod off") {
		return
	}

	targetlo := udc.loFrequency()
	if !udc.check(udc.sa.SetSpectrum(targetlo, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: spectrum set") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	if !udc.check(udc.sa.SetPhaseNoiseMeasurement(), "SA: phase noise mode") {
		return
	}

	time.Sleep(5 * time.Second)

	result := ConvertorResults{TestName: "LO Mon Port Phase Noise Measurement", TestCode: UCDCLOMonPhaseNoise, PhaseNoiseResults: true}
	udc.populateFreqs(&result)

	for i := 1000; i <= 1000000; i *= 10 {
		if !udc.check(udc.sa.SetMarkerValuePhaseNoise(float64(i), 1), "SA: PN marker set") {
			return
		}
		resp := udc.sa.GetPhaseNoiseMarkerY(1)
		if !udc.check(resp, "SA: PN read") {
			return
		}

		val := resp.Result["MarkerY"].Value
		logger.Log.Info("UDC LO Monitor Phase Noise Point", "freq", float64(i), "phaseNoise", val)
		result.PhaseNoiseResultValue.Frequency = append(result.PhaseNoiseResultValue.Frequency, float64(i))
		result.PhaseNoiseResultValue.PhaseNoise = append(result.PhaseNoiseResultValue.PhaseNoise, val)
	}

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	udc.measurementMonitor <- result

	if udc.save(result, "LO MON Port - Phase Noise Measurement") {
		logger.Log.Info("Completed UDC LO Monitor Phase Noise Measurement", "success", true)
		udc.finish("LO Mon Phase Noise Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) ExtLOPowerMatch() {
	logger.Log.Info("Starting UDC External LO Power Match")
	udc.notify("External LO Power Matching Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()
	defer udc.sgExt.SetRFOff()

	loFreq := udc.loFrequency()
	if !udc.check(udc.sgExt.SetFrequency(loFreq), "SG Ext: frequency set") {
		return
	}

	loPower, err := udc.getLOPower()
	if err != nil {
		udc.setError("LO Power measurement required before power matching")
		return
	}
	loPowerSA := loPower - udc.outputCableLoss[2]

	if !udc.check(udc.sgExt.SetPower(loPower+udc.loCableLoss), "SG Ext: initial power") {
		return
	}
	if !udc.check(udc.sgExt.SetModOff(), "SG Ext: mod off") {
		return
	}
	if !udc.check(udc.sgExt.SetRFOn(), "SG Ext: RF on") {
		return
	}

	if !udc.check(udc.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !udc.check(udc.sa.SetSpectrum(loFreq, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: spectrum set") {
		return
	}
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	resp := udc.sa.GetMaxMarkerValue()
	if !udc.check(resp, "SA: read marker") {
		return
	}

	extLoMeasured := resp.Result["MarkerY"].Value
	powerDev := loPowerSA - extLoMeasured
	powerSet := loPowerSA + udc.loCableLoss

	// Matching Loop: iterate until deviation is within 0.1 dB
	for math.Abs(powerDev) > 0.1 {
		if udc.stop {
			udc.setError("Aborted")
			return
		}

		powerSet += powerDev
		if !udc.check(udc.sgExt.SetPower(powerSet), "SG Ext: adjust power") {
			return
		}
		time.Sleep(time.Second)

		mResp := udc.sa.GetMaxMarkerValue()
		if !udc.check(mResp, "SA: re-read marker") {
			return
		}
		extLoMeasured = mResp.Result["MarkerY"].Value
		powerDev = loPowerSA - extLoMeasured
		logger.Log.Info("UDC External LO Power Match Point", "extLoMeasured", extLoMeasured, "powerDev", powerDev)
	}

	result := ConvertorResults{
		TestName: "Output Port - Ext LO Power Matching", TestCode: UCDCExtLOPowerMatch, PowerMatchingResults: true,
		PowerMatchingResultValue: PowerMatchingResults{
			InternalLOPowerMeasured: loPower,
			ExternalLOPowerMeasured: extLoMeasured + udc.outputCableLoss[2],
			ExternalSGPowerSet:      powerSet - udc.loCableLoss,
		},
	}
	udc.populateFreqs(&result)

	udc.measurementMonitor <- result
	if udc.save(result, "External LO Power Matching") {
		logger.Log.Info("Completed UDC External LO Power Match", "success", true)
		udc.finish("External LO Power Matching Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) getLOPower() (float64, error) {
	r, err := resultsDB.GetUpDownConverterResult(udc.converter.Name, "LO MON Port - Frequency & Power Measurement")
	if err != nil {
		return 0, err
	}
	var res ConvertorResults
	if err := json.Unmarshal([]byte(r.Results), &res); err != nil {
		return 0, err
	}
	return res.PowerOrLeakageResultValue.Power, nil
}

func (udc *UpDownConverterMeasurement) getExtLOPower() (float64, error) {
	r, err := resultsDB.GetUpDownConverterResult(udc.converter.Name, "External LO Power Matching")
	if err != nil {
		return 0, err
	}
	var res ConvertorResults
	if err := json.Unmarshal([]byte(r.Results), &res); err != nil {
		return 0, err
	}
	return res.PowerMatchingResultValue.ExternalSGPowerSet, nil
}

func mean(v []float64) float64 {
	var s float64
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

func (udc *UpDownConverterMeasurement) getResult(date string, time string) (ConvertorResults, error) {
	r, err := resultsDB.GetUpDownConverterResultWithDateAndTime(udc.converterName, date, time)
	if err != nil {
		return ConvertorResults{}, err
	}
	var res ConvertorResults
	if err := json.Unmarshal([]byte(r.Results), &res); err != nil {
		return ConvertorResults{}, err
	}
	return res, nil
}
