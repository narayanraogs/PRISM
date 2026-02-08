package tne

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/reports"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type spectrumSettings struct {
	span float64
	rbw  float64
	vbw  float64
}
type GroundTransmitterMeasurement struct {
	deviceProfile           string
	gtxName                 string
	component               string
	modScheme               string
	intermediateFrequency   float64
	outputCableLoss         float64
	subCarrierFreq          float64
	freqdeviation           float64
	modIndex                float64
	noOfHarmonics           int
	powerSpectrun           spectrumSettings
	frequencySpectrum       spectrumSettings
	inBandSpuriousSpectrum  spectrumSettings
	outBandSpuriousSpectrum spectrumSettings
	report                  reports.Report
	order                   []string
	images                  []reports.Images
	success                 bool
	result                  GTxResult

	sa            driver.SA
	gtx           driver.GTX
	currentStatus [][]string
	statusMonitor chan RTStatus
	resultMonitor chan GTxResult
	stop          bool
}

type GTxResult struct {
	PowerSpec                              float64
	PowerMeasured                          float64
	PowerDeviation                         float64
	PowerMeasurementCompleted              bool
	FreqSpecMHz                            float64
	FreqMeasuredMHz                        float64
	FreqDeviationkHz                       float64
	FreqMeasurementCompleted               bool
	InBandSpuriousFreqOffsetskHz           []float64
	InBandPowerOffsets                     []float64
	InBandSpuriousMeasurementCompleted     bool
	OutBandSpuriousFreqOffsetskHz          []float64
	OutBandPowerOffsets                    []float64
	OutBandSpuriousMeasurementCompleted    bool
	HarmonicsFreqMHz                       []float64
	HarmonicsMeasureddBm                   []float64
	HarmonicsPresent                       []bool
	HarmonicsNoiseFloor                    []float64
	HarmonicsMeasurementCompleted          bool
	ModIndexApplicable                     bool
	ModIndexSet                            float64
	ModIndexMeasured                       float64
	ModIndexDeviation                      float64
	ModIndexMeasurementCompleted           bool
	FrequencyDeviationApplicable           bool
	FrequencyDeviationSet                  float64
	FrequencyDeviationMeasured             float64
	FrequencyDeviationDeviation            float64
	FrequencyDeviationMeasurementCompleted bool
	PhaseNoiseAt1Khz                       float64
	PhaseNoiseAt10Khz                      float64
	PhaseNoiseAt100Khz                     float64
	PhaseNoiseAt1Mhz                       float64
	PhaseNoiseMeasurementCompleted         bool
}

func (res *GTxResult) Copy(gtxResult GTxResult) {
	res.PowerSpec = gtxResult.PowerSpec
	res.PowerMeasured = gtxResult.PowerMeasured
	res.PowerDeviation = gtxResult.PowerDeviation
	res.PowerMeasurementCompleted = gtxResult.PowerMeasurementCompleted
	res.FreqSpecMHz = gtxResult.FreqSpecMHz
	res.FreqMeasuredMHz = gtxResult.FreqMeasuredMHz
	res.FreqDeviationkHz = gtxResult.FreqDeviationkHz
	res.FreqMeasurementCompleted = gtxResult.FreqMeasurementCompleted
	res.InBandSpuriousFreqOffsetskHz = gtxResult.InBandSpuriousFreqOffsetskHz
	res.InBandPowerOffsets = gtxResult.InBandPowerOffsets
	res.InBandSpuriousMeasurementCompleted = gtxResult.InBandSpuriousMeasurementCompleted
	res.OutBandSpuriousFreqOffsetskHz = gtxResult.OutBandSpuriousFreqOffsetskHz
	res.OutBandPowerOffsets = gtxResult.OutBandPowerOffsets
	res.OutBandSpuriousMeasurementCompleted = gtxResult.OutBandSpuriousMeasurementCompleted
	res.HarmonicsFreqMHz = gtxResult.HarmonicsFreqMHz
	res.HarmonicsMeasureddBm = gtxResult.HarmonicsMeasureddBm
	res.HarmonicsPresent = gtxResult.HarmonicsPresent
	res.HarmonicsNoiseFloor = gtxResult.HarmonicsNoiseFloor
	res.HarmonicsMeasurementCompleted = gtxResult.HarmonicsMeasurementCompleted
	res.ModIndexApplicable = gtxResult.ModIndexApplicable
	res.ModIndexSet = gtxResult.ModIndexSet
	res.ModIndexMeasured = gtxResult.ModIndexMeasured
	res.ModIndexDeviation = gtxResult.ModIndexDeviation
	res.ModIndexMeasurementCompleted = gtxResult.ModIndexMeasurementCompleted
	res.FrequencyDeviationApplicable = gtxResult.FrequencyDeviationApplicable
	res.FrequencyDeviationSet = gtxResult.FrequencyDeviationSet
	res.FrequencyDeviationMeasured = gtxResult.FrequencyDeviationMeasured
	res.FrequencyDeviationDeviation = gtxResult.FrequencyDeviationDeviation
	res.FrequencyDeviationMeasurementCompleted = gtxResult.FrequencyDeviationMeasurementCompleted
	res.PhaseNoiseAt1Khz = gtxResult.PhaseNoiseAt1Khz
	res.PhaseNoiseAt10Khz = gtxResult.PhaseNoiseAt10Khz
	res.PhaseNoiseAt100Khz = gtxResult.PhaseNoiseAt100Khz
	res.PhaseNoiseAt1Mhz = gtxResult.PhaseNoiseAt1Mhz
	res.PhaseNoiseMeasurementCompleted = gtxResult.PhaseNoiseMeasurementCompleted
}

