package tne

import (
	"fmt"
	"math"
	"prismServer/database"
	"prismServer/driver"
	"strconv"
	"time"
)

type UpDownConverterMeasurement struct {
	deviceProfile   string
	externalSGName  string
	converterName   string
	spectrumProfile string
	maxPower        float64
	minPower        float64
	stepSize        float64
	inputFrequency  float64
	outputFrequency float64
	inputPower      float64
	inputCableLoss  float64
	outputCableLoss []float64
	LOCableLoss     float64
	LOPower         float64
	harmonics       []float64
	sa              driver.SA
	sg              driver.SG
	sgExt           driver.SG
	currentStatus   [][]string
	statusMonitor   chan MeasurementStatus
	stop            bool
}

func (udc *UpDownConverterMeasurement) GetStatusMonitor() chan MeasurementStatus {
	return udc.statusMonitor
}
func (udc *UpDownConverterMeasurement) Initialize(deviceProfile string, externalSGName string, converterName string, spectrumProfile string,
	maxPower float64, minPower float64, stepSize float64, inputFrequency float64, outputFrequency float64, inputPower float64,
	inputCableLoss float64, outputCableLoss []float64, LOCableLoss float64, LOPower float64, harmonics []float64, sgExt string) {
	udc.deviceProfile = deviceProfile
	udc.externalSGName = externalSGName
	udc.converterName = converterName
	udc.spectrumProfile = spectrumProfile
	udc.maxPower = maxPower
	udc.minPower = minPower
	udc.stepSize = stepSize
	udc.inputFrequency = inputFrequency
	udc.outputFrequency = outputFrequency
	udc.inputPower = inputPower
	udc.inputCableLoss = inputCableLoss
	udc.outputCableLoss = outputCableLoss
	udc.LOCableLoss = LOCableLoss
	udc.LOPower = LOPower
	udc.harmonics = harmonics
	udc.currentStatus = make([][]string, 0)
	udc.statusMonitor = make(chan MeasurementStatus, 20)
	udc.loadDevices()
	header := make([]string, 0)
	udc.currentStatus = append(udc.currentStatus, header)
	udc.stop = false
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
	var measure = MeasurementStatus{
		Completed:     true,
		Success:       false,
		Message:       message,
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) GainMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "UC/DC Gain Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.inputFrequency)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(udc.outputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(udc.minPower - udc.inputCableLoss)
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
	slNo := 1
	slNoStr := strconv.Itoa(slNo)
	row := make([]string, 0)
	for powerSet := udc.minPower + udc.inputCableLoss; powerSet <= udc.maxPower+udc.inputCableLoss; powerSet = powerSet + udc.stepSize {
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
		if powerSet == udc.minPower+udc.inputCableLoss || powerSet == udc.maxPower+udc.inputCableLoss {
			response = udc.sa.GetSpectrumDump()
			if !response.Success {
				udc.setError("Unable to get spectrum dump")
				return
			}
		}
		PowerOutStr := fmt.Sprintf("%.3f", powerOut)
		Gain := powerOut - referencePower
		GainStr := fmt.Sprintf("%.3f", Gain)
		row = append(row, slNoStr, powerStr, PowerOutStr, GainStr)
		slNo = slNo + 1
		udc.currentStatus = append(udc.currentStatus, row)
		measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Completed Measurement for " + powerStr,
			CurrentStatus: udc.currentStatus,
		}
		measure.CurrentStatus = append(measure.CurrentStatus, row)
		udc.statusMonitor <- measure
	}
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Gain Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) SpuriousMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "UC/DC Spurious Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.inputFrequency)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(udc.outputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	powerSet := udc.inputPower + udc.inputCableLoss
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
	response = udc.sg.SetRFOn()
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	time.Sleep(1000 * time.Millisecond)
	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	noiseFloor := response.Result["MinValue"].Value
	time.Sleep(1000 * time.Millisecond)
	response = udc.sa.GetMaxMarkerValue()
	if !response.Success {
		udc.setError("SA Power Cannot be read")
		return
	}
	powerOut := response.Result["MarkerY"].Value + udc.outputCableLoss[0]
	power_peaks := make([]float64, 0)
	freq_peaks := make([]float64, 0)
	deviation_peaks := make([]float64, 0)
	response = udc.sa.SetPeakThresholdAndExcursion(noiseFloor+10, 1)
	if !response.Success {
		udc.setError("excursion cannot be set")
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
	row := make([]string, 0)
	for i := 0; i < len(power_peaks); i = i + 1 {
		row = append(row, strconv.Itoa(i+1), strconv.FormatFloat(freq_peaks[i], 'f', 6, 64), strconv.FormatFloat(power_peaks[i], 'f', 6, 64), strconv.FormatFloat(deviation_peaks[i], 'f', 6, 64))
		udc.currentStatus = append(udc.currentStatus, row)
	}
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for Spurious",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure

	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " Spurious Measurement test Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)

}

