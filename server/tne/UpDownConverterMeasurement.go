package tne

import (
	"encoding/json"
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/driver"
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

	freq := resp.Result["Frequency"].Value
	deviation := math.Abs(freq - udc.converter.OutputFrequency)

	result := ConvertorResults{
		TestName: "Output Port - Frequency Measurement", TestCode: UCDCFreqMeas, FrequencyResults: true,
		FrequencyResultValue: FrequencyResults{ExpectedFrequency: udc.converter.OutputFrequency, MeasuredFrequency: freq, Deviation: deviation},
	}

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

		result.HarmonicResultValue.HarmonicNo = append(result.HarmonicResultValue.HarmonicNo, i)
		result.HarmonicResultValue.HarmonicFrequency = append(result.HarmonicResultValue.HarmonicFrequency, fmt.Sprintf("%.6f", hFreq))
		result.HarmonicResultValue.CarrierLevel = append(result.HarmonicResultValue.CarrierLevel, levelStr)
		result.HarmonicResultValue.NoiseFloor = append(result.HarmonicResultValue.NoiseFloor, noiseFloor)

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
	if !udc.check(udc.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	noiseFloor := udc.sa.SetReferenceNominal().Result["MinValue"].Value
	carrierPower := udc.sa.SetReferenceNominal().Result["MaxValue"].Value + udc.outputCableLoss[0]

	if !udc.check(udc.sa.SetPeakThresholdAndExcursion(noiseFloor+15, 1), "SA: excursion") {
		return
	}
	if !udc.check(udc.sa.SetMaxHold(), "SA: max hold") {
		return
	}
	udc.sa.WaitForSweeps(5)

	resp := udc.sa.GetMaxMarkerValue()
	if !udc.check(resp, "SA: markers") {
		return
	}
	prevFreq := resp.Result["MarkerX"].Value

	for {
		if !udc.check(udc.sa.SetMarkerNextPeak(1), "SA: next peak") {
			return
		}
		mResp := udc.sa.GetMarkerValue(1)
		if !udc.check(mResp, "SA: get marker") {
			return
		}

		f, v := mResp.Result["MarkerX"].Value, mResp.Result["MarkerY"].Value+udc.outputCableLoss[0]
		if f == prevFreq {
			break
		}

		result.SpuriousResultValue.Frequency = append(result.SpuriousResultValue.Frequency, fmt.Sprintf("%.6f", f))
		result.SpuriousResultValue.MeasuredPowerdBm = append(result.SpuriousResultValue.MeasuredPowerdBm, fmt.Sprintf("%.6f", v))
		result.SpuriousResultValue.SpuriousLeveldBC = append(result.SpuriousResultValue.SpuriousLeveldBC, fmt.Sprintf("%.6f", carrierPower-v))

		udc.measurementMonitor <- result
		prevFreq = f
	}

	if len(result.SpuriousResultValue.Frequency) == 0 {
		result.SpuriousResultValue.Frequency = []string{"NIL"}
		result.SpuriousResultValue.MeasuredPowerdBm = []string{"NIL"}
		result.SpuriousResultValue.SpuriousLeveldBC = []string{"NIL"}
		udc.measurementMonitor <- result
	}

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
	if !udc.check(udc.sgExt.SetPower(loPower-udc.loCableLoss), "SG Ext: power set") {
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

	power := udc.sa.GetMarkerValue(1).Result["MarkerY"].Value + udc.outputCableLoss[0]
	result := ConvertorResults{
		TestName: testName, TestCode: testCode, PowerMatchingResults: true,
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: targetFreq, Power: power},
	}

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

	resp := udc.sa.GetFrequencyInCounterMode(1)
	if !udc.check(resp, "SA: counter mode") {
		return
	}
	freq := resp.Result["Frequency"].Value

	result := ConvertorResults{
		TestName: "LO MON Port Frequency & Power Measurement", TestCode: UCDCLOMonPower, PowerOrLeakageResults: true, FrequencyResults: true,
		FrequencyResultValue:      FrequencyResults{ExpectedFrequency: targetlo, MeasuredFrequency: freq, Deviation: math.Abs(freq - targetlo)},
		PowerOrLeakageResultValue: PowerOrLeakageResults{Frequency: freq, Power: power},
	}

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

		udc.measurementMonitor <- result
	}

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
	powerDev := loPower - extLoMeasured - udc.loCableLoss
	powerSet := 0.0

	// Matching Loop: iterate until deviation is within 0.1 dB
	for math.Abs(powerDev) > 0.1 {
		if udc.stop {
			udc.setError("Aborted")
			return
		}
		powerSet = loPower + powerDev + udc.loCableLoss
		if !udc.check(udc.sgExt.SetPower(powerSet), "SG Ext: adjust power") {
			return
		}
		time.Sleep(time.Second)

		mResp := udc.sa.GetMaxMarkerValue()
		if !udc.check(mResp, "SA: re-read marker") {
			return
		}
		extLoMeasured = mResp.Result["MarkerY"].Value
		powerDev = loPower - extLoMeasured - udc.loCableLoss
	}

	result := ConvertorResults{
		TestName: "Output Port - Ext LO Power Matching", TestCode: UCDCExtLOPowerMatch, PowerMatchingResults: true,
		PowerMatchingResultValue: PowerMatchingResults{InternalLOPowerMeasured: loPower, ExternalLOPowerMeasured: extLoMeasured + udc.loCableLoss, ExternalSGPowerSet: powerSet},
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
