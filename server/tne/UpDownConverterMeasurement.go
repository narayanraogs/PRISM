package tne

import (
	"encoding/json"
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/resultsDB"
	"strconv"
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
	currentStatus      [][]string
	statusMonitor      chan RTStatus
	measurementMonitor chan ConvertorResults
	powerSpectrum      frequencyProfile
	frequencySpectrum  frequencyProfile
	inBandSpectrum     frequencyProfile
	outBandSpectrum    frequencyProfile
	stop               bool
}

type frequencyProfile struct {
	span float64
	rbw  float64
	vbw  float64
}

type ConvertorResults struct {
	TestName                  string
	TestCode                  string
	GainResults               bool
	FrequencyResults          bool
	HarmonicsResults          bool
	SpuriousResults           bool
	PowerOrLeakageResults     bool
	PhaseNoiseResults         bool
	PowerMatchingResults      bool
	GainResultValue           GainResults
	FrequencyResultValue      FrequencyResults
	HarmonicResultValue       HarmonicResults
	SpuriousResultValue       SpuriousResults
	PowerOrLeakageResultValue PowerOrLeakageResults
	PhaseNoiseResultValue     PhaseNoiseResults
	PowerMatchingResultValue  PowerMatchingResults
}

type GainResults struct {
	SetPower    []float64
	OutputPower []float64
	Gain        []float64
	AverageGain float64
}

type FrequencyResults struct {
	ExpectedFrequency float64
	MeasuredFrequency float64
	Deviation         float64
}

type HarmonicResults struct {
	HarmonicNo        []int
	HarmonicFrequency []string
	CarrierLevel      []string
	NoiseFloor        []float64
}

type SpuriousResults struct {
	Frequency        []string
	MeasuredPowerdBm []string
	SpuriousLeveldBC []string
}

type PowerOrLeakageResults struct {
	Frequency float64
	Power     float64
}

type PhaseNoiseResults struct {
	Frequency  []float64
	PhaseNoise []float64
}

type PowerMatchingResults struct {
	InternalLOPowerMeasured float64
	ExternalLOPowerMeasured float64
	ExternalSGPowerSet      float64
}

func (udc *UpDownConverterMeasurement) GetStatusMonitor() (chan RTStatus, chan ConvertorResults) {
	udc.statusMonitor = make(chan RTStatus, 20)
	udc.measurementMonitor = make(chan ConvertorResults, 20)
	return udc.statusMonitor, udc.measurementMonitor
}

func (udc *UpDownConverterMeasurement) Init(deviceProfile string, externalSGName string, converterName string) {
	udc.deviceProfile = deviceProfile
	udc.externalSGName = externalSGName
	udc.converterName = converterName
	udc.currentStatus = make([][]string, 0)
	ucdc, err := database.GetConverterDetails(converterName)
	if err != nil {
		udc.setError("Unable to Read Converter from Database")
		return
	}
	udc.converter = ucdc

	ok := udc.loadDevices()
	if !ok {
		udc.setError("Devices cannot be loaded")
		return
	}
	sgName, ok := database.GetSGFromDeviceProfile(udc.deviceProfile)
	if !ok {
		udc.setError("Unable to Read SG from Database")
		return
	}
	if strings.EqualFold(sgName, udc.externalSGName) {
		udc.setError("External LO SG and Internal LO SG cannot be same")
		return
	}
	header := make([]string, 0)
	udc.currentStatus = append(udc.currentStatus, header)

	udc.stop = false
}

func (udc *UpDownConverterMeasurement) SetInputCableLoss(inputCableLoss float64, inputPower float64) {
	udc.inputPower = inputPower
	udc.inputCableLoss = inputCableLoss
}

func (udc *UpDownConverterMeasurement) SetOutputCableLoss(outputCableLoss []float64) {
	udc.outputCableLoss = outputCableLoss
}

func (udc *UpDownConverterMeasurement) SetLOCableLoss(LOCableLoss float64) {
	udc.loCableLoss = LOCableLoss
}

func (udc *UpDownConverterMeasurement) SetPowerSpectrum(span float64, rbw float64, vbw float64) {
	udc.powerSpectrum = frequencyProfile{
		span: span,
		rbw:  rbw,
		vbw:  vbw,
	}
}