func (udc *UpDownConverterMeasurement) FrequencyMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Frequency Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.inputFrequency)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(udc.outputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	freq_deviation := math.Abs(frequency - udc.outputFrequency)
	row := make([]string, 0)
	row = append(row, strconv.Itoa(1), strconv.FormatFloat(udc.outputFrequency, 'f', 6, 64), strconv.FormatFloat(frequency, 'f', 6, 64), strconv.FormatFloat(freq_deviation, 'f', 6, 64))
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for frequency",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Frequency Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) LoMonFreqPowerMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "LO MON Port Frequency & Power Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.inputFrequency-udc.outputFrequency), spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	frequency := response.Result["MarkerX"].Value
	response = udc.sa.GetSpectrumDump()
	if !response.Success {
		udc.setError("Unable to get spectrum dump")
		return
	}
	udc.LOPower = LOMonPower
	freq_deviation := math.Abs(frequency - math.Abs(udc.inputFrequency-udc.outputFrequency))
	row := make([]string, 0)
	row = append(row, strconv.Itoa(1), strconv.FormatFloat(math.Abs(udc.inputFrequency-udc.outputFrequency), 'f', 6, 64), strconv.FormatFloat(frequency, 'f', 6, 64), strconv.FormatFloat(freq_deviation, 'f', 6, 64), strconv.FormatFloat(LOMonPower, 'f', 6, 64))
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for LO MON Power & frequency",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " LO MON Power & Frequency Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) PhaseNoiseMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "PhaseNoise Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.inputFrequency-udc.outputFrequency), spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	row := make([]string, 0)
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
		row = append(row, strconv.Itoa(j+1), strconv.FormatFloat(phaseNoise[j], 'f', 6, 64), strconv.Itoa(i))
		j = j + 1
	}
	time.Sleep(1000 * time.Millisecond)
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for phaseNoise",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " phaseNoise Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) LoLeakageMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "LoLeakage Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.outputFrequency-udc.inputFrequency), spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	Loleakage := response.Result["MarkerY"].Value + udc.outputCableLoss[2]
	row := make([]string, 0)
	row = append(row, strconv.Itoa(1), strconv.FormatFloat(Loleakage, 'f', 6, 64))
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for Lo Leakage",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " LO leakage Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) InputLeakageMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "InputLeakage Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.inputFrequency)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(udc.inputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	inputLeakage := response.Result["MarkerY"].Value + udc.outputCableLoss[1]
	row := make([]string, 0)
	row = append(row, strconv.Itoa(1), strconv.FormatFloat(inputLeakage, 'f', 6, 64))
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for InputLeakage",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " InputLeakage Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) OutputLeakageMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "OutputLeakage Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.inputFrequency)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(udc.outputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	outputLeakage := response.Result["MarkerY"].Value + udc.outputCableLoss[0]
	row := make([]string, 0)
	row = append(row, strconv.Itoa(1), strconv.FormatFloat(outputLeakage, 'f', 6, 64))
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for outputLeakage",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " outputLeakage Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) HarmonicsMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Harmonics Measurement Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}
	response = udc.sa.SetReferenceNominal()
	if !response.Success {
		udc.setError("Carrier Not found")
		return
	}
	response = udc.sa.SetSpectrum(udc.outputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sa.PeakSearch(true, 1)
	if !response.Success {
		udc.setError("Unable to operate SA in maxhold mode")
		return
	}
	udc.sa.WaitForSweeps(5)
	outputpower := response.Result["MarkerY"].Value
	HarmonicPower := make([]float64, 0)
	row := make([]string, 0)
	for i := 0; i < 3; i = i + 1 {
		response = udc.sa.SetSpectrum(udc.harmonics[i], spectrum.Span,
			float64(spectrum.RBW), float64(spectrum.VBW))
		if !response.Success {
			udc.setError("Unable to communicate with SA")
			return
		}
		response = udc.sa.CheckIfCarrierIsPresent()
		if !response.Success {
			udc.setError("carrier present fucntion error")
			return
		}
		if !response.Result["carrier"].Bool {
			HarmonicPower = append(HarmonicPower, 0.0)
			row = append(row, strconv.Itoa(i+1), strconv.FormatFloat(udc.harmonics[i], 'f', 6, 64), "NIL")
		} else {

			response = udc.sa.PeakSearch(true, 1)
			if !response.Success {
				udc.setError("Unable to operate SA in maxhold mode")
				return
			}
			udc.sa.WaitForSweeps(5)
			harmpower := response.Result["MarkerY"].Value
			harmfreq := response.Result["MarkerX"].Value
			lower := udc.harmonics[i] - (udc.harmonics[i] * 2 * 1e-6)
			higher := udc.harmonics[i] + (udc.harmonics[i] * 2 * 1e-6)
			if lower < harmfreq && harmfreq < higher {
				HarmonicPower = append(HarmonicPower, (outputpower - harmpower))
				row = append(row, strconv.Itoa(i+1), strconv.FormatFloat(udc.harmonics[i], 'f', 6, 64), strconv.FormatFloat(HarmonicPower[i], 'f', 6, 64))
			} else {
				HarmonicPower = append(HarmonicPower, 0.0)
				row = append(row, strconv.Itoa(i+1), strconv.FormatFloat(udc.harmonics[i], 'f', 6, 64), "NIL")
			}

		}
	}
	time.Sleep(1000 * time.Millisecond)
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Completed Measurement for Harmonics",
		CurrentStatus: udc.currentStatus,
	}
	measure.CurrentStatus = append(measure.CurrentStatus, row)
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " Harmonics Measurement Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) ExtLOGainMeasurement() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "UC/DC Gain Measurement with ext LO Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sg.SetFrequency(udc.inputFrequency)
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetFrequency(math.Abs(udc.inputFrequency - udc.outputFrequency))
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetPower(udc.LOPower - udc.LOCableLoss)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(udc.outputFrequency, spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}

	response = udc.sg.SetPower(udc.minPower - udc.inputCableLoss)
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
	slNo := 1
	row := make([]string, 0)
	for powerSet := udc.minPower + udc.inputCableLoss; powerSet <= udc.maxPower+udc.inputCableLoss; powerSet = powerSet + udc.stepSize {
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
		if powerSet == udc.minPower+udc.inputCableLoss || powerSet == udc.maxPower+udc.inputCableLoss {
			response = udc.sa.GetSpectrumDump()
			if !response.Success {
				udc.setError("Unable to get spectrum dump")
				return
			}
		}
		PowerOutStr := fmt.Sprintf("%.3f", powerOut)
		slNoStr := strconv.Itoa(slNo)
		slNo = slNo + 1
		Gain := powerOut - referencePower
		GainStr := fmt.Sprintf("%.3f", Gain)
		row = append(row, slNoStr, powerStr, PowerOutStr, GainStr)
		udc.currentStatus = append(udc.currentStatus, row)
		measure = MeasurementStatus{
			Completed:     false,
			Success:       true,
			Message:       "Completed Measurement for " + powerStr,
			CurrentStatus: udc.currentStatus,
		}
		measure.CurrentStatus = append(measure.CurrentStatus, row)
		udc.statusMonitor <- measure
	}
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       "Gain Measurement with Ext LO Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}