func NewGTxGroundTransmitterMeasurement() *GroundTransmitterMeasurement {
	return &GroundTransmitterMeasurement{}
}

func (gtm *GroundTransmitterMeasurement) GetStatusMonitor() (chan RTStatus, chan GTxResult) {
	return gtm.statusMonitor, gtm.resultMonitor
}

func (gtm *GroundTransmitterMeasurement) SetDevices(deviceProfile string, component string, intermediateFrequency float64, outputCableLoss float64) bool {
	gtm.images = make([]reports.Images, 0)
	gtm.statusMonitor = make(chan RTStatus, 10)
	gtm.resultMonitor = make(chan GTxResult, 10)
	gtm.noOfHarmonics = 3
	gtm.deviceProfile = deviceProfile
	gtxName, ok := database.GetGTxFromDeviceProfile(gtm.deviceProfile)
	if !ok {
		gtm.setError("Unable to Load GTx from Database")
		return false
	}
	gtm.gtxName = gtxName
	saName, ok := database.GetSAFromDeviceProfile(gtm.deviceProfile)
	if !ok {
		gtm.setError("Unable to Load SA from Database")
		return false
	}
	ok = gtm.gtx.LoadDevice(gtxName)
	if !ok {
		gtm.setError("Unable to Load GTx from Database")
		return false
	}
	ok = gtm.sa.LoadDevice(saName)
	if !ok {
		gtm.setError("Unable to Load SA from Database")
		return false
	}
	gtm.component = component
	gtm.intermediateFrequency = intermediateFrequency
	if outputCableLoss < 0 {
		outputCableLoss = outputCableLoss * -1
	}
	gtm.outputCableLoss = outputCableLoss
	testPhase, _ := database.GetSelectedTestPhase()
	gtm.report.SetHeader("", "GTX Measurement", "", testPhase)
	gtm.currentStatus = make([][]string, 0)
	gtm.report.Results = make(map[string]reports.Result)
	gtm.report.Order = make([]string, 0)
	gtm.order = make([]string, 0)
	return true
}

func (gtm *GroundTransmitterMeasurement) SetModulationParameters(modScheme string, subCarrierFrequency float64, frequencyDeviation float64, modIndex float64) {
	gtm.modScheme = modScheme
	gtm.subCarrierFreq = subCarrierFrequency
	gtm.freqdeviation = frequencyDeviation
	gtm.modIndex = modIndex
	switch modScheme {
	case "FM":
		gtm.result.FrequencyDeviationApplicable = true
		gtm.result.ModIndexApplicable = false
	case "PM":
		gtm.result.FrequencyDeviationApplicable = false
		gtm.result.ModIndexApplicable = true
	default:
		gtm.result.FrequencyDeviationApplicable = false
		gtm.result.ModIndexApplicable = false
	}
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result
}

func (gtm *GroundTransmitterMeasurement) SetPowerSpectrum(span float64, rbw float64, vbw float64) {
	gtm.powerSpectrun.span = span
	gtm.powerSpectrun.rbw = rbw
	gtm.powerSpectrun.vbw = vbw
}

func (gtm *GroundTransmitterMeasurement) SetFrequencySpectrum(span float64, rbw float64, vbw float64) {
	gtm.frequencySpectrum.span = span
	gtm.frequencySpectrum.rbw = rbw
	gtm.frequencySpectrum.vbw = vbw
}

func (gtm *GroundTransmitterMeasurement) SetInBandSpectrum(span float64, rbw float64, vbw float64) {
	gtm.inBandSpuriousSpectrum.span = span
	gtm.inBandSpuriousSpectrum.rbw = rbw
	gtm.inBandSpuriousSpectrum.vbw = vbw
}

func (gtm *GroundTransmitterMeasurement) SetOutBandSpectrum(span float64, rbw float64, vbw float64) {
	gtm.outBandSpuriousSpectrum.span = span
	gtm.outBandSpuriousSpectrum.rbw = rbw
	gtm.outBandSpuriousSpectrum.vbw = vbw
}

func (gtm *GroundTransmitterMeasurement) Stop() {
	gtm.stop = true
}

func (gtm *GroundTransmitterMeasurement) setError(message string) {
	var measure = RTStatus{
		Completed: true,
		Success:   false,
		Error:     true,
		Message:   message,
	}
	gtm.statusMonitor <- measure
	close(gtm.resultMonitor)
	close(gtm.statusMonitor)
}