func (udc *UpDownConverterMeasurement) SetFrequencySpectrum(span float64, rbw float64, vbw float64) {
	udc.frequencySpectrum = frequencyProfile{
		span: span,
		rbw:  rbw,
		vbw:  vbw,
	}
}

func (udc *UpDownConverterMeasurement) SetInBandSpectrum(span float64, rbw float64, vbw float64) {
	udc.inBandSpectrum = frequencyProfile{
		span: span,
		rbw:  rbw,
		vbw:  vbw,
	}
}

func (udc *UpDownConverterMeasurement) SetOutBandSpectrum(span float64, rbw float64, vbw float64) {
	udc.outBandSpectrum = frequencyProfile{
		span: span,
		rbw:  rbw,
		vbw:  vbw,
	}
}

func (udc *UpDownConverterMeasurement) Stop() {
	udc.stop = true
}

func (udc *UpDownConverterMeasurement) loadDevices() bool {
	saName, ok := database.GetSAFromDeviceProfile(udc.deviceProfile)
	if !ok {
		return false
	}
	sgName, ok := database.GetSGFromDeviceProfile(udc.deviceProfile)
	if !ok {
		return false
	}
	ok = udc.sa.LoadDevice(saName)
	if !ok {
		return false
	}
	ok = udc.sg.LoadDevice(sgName)
	if !ok {
		return false
	}
	ok = udc.sgExt.LoadDevice(udc.externalSGName)
	if !ok {
		return false
	}
	return true
}

func (udc *UpDownConverterMeasurement) setError(message string) {
	var measure = RTStatus{
		Completed: true,
		Success:   false,
		Error:     true,
		Message:   message,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) setStatus(message string) {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   message,
	}
	udc.statusMonitor <- measure
}

func (udc *UpDownConverterMeasurement) saveResults(result ConvertorResults, testType string) error {
	var resultString string
	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return err
	}
	resultString = string(data)

	return resultsDB.InsertUpDownConverterResult(udc.converterName, testType, resultString)
}

