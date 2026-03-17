package tne

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/reports"
	"prismServer/resultsDB"
	"prismServer/utils"
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
	udc.statusMonitor <- RTStatus{Message: msg, Success: true}
}

func (udc *UpDownConverterMeasurement) finish(msg string, success bool) {
	udc.statusMonitor <- RTStatus{
		Message:   msg,
		Success:   success,
		Error:     !success,
		Completed: true,
	}
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) setError(msg string) {
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

func (udc *UpDownConverterMeasurement) save(result ConvertorResults, testType string) bool {
	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		udc.setError("Failed to process results")
		return false
	}
	if err := resultsDB.InsertUpDownConverterResult(udc.converterName, testType, string(data)); err != nil {
		udc.setError("Failed to save results to database")
		return false
	}
	return true
}

func (udc *UpDownConverterMeasurement) GeneratePDF(name string, dates, times []string) (string, bool) {
	var report reports.Report
	testPhase, _ := database.GetSelectedTestPhase()
	report.SetHeader(name, "Up/Down Converter Results", "", testPhase)
	var avail = make(map[int]string)
	udc.converterName = name
	for i := 0; i < len(dates); i++ {
		result, err := udc.getResult(dates[i], times[i])
		if err != nil {
			return "", false
		}
		switch result.TestCode {
		case UCDCGainInternalCable:
			table := result.GainResultValue.getResultTable()
			reportName := "Output Port - Gain Measurement - Internal LO - Cable Mode"
			avail[0] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCGainInternalRadiated:
			table := result.GainResultValue.getResultTable()
			reportName := "Output Port - Gain Measurement - Internal LO - Radiated Mode"
			avail[1] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCFreqMeas:
			table := result.FrequencyResultValue.getResultTable()
			reportName := "Output Port - Frequency Measurement"
			avail[2] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCHarmonicMeas:
			table := result.HarmonicResultValue.getResultTable()
			reportName := "Output Port - Harmonics Measurement"
			avail[3] = reportName
			for i := 0; i < len(result.HarmonicResultValue.HarmonicNo); i++ {
				report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[i], Caption: fmt.Sprintf("%s - %d Harmonic", reportName, result.HarmonicResultValue.HarmonicNo[i])})
			}
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCSpuriousInBand:
			table := result.SpuriousResultValue.getResultTable()
			reportName := "Output Port - Spurious - In Band"
			avail[4] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCSpuriousOutBand:
			table := result.SpuriousResultValue.getResultTable()
			reportName := "Output Port - Spurious - Out of Band"
			avail[5] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCLOLeakage:
			table := result.PowerOrLeakageResultValue.getResultTable()
			reportName := "Output Port - LO Leakage"
			avail[6] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCInputLeakage:
			table := result.PowerOrLeakageResultValue.getResultTable()
			reportName := "Output Port - Input Leakage"
			avail[7] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCGainExternalCable:
			table := result.GainResultValue.getResultTable()
			reportName := "Output Port - Gain Measurement - External LO - Cable Mode"
			avail[8] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCGainExternalRadiated:
			table := result.GainResultValue.getResultTable()
			reportName := "Output Port - Gain Measurement - External LO - Radiated Mode"
			avail[9] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCOutputMonPower:
			table := result.PowerOrLeakageResultValue.getResultTable()
			reportName := "Output Monitor Port - Power Measurement"
			avail[10] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCInputMonPower:
			table := result.PowerOrLeakageResultValue.getResultTable()
			reportName := "Input Monitor Port - Power Measurement"
			avail[11] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCLOMonPower:
			table := result.PowerOrLeakageResultValue.getResultTable()
			reportName := "LO Monitor Port - Power Measurement"
			avail[12] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
			table2 := result.FrequencyResultValue.getResultTable()
			reportName2 := "LO Monitor Port - Frequency Measurement"
			avail[13] = reportName2
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName2})
			report.SetResults(reportName2, table2.Header, table2.Data)
		case UCDCLOMonPhaseNoise:
			table := result.PhaseNoiseResultValue.getResultTable()
			reportName := "LO Monitor Port - Phase Noise Measurement"
			avail[14] = reportName
			report.Screenshots = append(report.Screenshots, reports.Images{FileData: result.SpectrumDump[0], Caption: reportName})
			report.SetResults(reportName, table.Header, table.Data)
		case UCDCExtLOPowerMatch:
			table := result.PowerMatchingResultValue.getResultTable()
			reportName := "LO Monitor Port - External LO Power Match"
			avail[15] = reportName
			report.SetResults(reportName, table.Header, table.Data)
		}
	}
	order := make([]string, 0)
	for i := 0; i < 16; i++ {
		reportName, ok := avail[i]
		if ok {
			order = append(order, reportName)
		}
	}
	report.SetOrder(order)
	pdf, err := reports.GenerateResult(report, true, false, false, false, true)
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