func (gtm *GroundTransmitterMeasurement) startMeasurement() error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Measurement Started",
	}
	gtm.statusMonitor <- measure
	response := gtm.sa.SetAlignmentOff()
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}
	response = gtm.sa.SystemPreset()
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}
	response = gtm.gtx.SetFrequency(gtm.component, gtm.intermediateFrequency)
	if !response.Success {
		gtm.setError("Unable to communicate with GTx")
		return fmt.Errorf("unable to communicate with GTx")
	}
	measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "GTx Frequency Set",
	}
	gtm.statusMonitor <- measure
	response = gtm.gtx.SetPower(gtm.component, 0)
	if !response.Success {
		gtm.setError("Unable to communicate with GTx")
		return fmt.Errorf("unable to communicate with GTx")
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (gtm *GroundTransmitterMeasurement) stopMeasurement() {
	gtm.sa.SetAlignmentOn()
	gtm.sa.SystemPreset()
	gtm.gtx.SetCarrierOff(gtm.component)
}

func (gtm *GroundTransmitterMeasurement) powerMeasurement() error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Power Measurement Started",
	}
	gtm.statusMonitor <- measure
	response := gtm.gtx.SetModulationOff(gtm.component)
	if !response.Success {
		gtm.setError("Unable to communicate with GTx")
		return fmt.Errorf("unable to communicate with GTx")
	}
	response = gtm.gtx.SetCarrierOn(gtm.component)
	if !response.Success {
		gtm.setError("Unable to communicate with GTx")
		return fmt.Errorf("unable to communicate with GTx")
	}
	time.Sleep(200 * time.Millisecond)
	response = gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrun.span,
		gtm.powerSpectrun.rbw, gtm.powerSpectrun.vbw)
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}
	response = gtm.sa.WaitForSweeps(5)
	if !response.Success {
		gtm.setError("Unable to wait for sweeps")
		return fmt.Errorf("unable to wait for sweeps")
	}
	response = gtm.sa.SetReferenceNominal()
	if !response.Success {
		gtm.setError("Carrier Not found")
		return fmt.Errorf("unable to communicate with SA")
	}
	resp := gtm.sa.WaitForSweeps(2)
	if !resp.Success {
		gtm.setError("Unable to wait for sweeps")
		return fmt.Errorf("unable to wait for sweeps")
	}
	power := response.Result["ReferenceLevel"].Value - 10 + gtm.outputCableLoss
	header := []string{"Specification [dBm]", "Measured [dBm]", "Deviation [dB]"}
	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell("0"))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", power)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", -power)))
	rows := make([][]reports.DataCell, 0)
	rows = append(rows, row)
	gtm.report.SetResults("Power", header, rows)
	gtm.order = append(gtm.order, "Power")

	response = gtm.sa.GetSpectrumDump()
	if !response.Success {
		gtm.setError("cannot take spectrum dump")
		return fmt.Errorf("cannot take spectrum dump")
	}
	image := reports.Images{
		FileData: response.Result["SpectrumDump"].String,
		Caption:  "Power Measurement",
	}
	gtm.images = append(gtm.images, image)

	gtm.currentStatus = append(gtm.currentStatus, []string{"Power", "Specification", "Measured", "Deviation"})
	gtm.currentStatus = append(gtm.currentStatus, []string{"", "0", fmt.Sprintf("%.2f", power), fmt.Sprintf("%.2f", -power)})
	gtm.result.PowerMeasurementCompleted = true
	gtm.result.PowerSpec = 0
	gtm.result.PowerMeasured = power
	gtm.result.PowerDeviation = -power
	var status = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Power Measurement Completed",
	}
	gtm.statusMonitor <- status
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result
	return nil
}

func (gtm *GroundTransmitterMeasurement) frequencyMeasurement() error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Frequency Measurement Started",
	}
	gtm.statusMonitor <- measure
	response := gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.frequencySpectrum.span,
		gtm.frequencySpectrum.rbw, gtm.frequencySpectrum.vbw)
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}
	gtm.sa.WaitForSweeps(5)
	if !response.Success {
		gtm.setError("Unable to wait for sweeps")
		return fmt.Errorf("unable to wait for sweeps")
	}
	response = gtm.sa.PeakSearch(true, 1)
	if !response.Success {
		gtm.setError("Unable to operate SA in maxhold mode")
		return fmt.Errorf("unable to operate SA in maxhold mode")
	}
	response = gtm.sa.GetFrequencyInCounterMode(1)
	if !response.Success {
		gtm.setError("Unable to get frequency in counter mode")
		return fmt.Errorf("unable to get frequency in counter mode")
	}
	frequency := response.Result["Frequency"].Value - 10
	deviation := gtm.intermediateFrequency - frequency
	header := []string{"Specification [MHz]", "Measured [MHz]", "Deviation [kHz]", "Deviation PPM"}
	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.6f", gtm.intermediateFrequency/1e6)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.6f", frequency/1e6)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.3f", deviation/1e3)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", deviation/2*1e-6)))
	rows := make([][]reports.DataCell, 0)
	rows = append(rows, row)
	gtm.report.SetResults("Frequency", header, rows)
	gtm.order = append(gtm.order, "Frequency")

	response = gtm.sa.GetSpectrumDump()
	if !response.Success {
		gtm.setError("cannot take spectrum dump")
		return fmt.Errorf("cannot take spectrum dump")
	}
	image := reports.Images{
		FileData: response.Result["SpectrumDump"].String,
		Caption:  "Frequency Measurement",
	}
	gtm.images = append(gtm.images, image)

	gtm.currentStatus = append(gtm.currentStatus, []string{"Frequency", "Specification", "Measured", "Deviation kHz"})
	gtm.currentStatus = append(gtm.currentStatus, []string{"", fmt.Sprintf("%.6f", gtm.intermediateFrequency/1e6),
		fmt.Sprintf("%.6f", frequency/1e6), fmt.Sprintf("%.3f", deviation/1e3)})
	gtm.result.FreqMeasurementCompleted = true
	gtm.result.FreqSpecMHz = gtm.intermediateFrequency / 1e6
	gtm.result.FreqMeasuredMHz = frequency / 1e6
	gtm.result.FreqDeviationkHz = deviation / 1e3
	var status = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Frequency Measurement Completed",
	}
	gtm.statusMonitor <- status
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result
	return nil
}