func (udc *UpDownConverterMeasurement) OutputGainMeasurement(stepSize float64, cable bool) {
	var result ConvertorResults
	result.GainResults = true
	result.GainResultValue = GainResults{
		SetPower:    make([]float64, 0),
		OutputPower: make([]float64, 0),
		Gain:        make([]float64, 0),
	}

	if cable {
		result.TestName = "Output Port - Gain Measurement - Internal LO - Cable"
		result.TestCode = UCDCGainInternalCable
	} else {
		result.TestName = "Output Port - Gain Measurement - Internal LO - Radiated"
		result.TestCode = UCDCGainInternalRadiated
	}

	udc.setStatus("Gain Measurement Started")
	var maxPower, minPower float64
	if cable {
		maxPower = udc.converter.MaxPowerCable
		minPower = udc.converter.MinPowerCable
	} else {
		maxPower = udc.converter.MaxPowerRadiated.Float64
		minPower = udc.converter.MinPowerRadiated.Float64
	}

	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(minPower - udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)

	for powerSet := minPower + udc.inputCableLoss; powerSet <= maxPower+udc.inputCableLoss; powerSet = powerSet + stepSize {
		if udc.stop {
			udc.setError("Measurement Aborted by User")
			return
		}
		powerStr := fmt.Sprintf("%.3f", powerSet)
		response = udc.sg.SetPower(powerSet)
		if !response.Success {
			udc.setError("SG Power Cannot be set to " + powerStr)
			return
		}
		time.Sleep(1000 * time.Millisecond)
		referencePower := powerSet
		response = udc.sa.GetMaxMarkerValue()
		if !response.Success {
			udc.setError("SA Power Cannot be read")
			return
		}
		powerOut := response.Result["MarkerY"].Value + udc.outputCableLoss[0]
		if powerSet == minPower+udc.inputCableLoss || powerSet == maxPower+udc.inputCableLoss {
			response = udc.sa.GetSpectrumDump()
			if !response.Success {
				udc.setError("Unable to get spectrum dump")
				return
			}
		}
		Gain := powerOut - referencePower
		result.GainResultValue.SetPower = append(result.GainResultValue.SetPower, powerSet)
		result.GainResultValue.OutputPower = append(result.GainResultValue.OutputPower, powerOut)
		result.GainResultValue.Gain = append(result.GainResultValue.Gain, Gain)
		result.GainResultValue.AverageGain = mean(result.GainResultValue.Gain)
		udc.measurementMonitor <- result
		udc.setStatus("Completed Measurement for " + powerStr)
	}
	udc.setStatus("Saving Results")
	var err error
	if cable {
		err = udc.saveResults(result, "Output Port - Gain Measurement - Internal LO - Cable")
	} else {
		err = udc.saveResults(result, "Output Port - Gain Measurement - Internal LO - Radiated")
	}
	if err != nil {
		udc.setError("Unable to Save Results")
		return
	}
	var measure = RTStatus{
		Completed: true,
		Success:   true,
		Error:     false,
		Message:   "Gain Measurement Completed",
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func mean(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func (udc *UpDownConverterMeasurement) OutputFrequencyMeasurement() {
	var result ConvertorResults
	result.FrequencyResults = true
	result.TestName = "Output Port - Frequency Measurement"
	result.TestCode = UCDCFreqMeas

	udc.setStatus("Frequency Measurement Started")
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.frequencySpectrum.span,
		float64(udc.frequencySpectrum.rbw), float64(udc.frequencySpectrum.vbw))
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(udc.inputPower + udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = udc.sa.PeakSearch(true, 1)
	if !response.Success {
		udc.setError("Unable to operate SA in maxhold mode")
		return
	}
	udc.sa.WaitForSweeps(5)
	if !response.Success {
		udc.setError("Unable to wait for sweeps")
		return
	}
	response = udc.sa.GetFrequencyInCounterMode(1)
	if !response.Success {
		udc.setError("Unable to get frequency in counter mode")
		return
	}
	frequency := response.Result["MarkerX"].Value
	response = udc.sa.GetSpectrumDump()
	if !response.Success {
		udc.setError("Unable to get spectrum dump")
		return
	}
	freq_deviation := math.Abs(frequency - udc.converter.OutputFrequency)

	result.FrequencyResultValue.ExpectedFrequency = udc.converter.OutputFrequency
	result.FrequencyResultValue.MeasuredFrequency = frequency
	result.FrequencyResultValue.Deviation = freq_deviation
	udc.measurementMonitor <- result
	udc.setStatus("Saving Results")
	udc.saveResults(result, "Output Port - Frequency Measurement")

	var measure = RTStatus{
		Completed: true,
		Success:   true,
		Error:     false,
		Message:   "Frequency Measurement Completed",
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) OutputHarmonicsMeasurement() {
	udc.setStatus("Harmonics Measurement Started")
	var result ConvertorResults
	result.TestName = "Output Port - Harmonics Measurement"
	result.TestCode = UCDCHarmonicMeas
	result.HarmonicsResults = true
	result.HarmonicResultValue = HarmonicResults{
		HarmonicNo:        make([]int, 0),
		HarmonicFrequency: make([]string, 0),
		CarrierLevel:      make([]string, 0),
		NoiseFloor:        make([]float64, 0),
	}
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.frequencySpectrum.span,
		float64(udc.frequencySpectrum.rbw), float64(udc.frequencySpectrum.vbw))
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(udc.inputPower + udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)

	response = udc.sa.PeakSearch(true, 1)
	if !response.Success {
		udc.setError("Unable to operate SA in maxhold mode")
		return
	}
	outputpower := response.Result["MarkerY"].Value

	for i := 2; i < 5; i = i + 1 {
		harmonics := udc.converter.OutputFrequency * float64(i)
		response = udc.sa.SetSpectrum(harmonics, udc.frequencySpectrum.span,
			udc.frequencySpectrum.rbw, udc.frequencySpectrum.vbw)
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
		response = udc.sa.CheckIfCarrierIsPresent()
		if !response.Success {
			udc.setError("carrier present fucntion error")
			return
		}
		noiseFloor := response.Result["MinValue"].Value
		if !response.Result["Carrier"].Bool {
			result.HarmonicResultValue.HarmonicNo = append(result.HarmonicResultValue.HarmonicNo, i+1)
			result.HarmonicResultValue.HarmonicFrequency = append(result.HarmonicResultValue.HarmonicFrequency, strconv.FormatFloat(harmonics, 'f', 6, 64))
			result.HarmonicResultValue.CarrierLevel = append(result.HarmonicResultValue.CarrierLevel, "NIL")
			result.HarmonicResultValue.NoiseFloor = append(result.HarmonicResultValue.NoiseFloor, noiseFloor)
		} else {
			response = udc.sa.PeakSearch(true, 1)
			if !response.Success {
				udc.setError("Unable to operate SA in maxhold mode")
				return
			}
			udc.sa.WaitForSweeps(5)
			harmpower := response.Result["MarkerY"].Value
			harmfreq := response.Result["MarkerX"].Value
			lower := harmonics - (harmonics * 2 * 1e-6)
			higher := harmonics + (harmonics * 2 * 1e-6)
			if lower < harmfreq && harmfreq < higher {
				result.HarmonicResultValue.HarmonicNo = append(result.HarmonicResultValue.HarmonicNo, i+1)
				result.HarmonicResultValue.HarmonicFrequency = append(result.HarmonicResultValue.HarmonicFrequency, strconv.FormatFloat(harmonics, 'f', 6, 64))
				result.HarmonicResultValue.CarrierLevel = append(result.HarmonicResultValue.CarrierLevel, strconv.FormatFloat(outputpower-harmpower, 'f', 6, 64))
				result.HarmonicResultValue.NoiseFloor = append(result.HarmonicResultValue.NoiseFloor, noiseFloor)
			} else {
				result.HarmonicResultValue.HarmonicNo = append(result.HarmonicResultValue.HarmonicNo, i+1)
				result.HarmonicResultValue.HarmonicFrequency = append(result.HarmonicResultValue.HarmonicFrequency, strconv.FormatFloat(harmonics, 'f', 6, 64))
				result.HarmonicResultValue.CarrierLevel = append(result.HarmonicResultValue.CarrierLevel, "NIL")
				result.HarmonicResultValue.NoiseFloor = append(result.HarmonicResultValue.NoiseFloor, noiseFloor)
			}
		}

		udc.measurementMonitor <- result
		udc.setStatus("Harmonics measurement completed for " + fmt.Sprintf("%.2f MHz", harmonics/1e6))
	}
	udc.setStatus("Saving Result")
	udc.saveResults(result, "Output Port - Harmonics Measurement")
	udc.setStatus("Harmonics Measurement Completed")
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) OutputSpuriousMeasurement(inBand bool) {
	udc.setStatus("Spurious Measurement Started")
	var result ConvertorResults
	if inBand {
		result.TestName = "Output Port - Spurious Measurement - In Band"
		result.TestCode = UCDCSpuriousInBand
	} else {
		result.TestName = "Output Port - Spurious Measurement - Out of Band"
		result.TestCode = UCDCSpuriousOutBand
	}
	result.SpuriousResults = true
	result.SpuriousResultValue = SpuriousResults{
		Frequency:        make([]string, 0),
		MeasuredPowerdBm: make([]string, 0),
		SpuriousLeveldBC: make([]string, 0),
	}
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetPower(udc.inputPower + udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	if udc.stop {
		udc.setError("User aborted")
	}

	if inBand {
		udc.setStatus("Measuring In-Band Spurious")
		response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.inBandSpectrum.span,
			udc.inBandSpectrum.rbw, udc.inBandSpectrum.vbw)
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
	} else {
		udc.setStatus("Measuring Out-of-Band Spurious")
		response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.outBandSpectrum.span,
			udc.outBandSpectrum.rbw, udc.outBandSpectrum.vbw)
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
	}

	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)

	noiseFloor := response.Result["MinValue"].Value
	powerOut := response.Result["MaxValue"].Value + udc.outputCableLoss[0]
	power_peaks := make([]float64, 0)
	freq_peaks := make([]float64, 0)
	deviation_peaks := make([]float64, 0)
	response = udc.sa.SetPeakThresholdAndExcursion(noiseFloor+10, 1)
	if !response.Success {
		udc.setError("Excursion cannot be set")
		return
	}
	response = udc.sa.SetMaxHold()
	if !response.Success {
		udc.setError("Max Hold cannot be set")
		return
	}
	udc.sa.WaitForSweeps(5)
	response = udc.sa.GetMaxMarkerValue()
	if !response.Success {
		udc.setError("SA Power Cannot be read")
		return
	}
	prevFreq := response.Result["MarkerX"].Value
	for {
		response = udc.sa.SetMarkerNextPeak(1)
		if !response.Success {
			udc.setError("Marker cannot be set")
			return
		}
		response = udc.sa.GetMarkerValue(1)
		if !response.Success {
			udc.setError("SA Power Cannot be read")
			return
		}
		spurious_val := response.Result["MarkerY"].Value + udc.outputCableLoss[0]
		spurious_freq := response.Result["MarkerX"].Value

		if spurious_freq != prevFreq {
			freq_peaks = append(freq_peaks, spurious_freq)
			power_peaks = append(power_peaks, spurious_val)
			deviation_peaks = append(deviation_peaks, (powerOut - spurious_val))
			prevFreq = spurious_freq

		} else {
			break
		}
	}
	for i := 0; i < len(power_peaks); i = i + 1 {
		result.SpuriousResultValue.Frequency = append(result.SpuriousResultValue.Frequency, strconv.FormatFloat(freq_peaks[i], 'f', 6, 64))
		result.SpuriousResultValue.MeasuredPowerdBm = append(result.SpuriousResultValue.MeasuredPowerdBm, strconv.FormatFloat(power_peaks[i], 'f', 6, 64))
		result.SpuriousResultValue.SpuriousLeveldBC = append(result.SpuriousResultValue.SpuriousLeveldBC, strconv.FormatFloat(deviation_peaks[i], 'f', 6, 64))
		udc.measurementMonitor <- result
	}

	if len(power_peaks) == 0 {
		result.SpuriousResultValue.Frequency = append(result.SpuriousResultValue.Frequency, "NIL")
		result.SpuriousResultValue.MeasuredPowerdBm = append(result.SpuriousResultValue.MeasuredPowerdBm, "NIL")
		result.SpuriousResultValue.SpuriousLeveldBC = append(result.SpuriousResultValue.SpuriousLeveldBC, "NIL")
		udc.measurementMonitor <- result
	}
	udc.setStatus("Completed Measurement for Spurious")
	udc.setStatus("Saving Results")
	if inBand {
		udc.saveResults(result, "Output Port - Spurious Measurement - In Band")
	} else {
		udc.saveResults(result, "Output Port - Spurious Measurement - Out of Band")
	}
	udc.setStatus("Spurious Measurement Completed")

	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) LOLeakageMeasurement() {
	udc.setStatus("LO Leakage Measurement Started")
	var result ConvertorResults
	result.TestName = "Output Port - LO Leakage Measurement"
	result.TestCode = UCDCLOLeakage
	result.PowerOrLeakageResults = true
	result.PowerOrLeakageResultValue = PowerOrLeakageResults{
		Frequency: 0.0,
		Power:     0.0,
	}

	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.converter.InputFrequency-udc.converter.OutputFrequency), udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sa.CheckIfCarrierIsPresent()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	time.Sleep(1000 * time.Millisecond)

	Loleakage := response.Result["MaxValue"].Value + udc.outputCableLoss[2]
	result.PowerOrLeakageResultValue.Frequency = math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
	result.PowerOrLeakageResultValue.Power = Loleakage
	udc.measurementMonitor <- result

	udc.setStatus("Compelted Measurement for LO Leakage")
	udc.setStatus("Saving Results")
	udc.saveResults(result, "Output Port - LO Leakage Measurement")
	udc.setStatus("LO Leakage Measurement Completed")
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) OutputInputLeakageMeasurement() {
	udc.setStatus("Input Leakage Measurement Started")
	var result ConvertorResults
	result.TestName = "Output Port - Input Leakage Measurement"
	result.TestCode = UCDCInputLeakage
	result.PowerOrLeakageResults = true
	result.PowerOrLeakageResultValue = PowerOrLeakageResults{
		Frequency: 0.0,
		Power:     0.0,
	}

	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(udc.converter.InputFrequency, udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(udc.inputPower + udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.CheckIfCarrierIsPresent()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	response = udc.sa.PeakSearch(true, 1)
	if !response.Success {
		udc.setError("Unable to operate SA in maxhold mode")
		return
	}
	inputLeakage := response.Result["MaxValue"].Value + udc.outputCableLoss[1]
	result.PowerOrLeakageResultValue.Frequency = udc.converter.InputFrequency
	result.PowerOrLeakageResultValue.Power = inputLeakage
	udc.measurementMonitor <- result
	udc.setStatus("Completed Measurement for InputLeakage")
	udc.setStatus("Saving Results")
	udc.saveResults(result, "Output Port - Input Leakage Measurement")
	udc.setStatus("InputLeakage Measurement Completed")
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) OutputExtLOGainMeasurement(stepSize float64, cable bool) {
	udc.setStatus("Convertor Gain Measurement with External LO Started")
	var maxPower, minPower float64
	var result ConvertorResults
	result.GainResults = true
	result.GainResultValue = GainResults{
		SetPower:    make([]float64, 0),
		OutputPower: make([]float64, 0),
		Gain:        make([]float64, 0),
		AverageGain: 0.0,
	}
	if cable {
		maxPower = udc.converter.MaxPowerCable
		minPower = udc.converter.MinPowerCable
		result.TestName = "Output Port - Gain Measurement - External LO - Cable"
		result.TestCode = UCDCGainExternalCable
	} else {
		maxPower = udc.converter.MaxPowerRadiated.Float64
		minPower = udc.converter.MinPowerRadiated.Float64
		result.TestName = "Output Port - Gain Measurement - External LO - Radiated"
		result.TestCode = UCDCGainExternalRadiated
	}

	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetFrequency(math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency))
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	loPower, err := udc.getExtLOPower()
	if err != nil {
		udc.setError("Ext LO Power Matching to be done before Gain Measurement with Ext LO")
		return
	}
	response = udc.sgExt.SetPower(loPower - udc.loCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
		udc.sgExt.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(minPower - udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	for powerSet := minPower + udc.inputCableLoss; powerSet <= maxPower+udc.inputCableLoss; powerSet = powerSet + stepSize {
		if udc.stop {
			udc.setError("Measurement Aborted by User")
			return
		}
		powerStr := fmt.Sprintf("%.3f", powerSet)
		response = udc.sg.SetPower(powerSet)
		if !response.Success {
			udc.setError("SG Power Cannot be set to " + powerStr)
			return
		}

		time.Sleep(1000 * time.Millisecond)
		referencePower := powerSet
		response = udc.sa.GetMaxMarkerValue()
		if !response.Success {
			udc.setError("SA Power Cannot be read")
			return
		}
		powerOut := response.Result["MarkerY"].Value + udc.outputCableLoss[0]
		if powerSet == minPower+udc.inputCableLoss || powerSet == maxPower+udc.inputCableLoss {
			response = udc.sa.GetSpectrumDump()
			if !response.Success {
				udc.setError("Unable to get spectrum dump")
				return
			}
		}
		Gain := powerOut - referencePower
		result.GainResultValue.SetPower = append(result.GainResultValue.SetPower, powerSet)
		result.GainResultValue.OutputPower = append(result.GainResultValue.OutputPower, powerOut)
		result.GainResultValue.Gain = append(result.GainResultValue.Gain, Gain)
		result.GainResultValue.AverageGain = mean(result.GainResultValue.Gain)
		udc.measurementMonitor <- result
		udc.setStatus("Completed Measurement for " + powerStr)
	}
	udc.setStatus("Saving Results")
	if cable {
		udc.saveResults(result, "Output Port - Gain Measurement - External LO - Cable")
	} else {
		udc.saveResults(result, "Output Port - Gain Measurement - External LO - Radiated")
	}
	udc.setStatus("Gain Measurement with Ext LO Completed")
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) MonitorPowerMeasurement(output bool) {
	if output {
		udc.setStatus("Output Monitor Power Measurement Started")
	} else {
		udc.setStatus("Input Monitor Power Measurement Started")
	}
	var result ConvertorResults
	result.PowerMatchingResults = true
	if output {
		result.TestName = "Output Monitor Port - Power Measurement"
		result.TestCode = UCDCOutputMonPower
	} else {
		result.TestName = "Input Monitor Port - Power Measurement"
		result.TestCode = UCDCInputMonPower
	}
	result.PowerOrLeakageResultValue = PowerOrLeakageResults{
		Frequency: 0.0,
		Power:     0.0,
	}
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.converter.InputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	if output {
		response = udc.sa.SetSpectrum(udc.converter.OutputFrequency, udc.powerSpectrum.span,
			udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
	} else {
		response = udc.sa.SetSpectrum(udc.converter.InputFrequency, udc.powerSpectrum.span,
			udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
	}

	response = udc.sg.SetPower(udc.inputPower + udc.inputCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = udc.sa.PeakSearch(true, 1)
	if !response.Success {
		udc.setError("Unable to operate SA in maxhold mode")
		return
	}
	udc.sa.WaitForSweeps(5)
	outputLeakage := response.Result["MarkerY"].Value + udc.outputCableLoss[0]
	if output {
		result.PowerOrLeakageResultValue.Frequency = udc.converter.OutputFrequency
	} else {
		result.PowerOrLeakageResultValue.Frequency = udc.converter.InputFrequency
	}
	result.PowerOrLeakageResultValue.Power = outputLeakage
	udc.measurementMonitor <- result
	udc.setStatus("Completed Measurement for Monitor Power")
	udc.setStatus("Saving Results")
	if output {
		udc.saveResults(result, "Output Monitor Port - Power Measurement")
		udc.setStatus("Output Monitor Power Measurement Completed")
	} else {
		udc.saveResults(result, "Input Monitor Port - Power Measurement")
		udc.setStatus("Input Monitor Power Measurement Completed")
	}
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) LOMonFreqPowerMeasurement() {
	udc.setStatus("LO MON Port Frequency & Power Measurement Started")
	var result ConvertorResults
	result.TestName = "LO MON Port Frequency & Power Measurement"
	result.TestCode = UCDCLOMonPower
	result.PowerOrLeakageResults = true
	result.FrequencyResults = true
	result.FrequencyResultValue = FrequencyResults{
		ExpectedFrequency: 0.0,
		MeasuredFrequency: 0.0,
		Deviation:         0.0,
	}
	result.PowerOrLeakageResultValue = PowerOrLeakageResults{
		Frequency: 0.0,
		Power:     0.0,
	}
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()
	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.converter.InputFrequency-udc.converter.OutputFrequency), udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = udc.sa.PeakSearch(true, 1)
	if !response.Success {
		udc.setError("Unable to operate SA in maxhold mode")
		return
	}
	udc.sa.WaitForSweeps(5)
	LOMonPower := response.Result["MarkerY"].Value
	response = udc.sa.GetFrequencyInCounterMode(1)
	if !response.Success {
		udc.setError("Unable to get frequency in counter mode")
		return
	}
	udc.sa.WaitForSweeps(5)
	frequency := response.Result["Frequency"].Value
	response = udc.sa.GetSpectrumDump()
	if !response.Success {
		udc.setError("Unable to get spectrum dump")
		return
	}
	freqDeviation := math.Abs(frequency - math.Abs(udc.converter.InputFrequency-udc.converter.OutputFrequency))
	result.FrequencyResultValue.ExpectedFrequency = math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency)
	result.FrequencyResultValue.MeasuredFrequency = frequency
	result.FrequencyResultValue.Deviation = freqDeviation
	result.PowerOrLeakageResultValue.Frequency = frequency
	result.PowerOrLeakageResultValue.Power = LOMonPower
	udc.measurementMonitor <- result
	udc.setStatus("Completed Measurement for LO Mon Power and Frequency")
	udc.setStatus("Saving Results")
	udc.saveResults(result, "LO MON Port - Frequency & Power Measurement")
	udc.setStatus("LO Mon Power & Frequency Measurement Completed")
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) LOMonPhaseNoiseMeasurement() {
	udc.setStatus("LO Mon Port Phase Noise Measurement Started")
	var result ConvertorResults
	result.TestName = "LO Mon Port Phase Noise Measurement"
	result.TestCode = UCDCLOMonPhaseNoise
	result.PhaseNoiseResults = true
	result.PhaseNoiseResultValue = PhaseNoiseResults{
		Frequency:  make([]float64, 0),
		PhaseNoise: make([]float64, 0),
	}

	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
	}()

	response = udc.sg.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.converter.InputFrequency-udc.converter.OutputFrequency), udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	response = udc.sa.SetPhaseNoiseMeasurement()
	if !response.Success {
		udc.setError("SA cannot be set in phase noise mode")
		return
	}

	phaseNoise := make([]float64, 0)

	j := 0
	for i := 1000; i < 1000000; i = i * 10 {
		response = udc.sa.SetMarkerValuePhaseNoise(float64(i), 1)
		if !response.Success {
			udc.setError("Marker cannot be set in phase noise mode")
			return
		}
		response = udc.sa.GetPhaseNoiseMarkerY(1)
		if !response.Success {
			udc.setError("Marker cannot be set in phase noise mode")
			return
		}
		phase := response.Result["MarkerY"].Value
		if j == 1 {
			response = udc.sa.GetSpectrumDump()
			if !response.Success {
				udc.setError("Unable to get spectrum dump")
				return
			}
		}
		phaseNoise = append(phaseNoise, phase)
		result.PhaseNoiseResultValue.Frequency = append(result.PhaseNoiseResultValue.Frequency, float64(i))
		result.PhaseNoiseResultValue.PhaseNoise = append(result.PhaseNoiseResultValue.PhaseNoise, phase)
		udc.measurementMonitor <- result
		j = j + 1
	}
	time.Sleep(1000 * time.Millisecond)
	udc.setStatus("Completed Measurement for LO Mon Phase Noise")
	udc.setStatus("Saving Results")
	udc.saveResults(result, "LO MON Port - Phase Noise Measurement")
	udc.setStatus("LO Mon Phase Noise Measurement Completed")
	close(udc.statusMonitor)
	close(udc.measurementMonitor)
}

func (udc *UpDownConverterMeasurement) ExtLOPowerMatch() {
	udc.setStatus("External LO Power Matching started")
	var result ConvertorResults
	result.TestName = "Output Port - Ext LO Power Matching"
	result.TestCode = UCDCExtLOPowerMatch
	result.PowerMatchingResults = true
	result.PowerMatchingResultValue = PowerMatchingResults{
		InternalLOPowerMeasured: 0.0,
		ExternalLOPowerMeasured: 0.0,
		ExternalSGPowerSet:      0.0,
	}
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sgExt.SetFrequency(math.Abs(udc.converter.InputFrequency - udc.converter.OutputFrequency))
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	loPower, err := udc.getLOPower()
	if err != nil {
		udc.setError("LO Power Measurement has to be completed before Ext LO Power Matching")
		return
	}
	response = udc.sgExt.SetPower(loPower + udc.loCableLoss)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetModOff()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	defer func() {
		udc.sa.SetAlignmentOn()
		udc.sa.SystemPreset()
		udc.sg.SetRFOff()
		udc.sgExt.SetRFOff()
	}()

	response = udc.sa.SetSpectrum(math.Abs(udc.converter.InputFrequency-udc.converter.OutputFrequency), udc.powerSpectrum.span,
		udc.powerSpectrum.rbw, udc.powerSpectrum.vbw)
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = udc.sa.GetMaxMarkerValue()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	ExtLoPower := response.Result["MarkerY"].Value
	powerDev := loPower - ExtLoPower - udc.loCableLoss
	powerSet := 0.0
	for math.Abs(powerDev) <= 0.1 {
		response = udc.sgExt.SetPower(loPower + powerDev + udc.loCableLoss)
		powerSet = loPower + powerDev + udc.loCableLoss
		if !response.Success {
			udc.setError("Unable to communicate with SG")
			return
		}
		response = udc.sa.GetMaxMarkerValue()
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
		i := response.Result["MarkerY"].Value
		powerDev = loPower - i - udc.loCableLoss
		ExtLoPower = i + udc.loCableLoss
	}
	result.PowerMatchingResultValue.InternalLOPowerMeasured = loPower
	result.PowerMatchingResultValue.ExternalLOPowerMeasured = ExtLoPower
	result.PowerMatchingResultValue.ExternalSGPowerSet = powerSet
	udc.measurementMonitor <- result
	udc.setStatus("Saving Results")
	udc.saveResults(result, "External LO Power Matching")
	udc.setStatus("External LO Power Matching Completed")
	close(udc.measurementMonitor)
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) getLOPower() (float64, error) {
	result, err := resultsDB.GetUpDownConverterResult(udc.converter.Name, "LO MON Port - Frequency & Power Measurement")
	if err != nil {
		return 0.0, err
	}
	var res ConvertorResults
	err = json.Unmarshal([]byte(result.Results), &res)
	if err != nil {
		return 0.0, err
	}
	return res.PowerOrLeakageResultValue.Power, nil
}

func (udc *UpDownConverterMeasurement) getExtLOPower() (float64, error) {
	result, err := resultsDB.GetUpDownConverterResult(udc.converter.Name, "External LO Power Matching")
	if err != nil {
		return 0.0, err
	}
	var res ConvertorResults
	err = json.Unmarshal([]byte(result.Results), &res)
	if err != nil {
		return 0.0, err
	}
	return res.PowerMatchingResultValue.ExternalSGPowerSet, nil
}