func (udc *UpDownConverterMeasurement) Init(deviceProfile, externalSGName, converterName string) {
	udc.deviceProfile = deviceProfile
	udc.externalSGName = externalSGName
	udc.converterName = converterName

	ucdc, err := database.GetConverterDetails(converterName)
	if err != nil {
		udc.setError("Unable to read converter from database")
		return
	}
	udc.converter = ucdc

	saName, okSa := database.GetSAFromDeviceProfile(deviceProfile)
	sgName, okSg := database.GetSGFromDeviceProfile(deviceProfile)
	if !okSa || !okSg {
		udc.setError("Unable to resolve devices from profile")
		return
	}

	if strings.EqualFold(sgName, externalSGName) {
		udc.setError("External LO SG and Internal LO SG cannot be the same")
		return
	}

	if !udc.sa.LoadDevice(saName) || !udc.sg.LoadDevice(sgName) || !udc.sgExt.LoadDevice(externalSGName) {
		udc.setError("Failed to load device drivers")
		return
	}
	udc.stop = false
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
	udc.notify("Gain Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
		return
	}

	maxPower, minPower := udc.converter.MaxPowerCable, udc.converter.MinPowerCable
	testName, testCode := "Output Port - Gain Measurement - Internal LO - Cable", UCDCGainInternalCable
	if !cable {
		maxPower, minPower = udc.converter.MaxPowerRadiated.Float64, udc.converter.MinPowerRadiated.Float64
		testName, testCode = "Output Port - Gain Measurement - Internal LO - Radiated", UCDCGainInternalRadiated
	}

	result := ConvertorResults{TestName: testName, TestCode: testCode, GainResults: true}

	if !udc.check(udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
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

		udc.measurementMonitor <- result
		udc.notify(fmt.Sprintf("Completed measurement for %.3f dBm", p))
	}

	if udc.save(result, testName) {
		udc.finish("Gain Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputFrequencyMeasurement() {
	udc.notify("Frequency Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
		return
	}

	if !udc.check(udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.frequencySpectrum.span, udc.frequencySpectrum.rbw, udc.frequencySpectrum.vbw), "SA: set spectrum") {
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

	resp := udc.sa.GetFrequencyInCounterMode(1)
	if !udc.check(resp, "SA: frequency counter") {
		return
	}

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(resp, "SA: Spectrum dump") {
		return
	}

	freq := resp.Result["Frequency"].Value
	deviation := math.Abs(freq - udc.converter.OutputFrequency)

	result := ConvertorResults{
		TestName: "Output Port - Frequency Measurement", TestCode: UCDCFreqMeas, FrequencyResults: true,
		FrequencyResultValue: FrequencyResults{ExpectedFrequency: udc.converter.OutputFrequency, MeasuredFrequency: freq, Deviation: deviation},
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	udc.measurementMonitor <- result
	if udc.save(result, result.TestName) {
		udc.finish("Frequency Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputHarmonicsMeasurement() {
	udc.notify("Harmonics Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
		return
	}

	if !udc.check(udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.frequencySpectrum.span, udc.frequencySpectrum.rbw, udc.frequencySpectrum.vbw), "SA: set spectrum") {
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

	for i := 2; i < 5; i++ {
		hFreq := udc.converter.OutputFrequency * float64(i)
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

		udc.measurementMonitor <- result
		udc.notify(fmt.Sprintf("Harmonics measurement completed for %.2f MHz", hFreq/1e6))
	}

	if udc.save(result, result.TestName) {
		udc.finish("Harmonics Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputSpuriousMeasurement(inBand bool) {
	udc.notify("Spurious Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
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

	if !udc.check(udc.sa.SetSpectrum(udc.converter.OutputFrequency, spec.span, spec.rbw, spec.vbw), "SA: set spectrum") {
		return
	}
	resp := udc.sa.SetReferenceNominal()
	if !udc.check(resp, "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	noiseFloor := resp.Result["MinValue"].Value
	carrierPower := resp.Result["MaxValue"].Value

	if !udc.check(udc.sa.SetPeakThresholdAndExcursion(noiseFloor+15, 1), "SA: excursion") {
		return
	}
	if !udc.check(udc.sa.SetMaxHold(), "SA: max hold") {
		return
	}
	if !udc.check(udc.sa.WaitForSweeps(10), "SA: Waiting for sweeps") {
		return
	}

	if !udc.check(udc.sa.GetTraceDump(1001), "SA: Taking trace dump") {
		return
	}
	respScreendump := udc.sa.GetSpectrumDump()
	if !udc.check(respScreendump, "SA: Taking Spectrum dump") {
		return
	}

	resp = udc.sa.GetAllPeaksAbove(noiseFloor+10, 1)
	if !udc.check(resp, "SA: get all peaks") {
		return
	}

	frequencies := resp.Result["Frequencies"].Values
	peaks := resp.Result["Peaks"].Values

	if len(frequencies) == 1 {
		result.SpuriousResultValue.Frequency = []string{"NIL"}
		result.SpuriousResultValue.MeasuredPowerdBm = []string{"NIL"}
		result.SpuriousResultValue.SpuriousLeveldBC = []string{"NIL"}
	} else {
		result.SpuriousResultValue.Frequency = make([]string, len(frequencies)-1)
		result.SpuriousResultValue.MeasuredPowerdBm = make([]string, len(frequencies)-1)
		result.SpuriousResultValue.SpuriousLeveldBC = make([]string, len(frequencies)-1)
		for i := 1; i < len(frequencies); i++ {
			result.SpuriousResultValue.Frequency[i-1] = fmt.Sprintf("%.6f", frequencies[i])
			result.SpuriousResultValue.MeasuredPowerdBm[i-1] = fmt.Sprintf("%.2f", peaks[i])
			result.SpuriousResultValue.SpuriousLeveldBC[i-1] = fmt.Sprintf("%.2f", carrierPower-peaks[i])
		}
	}
	result.SpectrumDump = []string{respScreendump.Result["SpectrumDump"].String}
	udc.measurementMonitor <- result

	if udc.save(result, testName) {
		udc.finish("Spurious Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) LOLeakageMeasurement() {
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

	targetFreq := math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
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

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	udc.measurementMonitor <- result
	if udc.save(result, result.TestName) {
		udc.finish("LO Leakage Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputInputLeakageMeasurement() {
	udc.notify("Input Leakage Measurement Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
		return
	}
	if !udc.check(udc.sa.SetSpectrum(udc.converter.InputFrequency, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
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
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: udc.converter.InputFrequency, Power: power},
	}
	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	udc.measurementMonitor <- result
	if udc.save(result, result.TestName) {
		udc.finish("Input Leakage Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) OutputExtLOGainMeasurement(stepSize float64, cable bool) {
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

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
		return
	}
	loFreq := math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
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
	if !udc.check(udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.powerSpectrum.span, udc.powerSpectrum.rbw, udc.powerSpectrum.vbw), "SA: set spectrum") {
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

		udc.measurementMonitor <- result
		udc.notify(fmt.Sprintf("Completed for %.3f dBm", p))
	}

	if udc.save(result, testName) {
		udc.finish("External LO Gain Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) MonitorPowerMeasurement(isOutput bool) {
	testName, testCode := "Input Monitor Port - Power Measurement", UCDCInputMonPower
	targetFreq := udc.converter.InputFrequency
	if isOutput {
		testName, testCode = "Output Monitor Port - Power Measurement", UCDCOutputMonPower
		targetFreq = udc.converter.OutputFrequency
	}
	udc.notify(testName + " Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()

	if !udc.setupBasic(udc.converter.InputFrequency, udc.converter.OutputFrequency) {
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

	respSpectrum := udc.sa.GetSpectrumDump()
	if !udc.check(respSpectrum, "SA: Spectrum dump") {
		return
	}
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}

	udc.measurementMonitor <- result
	if udc.save(result, testName) {
		udc.finish(testName+" Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) LOMonFreqPowerMeasurement() {
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

	targetlo := math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
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
	result.SpectrumDump = []string{respSpectrum.Result["SpectrumDump"].String}
	udc.measurementMonitor <- result

	if udc.save(result, "LO MON Port - Frequency & Power Measurement") {
		udc.finish("LO Mon Power & Frequency Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) LOMonPhaseNoiseMeasurement() {
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

	targetlo := math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
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

	for i := 1000; i <= 1000000; i *= 10 {
		if !udc.check(udc.sa.SetMarkerValuePhaseNoise(float64(i), 1), "SA: PN marker set") {
			return
		}
		resp := udc.sa.GetPhaseNoiseMarkerY(1)
		if !udc.check(resp, "SA: PN read") {
			return
		}

		val := resp.Result["MarkerY"].Value
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
		udc.finish("LO Mon Phase Noise Measurement Completed", true)
	}
}

func (udc *UpDownConverterMeasurement) ExtLOPowerMatch() {
	udc.notify("External LO Power Matching Started")
	defer udc.sa.SetAlignmentOn()
	defer udc.sa.SystemPreset()
	defer udc.sg.SetRFOff()
	defer udc.sgExt.SetRFOff()

	loFreq := math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
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
	}

	result := ConvertorResults{
		TestName: "Output Port - Ext LO Power Matching", TestCode: UCDCExtLOPowerMatch, PowerMatchingResults: true,
		PowerMatchingResultValue: PowerMatchingResults{
			InternalLOPowerMeasured: loPower,
			ExternalLOPowerMeasured: extLoMeasured + udc.outputCableLoss[2],
			ExternalSGPowerSet:      powerSet - udc.loCableLoss,
		},
	}

	udc.measurementMonitor <- result
	if udc.save(result, "External LO Power Matching") {
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