func (gtm *GroundTransmitterMeasurement) spuriousMeasurement(inband bool) error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Spurious Measurement Started",
	}
	gtm.statusMonitor <- measure
	spurType := ""
	gtm.currentStatus = append(gtm.currentStatus, []string{"Spurious", "Frequency", "Frequency Offset", "Level"})
	if inband {
		response := gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.inBandSpuriousSpectrum.span,
			gtm.inBandSpuriousSpectrum.rbw, gtm.inBandSpuriousSpectrum.vbw)
		if !response.Success {
			gtm.setError("Unable to communicate with SA")
			return fmt.Errorf("unable to communicate with SA")
		}
		spurType = "In-Band"
	} else {
		response := gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.outBandSpuriousSpectrum.span,
			gtm.outBandSpuriousSpectrum.rbw, gtm.outBandSpuriousSpectrum.vbw)
		if !response.Success {
			gtm.setError("Unable to communicate with SA")
			return fmt.Errorf("unable to communicate with SA")
		}
		spurType = "Out-Band"
	}
	response := gtm.sa.SetReferenceNominal()
	if !response.Success {
		gtm.setError("Carrier Not found")
		return fmt.Errorf("carrier not found")
	}
	noiseFloor := response.Result["MinValue"].Value
	powerOut := response.Result["ReferenceLevel"].Value - 10 + gtm.outputCableLoss
	time.Sleep(200 * time.Millisecond)
	response = gtm.sa.SetPeakThresholdAndExcursion(noiseFloor+10, 1)
	if !response.Success {
		gtm.setError("excursion cannot be set")
		return fmt.Errorf("excursion cannot be set")
	}
	response = gtm.sa.SetMaxHold()
	if !response.Success {
		gtm.setError("Max Hold cannot be set")
		return fmt.Errorf("max hold cannot be set")
	}
	response = gtm.sa.WaitForSweeps(5)
	if !response.Success {
		gtm.setError("Wait operation not completed")
		return fmt.Errorf("wait operation not completed")
	}

	powers := make([]float64, 0)
	frequencies := make([]float64, 0)
	powerOffsets := make([]float64, 0)
	frequencyOffset := make([]float64, 0)

	response = gtm.sa.GetMaxMarkerValue()
	if !response.Success {
		gtm.setError("SA Power Cannot be read")
		return fmt.Errorf("sa power cannot be read")
	}
	prevFreq := response.Result["MarkerX"].Value
	header := make([]string, 0)
	header = append(header, "Frequency Offset [kHz]", "Power Level [dBc]")
	rows := make([][]reports.DataCell, 0)
	if inband {
		gtm.result.InBandSpuriousFreqOffsetskHz = make([]float64, 0)
		gtm.result.InBandPowerOffsets = make([]float64, 0)
	} else {
		gtm.result.OutBandSpuriousFreqOffsetskHz = make([]float64, 0)
		gtm.result.OutBandPowerOffsets = make([]float64, 0)
	}

	for {
		response = gtm.sa.SetMarkerNextPeak(1)
		if !response.Success {
			gtm.setError("Marker cannot be set")
			return fmt.Errorf("marker cannot be set")
		}
		response = gtm.sa.GetMarkerValue(1)
		if !response.Success {
			gtm.setError("SA Power Cannot be read")
			return fmt.Errorf("sa power cannot be read")
		}

		spuriousValue := response.Result["MarkerY"].Value + gtm.outputCableLoss
		spuriousFreq := response.Result["MarkerX"].Value

		if spuriousFreq != prevFreq {
			frequencies = append(frequencies, spuriousFreq)
			frequencyOffset = append(frequencyOffset, spuriousFreq-gtm.intermediateFrequency)
			powers = append(powers, spuriousValue)
			powerOffsets = append(powerOffsets, spuriousValue-powerOut)
			prevFreq = spuriousFreq
			row := []reports.DataCell{reports.GetDataCell(fmt.Sprintf("%.3f", frequencyOffset)),
				reports.GetDataCell(fmt.Sprintf("%.2f", spuriousValue-powerOut))}
			rows = append(rows, row)
			gtm.currentStatus = append(gtm.currentStatus, []string{spurType, fmt.Sprintf("%.6f", spuriousFreq),
				fmt.Sprintf("%.3f", frequencyOffset), fmt.Sprintf("%.2f", spuriousValue-powerOut)})
		} else {
			response = gtm.sa.GetSpectrumDump()
			if !response.Success {
				gtm.setError("cannot take spectrum dump")
				return fmt.Errorf("cannot take spectrum dump")
			}
			image := reports.Images{
				FileData: response.Result["SpectrumDump"].String,
				Caption:  "Spurious Measurement " + spurType,
			}
			gtm.images = append(gtm.images, image)
			break
		}
	}
	if len(frequencyOffset) == 0 {
		row := []reports.DataCell{reports.GetDataCell("-"), reports.GetDataCell("-")}
		rows = append(rows, row)
		gtm.currentStatus = append(gtm.currentStatus, []string{"-", "-", "-", "-"})
	}
	if inband {
		for _, freq := range frequencyOffset {
			gtm.result.InBandSpuriousFreqOffsetskHz = append(gtm.result.InBandSpuriousFreqOffsetskHz, freq/1000)
		}
		gtm.result.InBandPowerOffsets = powerOffsets
		gtm.result.InBandSpuriousMeasurementCompleted = true
	} else {
		for _, freq := range frequencyOffset {
			gtm.result.OutBandSpuriousFreqOffsetskHz = append(gtm.result.OutBandSpuriousFreqOffsetskHz, freq/1000)
		}
		gtm.result.OutBandPowerOffsets = powerOffsets
		gtm.result.OutBandSpuriousMeasurementCompleted = true
	}
	gtm.report.SetResults("Spurious "+spurType, header, rows)
	gtm.order = append(gtm.order, "Spurious "+spurType)

	var sts = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Spurious Measurement Completed",
	}
	gtm.statusMonitor <- sts
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result
	return nil
}