func (udc *UpDownConverterMeasurement) ExtLOPowerMatch() {
	var measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "External LO power Matching Started",
		CurrentStatus: make([][]string, 0),
	}
	udc.statusMonitor <- measure
	response := udc.sa.SetAlignmentOff()
	if !response.Success {
		udc.setError("Unable to communicate with SA")
		return
	}
	response = udc.sgExt.SetFrequency(math.Abs(udc.inputFrequency - udc.outputFrequency))
	if !response.Success {
		udc.setError("Unable to communicate with SG")
		return
	}
	response = udc.sgExt.SetPower(udc.LOPower + udc.LOCableLoss)
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

	spectrum, ok := database.GetSpectrumProfile(udc.spectrumProfile)
	if !ok {
		udc.setError("Unable to Read Spectrum from Database")
		return
	}

	response = udc.sa.SetSpectrum(math.Abs(udc.inputFrequency-udc.outputFrequency), spectrum.Span,
		float64(spectrum.RBW), float64(spectrum.VBW))
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
	powerDev := udc.LOPower - ExtLoPower - udc.LOCableLoss
	for math.Abs(powerDev) <= 0.1 {
		response = udc.sgExt.SetPower(udc.LOPower + powerDev + udc.LOCableLoss)
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
		powerDev = udc.LOPower - i - udc.LOCableLoss
		ExtLoPower = i + udc.LOCableLoss
	}
	row := make([]string, 0)
	row = append(row, strconv.Itoa(1), strconv.FormatFloat(ExtLoPower, 'f', 6, 64))
	udc.currentStatus = append(udc.currentStatus, row)
	measure = MeasurementStatus{
		Completed:     false,
		Success:       true,
		Message:       "Saving Results",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	measure = MeasurementStatus{
		Completed:     true,
		Success:       true,
		Message:       " Ext LO power Matching Completed",
		CurrentStatus: udc.currentStatus,
	}
	udc.statusMonitor <- measure
	close(udc.statusMonitor)
}