func (gtm *GroundTransmitterMeasurement) harmonicsMeasurement() error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Harmonics Measurement Started",
	}
	gtm.statusMonitor <- measure
	gtm.result.HarmonicsFreqMHz = make([]float64, 0)
	gtm.result.HarmonicsMeasureddBm = make([]float64, 0)
	gtm.result.HarmonicsPresent = make([]bool, 0)
	gtm.result.HarmonicsNoiseFloor = make([]float64, 0)
	gtm.currentStatus = append(gtm.currentStatus, []string{"Harmonics", "Frequency", "Level", "Noise Floor"})
	header := []string{"Frequency [MHz]", "Level [dBm]", "Noise Floor [dBm]"}
	rows := make([][]reports.DataCell, 0)
	response := gtm.sa.SystemPreset()
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}
	response = gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrun.span,
		gtm.powerSpectrun.rbw, gtm.powerSpectrun.vbw)
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}

	response = gtm.sa.SetReferenceNominal()
	if !response.Success {
		gtm.setError("Carrier Not found")
		return fmt.Errorf("carrier not found")
	}
	noiseFloor := response.Result["MinValue"].Value
	time.Sleep(200 * time.Millisecond)

	for i := 0; i < gtm.noOfHarmonics; i = i + 1 {
		harmonicFreq := gtm.intermediateFrequency * (float64(i) + 2)
		response = gtm.sa.SetSpectrum(harmonicFreq, gtm.powerSpectrun.span, gtm.powerSpectrun.rbw, gtm.powerSpectrun.vbw)
		if !response.Success {
			gtm.setError("Unable to communicate with SA")
			return fmt.Errorf("unable to communicate with SA")
		}
		response = gtm.sa.CheckIfCarrierIsPresent()
		if !response.Success {
			gtm.setError("carrier present fucntiom error")
			return fmt.Errorf("unable to communicate with SA")
		}
		response = gtm.sa.GetMaxMinPeak()
		power := response.Result["MaxValue"].Value
		noiseFloor = response.Result["MinValue"].Value
		if !response.Result["Carrier"].Bool {
			gtm.currentStatus = append(gtm.currentStatus,
				[]string{"", fmt.Sprintf("%.6f", harmonicFreq), "NIL", fmt.Sprintf("%.2f", noiseFloor)})
			row := make([]reports.DataCell, 0)
			row = append(row, reports.GetDataCell(fmt.Sprintf("%.6f", harmonicFreq)))
			row = append(row, reports.GetDataCell("Nil"))
			row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", noiseFloor)))
			rows = append(rows, row)
			gtm.result.HarmonicsFreqMHz = append(gtm.result.HarmonicsFreqMHz, harmonicFreq)
			gtm.result.HarmonicsMeasureddBm = append(gtm.result.HarmonicsMeasureddBm, 0)
			gtm.result.HarmonicsPresent = append(gtm.result.HarmonicsPresent, false)
			gtm.result.HarmonicsNoiseFloor = append(gtm.result.HarmonicsNoiseFloor, noiseFloor)
		} else {
			gtm.currentStatus = append(gtm.currentStatus,
				[]string{"", fmt.Sprintf("%.6f", harmonicFreq), fmt.Sprintf("%.2f", power), fmt.Sprintf("%.2f", noiseFloor)})
			row := make([]reports.DataCell, 0)
			row = append(row, reports.GetDataCell(fmt.Sprintf("%.6f", harmonicFreq)))
			row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", power)))
			row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", noiseFloor)))
			rows = append(rows, row)
			gtm.result.HarmonicsFreqMHz = append(gtm.result.HarmonicsFreqMHz, harmonicFreq)
			gtm.result.HarmonicsMeasureddBm = append(gtm.result.HarmonicsMeasureddBm, power)
			gtm.result.HarmonicsPresent = append(gtm.result.HarmonicsPresent, true)
			gtm.result.HarmonicsNoiseFloor = append(gtm.result.HarmonicsNoiseFloor, noiseFloor)
		}
		response = gtm.sa.WaitForSweeps(2)
		if !response.Success {
			gtm.setError("cannot take spectrum dump")
			return fmt.Errorf("cannot take spectrum dump")
		}
		response = gtm.sa.GetSpectrumDump()
		if !response.Success {
			gtm.setError("cannot take spectrum dump")
			return fmt.Errorf("cannot take spectrum dump")
		}
		image := reports.Images{
			FileData: response.Result["SpectrumDump"].String,
			Caption:  "Harmonics Measuremet - " + fmt.Sprintf("%.2f MHz", harmonicFreq/1e6),
		}
		gtm.images = append(gtm.images, image)
		var sts = RTStatus{
			Completed: false,
			Success:   true,
			Error:     false,
			Message:   "Harmonics Measurement in progress",
		}
		gtm.statusMonitor <- sts
	}

	gtm.report.SetResults("Harmonics", header, rows)
	gtm.order = append(gtm.order, "Harmonics")
	gtm.result.HarmonicsMeasurementCompleted = true
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result

	var sts = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Harmonics Measurement Completed",
	}
	gtm.statusMonitor <- sts
	return nil
}

func (gtm *GroundTransmitterMeasurement) modIndexMeasurement() error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Mod Index Measurement Started",
	}
	gtm.statusMonitor <- measure
	gtm.currentStatus = append(gtm.currentStatus, []string{"Carrier", "Set MI", "Measured MI", "Deviation"})
	header := []string{"Modulation Type", "Set MI", "Measured MI", "Deviation"}
	rows := make([][]reports.DataCell, 0)

	response := gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrun.span,
		gtm.powerSpectrun.rbw, gtm.powerSpectrun.vbw)
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}

	response = gtm.sa.SetReferenceNominal()
	if !response.Success {
		gtm.setError("Carrier Not found")
		return fmt.Errorf("carrier not found")
	}
	time.Sleep(200 * time.Millisecond)

	response = gtm.gtx.SetOnlyTC(gtm.component)
	if !response.Success {
		gtm.setError("Only TC cannot be set")
		return fmt.Errorf("only TC cannot be set")
	}
	time.Sleep(1000 * time.Millisecond)

	response = gtm.gtx.SetModIndexTC(gtm.component, gtm.modIndex)
	if !response.Success {
		gtm.setError("Unable to set modulation index")
		return fmt.Errorf("unable to set modulation index")
	}
	response = gtm.gtx.SetModulationOn(gtm.component)
	if !response.Success {
		gtm.setError("Unable to set Modulation ON")
		return fmt.Errorf("unable to set modulation ON")
	}
	response = gtm.gtx.SetIdlePatternOn()
	if !response.Success {
		gtm.setError("Unable to set Idle Pattern ON")
		return fmt.Errorf("unable to set  Idle Pattern ON")
	}
	response = gtm.sa.GetModIndex(gtm.subCarrierFreq)
	mod1 := response.Result["modIndexForLeft"].Value
	mod2 := response.Result["modIndexForRight"].Value

	dev1 := math.Abs(mod1 - gtm.modIndex)
	dev2 := math.Abs(mod2 - gtm.modIndex)
	gtm.currentStatus = append(gtm.currentStatus, []string{"Left", fmt.Sprintf("%.2f", gtm.modIndex), fmt.Sprintf("%.2f", mod1), fmt.Sprintf("%.2f", dev1)})
	gtm.currentStatus = append(gtm.currentStatus, []string{"Right", fmt.Sprintf("%.2f", gtm.modIndex), fmt.Sprintf("%.2f", mod2), fmt.Sprintf("%.2f", dev2)})
	row1 := make([]reports.DataCell, 0)
	row2 := make([]reports.DataCell, 0)
	row1 = append(row1, reports.GetDataCell("Left"), reports.GetDataCell(fmt.Sprintf("%.2f", gtm.modIndex)),
		reports.GetDataCell(fmt.Sprintf("%.2f", mod1)), reports.GetDataCell(fmt.Sprintf("%.2f", dev1)))
	row2 = append(row2, reports.GetDataCell("Right"), reports.GetDataCell(fmt.Sprintf("%.2f", gtm.modIndex)),
		reports.GetDataCell(fmt.Sprintf("%.2f", mod2)), reports.GetDataCell(fmt.Sprintf("%.2f", dev2)))
	rows = append(rows, row1, row2)

	gtm.result.ModIndexSet = gtm.modIndex
	gtm.result.ModIndexMeasured = (mod1 + mod2) / 2
	gtm.result.ModIndexDeviation = (dev1 + dev2) / 2
	gtm.result.ModIndexMeasurementCompleted = true
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result

	response = gtm.sa.GetSpectrumDump()
	if !response.Success {
		gtm.setError("cannot take spectrum dump")
		return fmt.Errorf("cannot take spectrum dump")
	}
	image := reports.Images{
		FileData: response.Result["SpectrumDump"].String,
		Caption:  "Phase Modulation Measurement",
	}
	gtm.images = append(gtm.images, image)
	response = gtm.gtx.SetIdlePatternOff()
	if !response.Success {
		gtm.setError("Unable to set Idle Pattern Off")
		return fmt.Errorf("unable to set  Idle Pattern Off")
	}

	gtm.report.SetResults("Phase Modulation", header, rows)
	gtm.order = append(gtm.order, "Phase Modulation")

	measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Completed Measurement for PM",
	}
	gtm.statusMonitor <- measure
	return nil
}

func (gtm *GroundTransmitterMeasurement) freqDeviationMeasurement() error {
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Frequency Deviation Measurement Started",
	}
	gtm.statusMonitor <- measure
	gtm.currentStatus = append(gtm.currentStatus, []string{"Modulation", "Set Frequency Deviation", "Measured Frequency Deviation", "Deviation"})
	header := []string{"Modulation Type", "Set Frequency Deviation", "Measured Frequency Deviation", "Deviation"}
	rows := make([][]reports.DataCell, 0)

	response := gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrun.span,
		gtm.powerSpectrun.rbw, gtm.powerSpectrun.vbw)
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}

	response = gtm.sa.SetReferenceNominal()
	if !response.Success {
		gtm.setError("Carrier Not found")
		return fmt.Errorf("carrier not found")
	}
	time.Sleep(200 * time.Millisecond)

	response = gtm.gtx.SetOnlyTC(gtm.component)
	if !response.Success {
		gtm.setError("Only TC cannot be set")
		return fmt.Errorf("only TC cannot be set")
	}
	time.Sleep(1000 * time.Millisecond)

	response = gtm.gtx.SetFrequencyDeviationTC(gtm.component, gtm.freqdeviation)
	if !response.Success {
		gtm.setError("Unable to set frequency deviation")
		return fmt.Errorf("unable to set frequency deviation")
	}
	response = gtm.gtx.SetModulationOn(gtm.component)
	if !response.Success {
		gtm.setError("unable to set modulation ON")
		return fmt.Errorf("unable to set modulation ON")
	}
	response = gtm.gtx.SetIdlePatternOn()
	if !response.Success {
		gtm.setError("Unable to set Idle Pattern ON")
		return fmt.Errorf("unable to set  Idle Pattern ON")
	}
	response = gtm.sa.GetFrequencyDeviationFM(gtm.intermediateFrequency)
	if !response.Success {
		gtm.setError("unable to measure frequency deviation")
		return fmt.Errorf("unable to measure frequency deivation")
	}
	freqDev := response.Result["FrequencyDeviation"].Value
	dev := gtm.freqdeviation*2 - freqDev
	gtm.currentStatus = append(gtm.currentStatus, []string{"FM", fmt.Sprintf("%.2f", gtm.freqdeviation),
		fmt.Sprintf("%.2f", freqDev), fmt.Sprintf("%.2f", dev)})
	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell("FM"), reports.GetDataCell(fmt.Sprintf("%.2f", gtm.freqdeviation)),
		reports.GetDataCell(fmt.Sprintf("%.2f", freqDev)), reports.GetDataCell(fmt.Sprintf("%.2f", dev)))

	rows = append(rows, row)

	gtm.result.FrequencyDeviationSet = gtm.freqdeviation
	gtm.result.FrequencyDeviationMeasured = freqDev
	gtm.result.FrequencyDeviationDeviation = dev
	gtm.result.FrequencyDeviationMeasurementCompleted = true
	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result

	response = gtm.sa.GetSpectrumDump()
	if !response.Success {
		gtm.setError("cannot take spectrum dump")
		return fmt.Errorf("cannot take spectrum dump")
	}
	image := reports.Images{
		FileData: response.Result["SpectrumDump"].String,
		Caption:  "Frequency Deviation Measurement",
	}
	gtm.images = append(gtm.images, image)
	response = gtm.gtx.SetIdlePatternOff()
	if !response.Success {
		gtm.setError("Unable to set Idle Pattern Off")
		return fmt.Errorf("unable to set  Idle Pattern Off")
	}

	gtm.report.SetResults("Frequency Deviation", header, rows)
	gtm.order = append(gtm.order, "Frequency Deviation")

	measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Completed Measurement for Frequency Deviation",
	}
	gtm.statusMonitor <- measure

	return nil
}

func (gtm *GroundTransmitterMeasurement) phaseNoiseMeasurement() error {
	gtm.currentStatus = append(gtm.currentStatus, []string{"1 kHz", "10 kHz", "100 kHz", "1 MHz"})
	var measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Phase Noise Measurement In Progress",
	}
	gtm.statusMonitor <- measure

	header := []string{"1 kHz", "10 kHz", "100 kHz", "1 MHz"}
	rows := make([][]reports.DataCell, 0)

	response := gtm.gtx.SetModulationOff(gtm.component)
	if !response.Success {
		gtm.setError("Unable to communicate with GTx")
		return fmt.Errorf("unable to communicate with GTx")
	}
	response = gtm.gtx.SetCarrierOn(gtm.component)
	if !response.Success {
		gtm.setError("Unable to communicate with GTx")
		return fmt.Errorf("unable to communicate with GTx")
	}
	time.Sleep(200 * time.Millisecond)

	response = gtm.sa.SetSpectrum(math.Abs(gtm.intermediateFrequency), gtm.powerSpectrun.span,
		gtm.powerSpectrun.rbw, gtm.powerSpectrun.vbw)
	if !response.Success {
		gtm.setError("Unable to communicate with SA")
		return fmt.Errorf("unable to communicate with SA")
	}
	response = gtm.sa.SetReferenceNominal()
	if !response.Success {
		gtm.setError("Carrier Not found")
		return fmt.Errorf("unable to communicate with SA")
	}
	response = gtm.sa.SetPhaseNoiseMeasurement()
	if !response.Success {
		gtm.setError("SA cannot be set in phase noise mode")
		return fmt.Errorf("unable to communicate with SA")
	}
	time.Sleep(2 * time.Second)
	response = gtm.sa.SetPhaseNoiseMeasurement()
	if !response.Success {
		gtm.setError("SA cannot be set in phase noise mode")
		return fmt.Errorf("unable to communicate with SA")
	}
	time.Sleep(1 * time.Second)

	status := make([]string, 0)
	row := make([]reports.DataCell, 0)

	measureValue := 1000
	for i := 0; i < 4; i = i + 1 {
		response = gtm.sa.SetMarkerValuePhaseNoise(float64(measureValue), 1)
		if !response.Success {
			gtm.setError("Marker cannot be set in phase noise mode")
			return fmt.Errorf("unable to communicate with GTx")
		}
		response = gtm.sa.GetPhaseNoiseMarkerY(1)
		if !response.Success {
			gtm.setError("Marker cannot be set in phase noise mode")
			return fmt.Errorf("unable to communicate with GTx")
		}
		phase := response.Result["MarkerY"].Value
		status = append(status, fmt.Sprintf("%.2f", phase))
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", phase)))
		measureValue = measureValue * 10
	}
	time.Sleep(1000 * time.Millisecond)
	measureValue = 1000
	for i := 0; i < 4; i = i + 1 {
		response = gtm.sa.SetMarkerValuePhaseNoise(float64(measureValue), i+1)
		if !response.Success {
			gtm.setError("Marker cannot be set in phase noise mode")
			return fmt.Errorf("unable to communicate with GTx")
		}
		measureValue = measureValue * 10
	}
	time.Sleep(1000 * time.Millisecond)
	response = gtm.sa.GetSpectrumDump()
	if !response.Success {
		gtm.setError("cannot take spectrum dump")
		return fmt.Errorf("cannot take spectrum dump")
	}
	image := reports.Images{
		FileData: response.Result["SpectrumDump"].String,
		Caption:  "Phase Noise Measurement",
	}
	gtm.images = append(gtm.images, image)

	response = gtm.sa.SetSAMode()
	if !response.Success {
		gtm.setError("SA Cannot be put in Swept SA mode")
		return fmt.Errorf("sa Cannot be put in Swept SA mode")
	}
	rows = append(rows, row)

	gtm.report.SetResults("Phase Noise", header, rows)
	gtm.order = append(gtm.order, "Phase Noise")

	gtm.currentStatus = append(gtm.currentStatus, status)
	measure = RTStatus{
		Completed: false,
		Success:   true,
		Error:     false,
		Message:   "Completed Measurement for phaseNoise",
	}
	gtm.statusMonitor <- measure
	gtm.result.PhaseNoiseMeasurementCompleted = true
	gtm.result.PhaseNoiseAt1Khz, _ = strconv.ParseFloat(status[0], 64)
	gtm.result.PhaseNoiseAt10Khz, _ = strconv.ParseFloat(status[1], 64)
	gtm.result.PhaseNoiseAt100Khz, _ = strconv.ParseFloat(status[2], 64)
	gtm.result.PhaseNoiseAt1Mhz, _ = strconv.ParseFloat(status[3], 64)

	var result GTxResult
	result.Copy(gtm.result)
	gtm.resultMonitor <- result

	return nil
}

func (gtm *GroundTransmitterMeasurement) StartMeasurement() {
	defer gtm.stopMeasurement()

	err := gtm.startMeasurement()
	if err != nil {
		return
	}

	err = gtm.powerMeasurement()
	if err != nil {
		return
	}

	err = gtm.frequencyMeasurement()
	if err != nil {
		return
	}

	err = gtm.spuriousMeasurement(true)
	if err != nil {
		return
	}

	err = gtm.spuriousMeasurement(false)
	if err != nil {
		return
	}

	err = gtm.harmonicsMeasurement()
	if err != nil {
		return
	}

	if strings.EqualFold(gtm.modScheme, "PM") {
		err = gtm.modIndexMeasurement()
		if err != nil {
			return
		}
	}
	if strings.EqualFold(gtm.modScheme, "FM") {
		err = gtm.freqDeviationMeasurement()
		if err != nil {
			return
		}
	}

	err = gtm.phaseNoiseMeasurement()
	if err != nil {
		return
	}

	gtm.report.SetOrder(gtm.order)
	gtm.report.SetScreenshots(gtm.images)

	resultDir := utils.GetTNEResultDirectory()
	resultDir = filepath.Join(resultDir, "GTxMeasurement")
	_ = os.MkdirAll(resultDir, 0755)
	fileName := "GTxMeasurement"
	fileName = utils.GetTimeStampedFileName(fileName) + ".pdf"
	fileName = filepath.Join(resultDir, fileName)

	filename, err := reports.GenerateResult(gtm.report)
	if err != nil {
		return
	}

	err = os.Rename(filename, fileName)
	if err != nil {
		return
	}

	var measure = RTStatus{
		Completed: true,
		Success:   true,
		Error:     false,
		Message:   "Report Saved",
	}
	gtm.statusMonitor <- measure
	close(gtm.statusMonitor)
	close(gtm.resultMonitor)
}
