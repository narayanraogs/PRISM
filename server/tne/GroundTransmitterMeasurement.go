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

type GroundTransmitterMeasurement struct {
	deviceProfile           string
	gtxName                 string
	component               string
	modScheme               string
	intermediateFrequency   float64
	outputCableLoss         float64
	subCarrierFreq          float64
	freqDeviation           float64
	modIndex                float64
	noOfHarmonics           int
	powerSpectrum           spectrumSettings
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
	statusMonitor chan RTStatus
	resultMonitor chan GTxResult
	stop          bool
}

func NewGTxGroundTransmitterMeasurement() *GroundTransmitterMeasurement {
	return &GroundTransmitterMeasurement{
		noOfHarmonics: 3,
	}
}

func (gtm *GroundTransmitterMeasurement) notify(msg string, completed bool) {
	gtm.statusMonitor <- RTStatus{
		Completed: completed,
		Success:   true,
		Error:     false,
		Message:   msg,
	}
}

func (gtm *GroundTransmitterMeasurement) fail(msg string) {
	gtm.statusMonitor <- RTStatus{
		Completed: true,
		Success:   false,
		Error:     true,
		Message:   msg,
	}
	close(gtm.resultMonitor)
	close(gtm.statusMonitor)
}

func (gtm *GroundTransmitterMeasurement) check(resp utils.CommandResponse, errorMsg string) error {
	if !resp.Success {
		gtm.fail(errorMsg)
		return fmt.Errorf("%s: %v", errorMsg, resp.ErrorMessage)
	}
	return nil
}

func (gtm *GroundTransmitterMeasurement) publishResult() {
	var res GTxResult
	res.Copy(gtm.result)
	gtm.resultMonitor <- res
}

func (gtm *GroundTransmitterMeasurement) addReportRow(section string, header []string, values []string) {
	row := make([]reports.DataCell, len(values))
	for i, v := range values {
		row[i] = reports.GetDataCell(v)
	}
	if _, exists := gtm.report.Results[section]; !exists {
		gtm.report.SetResults(section, header, [][]reports.DataCell{row})
		gtm.order = append(gtm.order, section)
	} else {
		res := gtm.report.Results[section]
		res.Data = append(res.Data, row)
		gtm.report.Results[section] = res
	}
}

func (gtm *GroundTransmitterMeasurement) captureSpectrum(caption string) error {
	resp := gtm.sa.GetSpectrumDump()
	if err := gtm.check(resp, "Cannot take spectrum dump"); err != nil {
		return err
	}
	gtm.images = append(gtm.images, reports.Images{
		FileData: resp.Result["SpectrumDump"].String,
		Caption:  caption,
	})
	return nil
}

func (gtm *GroundTransmitterMeasurement) GetStatusMonitor() (chan RTStatus, chan GTxResult) {
	return gtm.statusMonitor, gtm.resultMonitor
}

func (gtm *GroundTransmitterMeasurement) SetDevices(deviceProfile string, component string, intermediateFrequency float64, outputCableLoss float64) bool {
	gtm.images = make([]reports.Images, 0)
	gtm.statusMonitor = make(chan RTStatus, 10)
	gtm.resultMonitor = make(chan GTxResult, 10)
	gtm.deviceProfile = deviceProfile
	gtm.component = component
	gtm.intermediateFrequency = intermediateFrequency
	gtm.outputCableLoss = math.Abs(outputCableLoss)

	gtxName, ok1 := database.GetGTxFromDeviceProfile(deviceProfile)
	saName, ok2 := database.GetSAFromDeviceProfile(deviceProfile)
	if !ok1 || !ok2 {
		gtm.fail("Unable to load devices from Database")
		return false
	}
	gtm.gtxName = gtxName

	if !gtm.gtx.LoadDevice(gtxName) || !gtm.sa.LoadDevice(saName) {
		gtm.fail("Unable to initialize device drivers")
		return false
	}
	testPhase, _ := database.GetSelectedTestPhase()
	gtm.report.SetHeader("", "GTX Measurement", "", testPhase)
	gtm.report.Results = make(map[string]reports.Result)
	gtm.report.Order = []string{}
	gtm.order = []string{}

	return true
}

func (gtm *GroundTransmitterMeasurement) SetModulationParameters(modScheme string, subCarrierFreq float64, freqDeviation float64, modIndex float64) {
	gtm.modScheme = modScheme
	gtm.subCarrierFreq = subCarrierFreq
	gtm.freqDeviation = freqDeviation
	gtm.modIndex = modIndex

	gtm.result.FrequencyDeviationApplicable = strings.EqualFold(modScheme, "FM")
	gtm.result.ModIndexApplicable = strings.EqualFold(modScheme, "PM")

	gtm.publishResult()
}

func (gtm *GroundTransmitterMeasurement) SetPowerSpectrum(span, rbw, vbw float64) {
	gtm.powerSpectrum = spectrumSettings{span, rbw, vbw}
}

func (gtm *GroundTransmitterMeasurement) SetFrequencySpectrum(span, rbw, vbw float64) {
	gtm.frequencySpectrum = spectrumSettings{span, rbw, vbw}
}

func (gtm *GroundTransmitterMeasurement) SetInBandSpectrum(span, rbw, vbw float64) {
	gtm.inBandSpuriousSpectrum = spectrumSettings{span, rbw, vbw}
}

func (gtm *GroundTransmitterMeasurement) SetOutBandSpectrum(span, rbw, vbw float64) {
	gtm.outBandSpuriousSpectrum = spectrumSettings{span, rbw, vbw}
}

func (gtm *GroundTransmitterMeasurement) Stop() {
	gtm.stop = true
}

func (gtm *GroundTransmitterMeasurement) startMeasurement() error {
	gtm.notify("Measurement Started", false)

	if err := gtm.check(gtm.sa.SetAlignmentOff(), "Unable to communicate with SA (Alignment)"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SystemPreset(), "Unable to communicate with SA (Preset)"); err != nil {
		return err
	}
	if err := gtm.check(gtm.gtx.SetFrequency(gtm.component, gtm.intermediateFrequency), "Unable to communicate with GTx (Freq)"); err != nil {
		return err
	}

	gtm.notify("GTx Frequency Set", false)

	if err := gtm.check(gtm.gtx.SetPower(gtm.component, 0), "Unable to communicate with GTx (Power)"); err != nil {
		return err
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
	gtm.notify("Power Measurement Started", false)

	if err := gtm.check(gtm.gtx.SetModulationOff(gtm.component), "GTx: modulation off"); err != nil {
		return err
	}
	if err := gtm.check(gtm.gtx.SetCarrierOn(gtm.component), "GTx: carrier on"); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	if err := gtm.check(gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrum.span, gtm.powerSpectrum.rbw, gtm.powerSpectrum.vbw), "SA: set spectrum"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.WaitForSweeps(5), "SA: wait for sweeps"); err != nil {
		return err
	}
	resp := gtm.sa.SetReferenceNominal()
	if err := gtm.check(resp, "Carrier not found"); err != nil {
		return err
	}
	gtm.sa.WaitForSweeps(2)

	power := resp.Result["ReferenceLevel"].Value - 10 + gtm.outputCableLoss
	header := []string{"", "Specification [dBm]", "Measured [dBm]", "Deviation [dB]"}
	values := []string{"Power", "0", fmt.Sprintf("%.2f", power), fmt.Sprintf("%.2f", -power)}

	gtm.addReportRow("Power", header, values)

	gtm.sa.SetMaxHold()
	gtm.sa.WaitForSweeps(2)

	if err := gtm.captureSpectrum("Power Measurement"); err != nil {
		return err
	}
	gtm.sa.SetNormalMode()

	gtm.result.PowerMeasurementCompleted = true
	gtm.result.PowerSpec = 0
	gtm.result.PowerMeasured = power
	gtm.result.PowerDeviation = -power

	gtm.publishResult()
	gtm.notify("Power Measurement Completed", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) frequencyMeasurement() error {
	gtm.notify("Frequency Measurement Started", false)

	if err := gtm.check(gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.frequencySpectrum.span, gtm.frequencySpectrum.rbw, gtm.frequencySpectrum.vbw), "SA: set spectrum"); err != nil {
		return err
	}
	gtm.sa.WaitForSweeps(5)

	if err := gtm.check(gtm.sa.PeakSearch(true, 1), "SA: peak search failed"); err != nil {
		return err
	}
	resp := gtm.sa.GetFrequencyInCounterMode(1)
	if err := gtm.check(resp, "SA: frequency counter failed"); err != nil {
		return err
	}

	frequency := resp.Result["Frequency"].Value - 10
	deviation := gtm.intermediateFrequency - frequency

	header := []string{"", "Specification [MHz]", "Measured [MHz]", "Deviation [kHz]", "Deviation PPM"}
	values := []string{"Frequency",
		fmt.Sprintf("%.6f", gtm.intermediateFrequency/1e6),
		fmt.Sprintf("%.6f", frequency/1e6),
		fmt.Sprintf("%.3f", deviation/1e3),
		fmt.Sprintf("%.2f", deviation/2*1e-6),
	}

	gtm.addReportRow("Frequency", header, values)

	if err := gtm.captureSpectrum("Frequency Measurement"); err != nil {
		return err
	}

	gtm.result.FreqMeasurementCompleted = true
	gtm.result.FreqSpecMHz = gtm.intermediateFrequency / 1e6
	gtm.result.FreqMeasuredMHz = frequency / 1e6
	gtm.result.FreqDeviationkHz = deviation / 1e3

	gtm.publishResult()
	gtm.notify("Frequency Measurement Completed", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) spuriousMeasurement(inBand bool) error {
	gtm.notify("Spurious Measurement Started", false)

	var spec spectrumSettings
	var spurType string
	if inBand {
		spec = gtm.inBandSpuriousSpectrum
		spurType = "In-Band"
	} else {
		spec = gtm.outBandSpuriousSpectrum
		spurType = "Out-Band"
	}

	if err := gtm.check(gtm.sa.SetSpectrum(gtm.intermediateFrequency, spec.span, spec.rbw, spec.vbw), "SA: set spectrum"); err != nil {
		return err
	}

	resp := gtm.sa.SetReferenceNominal()
	if err := gtm.check(resp, "Carrier not found"); err != nil {
		return err
	}

	noiseFloor := resp.Result["MinValue"].Value
	powerOut := resp.Result["ReferenceLevel"].Value - 10 + gtm.outputCableLoss

	time.Sleep(200 * time.Millisecond)

	if err := gtm.check(gtm.sa.SetPeakThresholdAndExcursion(noiseFloor+15, 1), "SA: excursion error"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SetMaxHold(), "SA: max hold error"); err != nil {
		return err
	}
	gtm.sa.WaitForSweeps(5)

	// Marker Loop
	resp = gtm.sa.GetMaxMarkerValue()
	if err := gtm.check(resp, "SA: cannot read power"); err != nil {
		return err
	}

	prevFreq := resp.Result["MarkerX"].Value
	header := []string{"", "Frequency [MHz]", "Frequency Offset [kHz]", "Level [dBc]"}

	var powerOffsets, freqOffsets []float64

	for {
		if err := gtm.check(gtm.sa.SetMarkerNextPeak(1), "SA: marker next peak failed"); err != nil {
			return err
		}
		mResp := gtm.sa.GetMarkerValue(1)
		if err := gtm.check(mResp, "SA: cannot read marker"); err != nil {
			return err
		}

		spuriousVal := mResp.Result["MarkerY"].Value + gtm.outputCableLoss
		spuriousFreq := mResp.Result["MarkerX"].Value

		if spuriousFreq == prevFreq {
			break // No more unique peaks
		}

		currentOffset := spuriousFreq - gtm.intermediateFrequency
		currentPowerOffset := spuriousVal - powerOut

		gtm.addReportRow("Spurious "+spurType, header, []string{
			spurType,
			fmt.Sprintf("%.6f", spuriousFreq/1e6),
			fmt.Sprintf("%.3f", currentOffset/1e3),
			fmt.Sprintf("%.2f", currentPowerOffset),
		})

		freqOffsets = append(freqOffsets, currentOffset/1e3)
		powerOffsets = append(powerOffsets, currentPowerOffset)
		prevFreq = spuriousFreq
	}

	if len(freqOffsets) == 0 {
		gtm.addReportRow("Spurious "+spurType, header, []string{"-", "-", "-", "-"})
	}

	if inBand {
		gtm.result.InBandSpuriousFreqOffsetskHz = freqOffsets
		gtm.result.InBandPowerOffsets = powerOffsets
		gtm.result.InBandSpuriousMeasurementCompleted = true
	} else {
		gtm.result.OutBandSpuriousFreqOffsetskHz = freqOffsets
		gtm.result.OutBandPowerOffsets = powerOffsets
		gtm.result.OutBandSpuriousMeasurementCompleted = true
	}

	if err := gtm.captureSpectrum("Spurious Measurement " + spurType); err != nil {
		return err
	}

	gtm.publishResult()
	gtm.notify("Spurious Measurement Completed", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) harmonicsMeasurement() error {
	gtm.notify("Harmonics Measurement Started", false)

	gtm.result.HarmonicsFreqMHz = []float64{}
	gtm.result.HarmonicsMeasureddBm = []float64{}
	gtm.result.HarmonicsPresent = []bool{}
	gtm.result.HarmonicsNoiseFloor = []float64{}

	header := []string{"", "Frequency [MHz]", "Level [dBm]", "Noise Floor [dBm]"}

	if err := gtm.check(gtm.sa.SystemPreset(), "SA: preset fail"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrum.span, gtm.powerSpectrum.rbw, gtm.powerSpectrum.vbw), "SA: set spectrum"); err != nil {
		return err
	}

	resp := gtm.sa.SetReferenceNominal()
	if err := gtm.check(resp, "Carrier not found"); err != nil {
		return err
	}

	time.Sleep(200 * time.Millisecond)

	for i := 0; i < gtm.noOfHarmonics; i++ {
		harmonicFreq := gtm.intermediateFrequency * float64(i+2)

		if err := gtm.check(gtm.sa.SetSpectrum(harmonicFreq, gtm.powerSpectrum.span, gtm.powerSpectrum.rbw, gtm.powerSpectrum.vbw), "SA: set spectrum (harmonic)"); err != nil {
			return err
		}

		gtm.sa.CheckIfCarrierIsPresent()
		mResp := gtm.sa.GetMaxMinPeak()
		if err := gtm.check(mResp, "SA: measure harmonic failed"); err != nil {
			return err
		}

		power := mResp.Result["MaxValue"].Value
		noiseFloor := mResp.Result["MinValue"].Value
		isPresent := mResp.Result["Carrier"].Bool

		levelStr := "Nil"
		powerVal := 0.0
		if isPresent {
			levelStr = fmt.Sprintf("%.2f", power)
			powerVal = power
		}

		gtm.addReportRow("Harmonics", header, []string{
			"Harmonic",
			fmt.Sprintf("%.6f", harmonicFreq/1e6),
			levelStr,
			fmt.Sprintf("%.2f", noiseFloor),
		})

		gtm.result.HarmonicsFreqMHz = append(gtm.result.HarmonicsFreqMHz, harmonicFreq/1e6)
		gtm.result.HarmonicsMeasureddBm = append(gtm.result.HarmonicsMeasureddBm, powerVal)
		gtm.result.HarmonicsPresent = append(gtm.result.HarmonicsPresent, isPresent)
		gtm.result.HarmonicsNoiseFloor = append(gtm.result.HarmonicsNoiseFloor, noiseFloor)

		gtm.sa.WaitForSweeps(2)
		gtm.captureSpectrum(fmt.Sprintf("Harmonics Measurement - %.2f MHz", harmonicFreq/1e6))
		gtm.notify("Harmonics Measurement in progress", false)
	}

	gtm.result.HarmonicsMeasurementCompleted = true
	gtm.publishResult()
	gtm.notify("Harmonics Measurement Completed", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) modIndexMeasurement() error {
	gtm.notify("Mod Index Measurement Started", false)

	header := []string{"", "Set MI", "Measured MI", "Deviation"}

	if err := gtm.check(gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrum.span, gtm.powerSpectrum.rbw, gtm.powerSpectrum.vbw), "SA: set spectrum"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SetReferenceNominal(), "Carrier not found"); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	steps := []struct {
		fn  func() utils.CommandResponse
		msg string
	}{
		{func() utils.CommandResponse { return gtm.gtx.SetOnlyTC(gtm.component) }, "GTx: only TC fail"},
		{func() utils.CommandResponse { return gtm.gtx.SetModIndexTC(gtm.component, gtm.modIndex) }, "GTx: set mod index fail"},
		{func() utils.CommandResponse { return gtm.gtx.SetModulationOn(gtm.component) }, "GTx: mod on fail"},
		{func() utils.CommandResponse { return gtm.gtx.SetIdlePatternOn() }, "GTx: idle pattern on fail"},
	}

	for i, step := range steps {
		if err := gtm.check(step.fn(), step.msg); err != nil {
			return err
		}
		if i == 0 {
			time.Sleep(1000 * time.Millisecond)
		}
	}

	resp := gtm.sa.GetModIndex(gtm.subCarrierFreq)
	if err := gtm.check(resp, "SA: measure MI fail"); err != nil {
		return err
	}

	modL := resp.Result["modIndexForLeft"].Value
	modR := resp.Result["modIndexForRight"].Value
	devL := math.Abs(modL - gtm.modIndex)
	devR := math.Abs(modR - gtm.modIndex)

	gtm.addReportRow("Phase Modulation", header, []string{"Left", fmt.Sprintf("%.2f", gtm.modIndex), fmt.Sprintf("%.2f", modL), fmt.Sprintf("%.2f", devL)})
	gtm.addReportRow("Phase Modulation", header, []string{"Right", fmt.Sprintf("%.2f", gtm.modIndex), fmt.Sprintf("%.2f", modR), fmt.Sprintf("%.2f", devR)})

	gtm.result.ModIndexSet = gtm.modIndex
	gtm.result.ModIndexMeasured = (modL + modR) / 2
	gtm.result.ModIndexDeviation = (devL + devR) / 2
	gtm.result.ModIndexMeasurementCompleted = true

	gtm.captureSpectrum("Phase Modulation Measurement")
	gtm.gtx.SetIdlePatternOff()
	gtm.publishResult()
	gtm.notify("Completed Measurement for PM", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) freqDeviationMeasurement() error {
	gtm.notify("Frequency Deviation Measurement Started", false)

	header := []string{"", "Set Freq Dev", "Measured Freq Dev", "Deviation"}

	if err := gtm.check(gtm.sa.SetSpectrum(gtm.intermediateFrequency, gtm.powerSpectrum.span, gtm.powerSpectrum.rbw, gtm.powerSpectrum.vbw), "SA: set spectrum"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SetReferenceNominal(), "Carrier not found"); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	if err := gtm.check(gtm.gtx.SetOnlyTC(gtm.component), "GTx: only TC fail"); err != nil {
		return err
	}
	time.Sleep(1000 * time.Millisecond)
	if err := gtm.check(gtm.gtx.SetFrequencyDeviationTC(gtm.component, gtm.freqDeviation), "GTx: set dev fail"); err != nil {
		return err
	}
	if err := gtm.check(gtm.gtx.SetModulationOn(gtm.component), "GTx: mod on fail"); err != nil {
		return err
	}
	if err := gtm.check(gtm.gtx.SetIdlePatternOn(), "GTx: idle pattern on fail"); err != nil {
		return err
	}

	resp := gtm.sa.GetFrequencyDeviationFM(gtm.intermediateFrequency)
	if err := gtm.check(resp, "SA: measure freq dev fail"); err != nil {
		return err
	}

	freqDev := resp.Result["FrequencyDeviation"].Value
	dev := gtm.freqDeviation*2 - freqDev

	gtm.addReportRow("Frequency Deviation", header, []string{"FM", fmt.Sprintf("%.2f", gtm.freqDeviation), fmt.Sprintf("%.2f", freqDev), fmt.Sprintf("%.2f", dev)})

	gtm.result.FrequencyDeviationSet = gtm.freqDeviation
	gtm.result.FrequencyDeviationMeasured = freqDev
	gtm.result.FrequencyDeviationDeviation = dev
	gtm.result.FrequencyDeviationMeasurementCompleted = true

	gtm.captureSpectrum("Frequency Deviation Measurement")
	gtm.gtx.SetIdlePatternOff()
	gtm.publishResult()
	gtm.notify("Completed Measurement for Frequency Deviation", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) phaseNoiseMeasurement() error {
	gtm.notify("Phase Noise Measurement In Progress", false)

	header := []string{"", "1 kHz", "10 kHz", "100 kHz", "1 MHz"}

	if err := gtm.check(gtm.gtx.SetModulationOff(gtm.component), "GTx: mod off"); err != nil {
		return err
	}
	if err := gtm.check(gtm.gtx.SetCarrierOn(gtm.component), "GTx: carrier on"); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	if err := gtm.check(gtm.sa.SetSpectrum(math.Abs(gtm.intermediateFrequency), gtm.powerSpectrum.span, gtm.powerSpectrum.rbw, gtm.powerSpectrum.vbw), "SA: set spectrum"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SetReferenceNominal(), "Carrier not found"); err != nil {
		return err
	}
	if err := gtm.check(gtm.sa.SetPhaseNoiseMeasurement(), "SA: phase noise mode fail"); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	gtm.sa.SetPhaseNoiseMeasurement() // Double toggle as per original code
	time.Sleep(1 * time.Second)

	offsets := []int{1000, 10000, 100000, 1000000}
	results := make([]string, 0, 4)

	for _, offset := range offsets {
		if err := gtm.check(gtm.sa.SetMarkerValuePhaseNoise(float64(offset), 1), "SA: set marker fail"); err != nil {
			return err
		}
		resp := gtm.sa.GetPhaseNoiseMarkerY(1)
		if err := gtm.check(resp, "SA: get marker fail"); err != nil {
			return err
		}
		results = append(results, fmt.Sprintf("%.2f", resp.Result["MarkerY"].Value))
	}

	for i, offset := range offsets {
		gtm.sa.SetMarkerValuePhaseNoise(float64(offset), i+1)
	}
	time.Sleep(1 * time.Second)
	gtm.captureSpectrum("Phase Noise Measurement")
	gtm.sa.SetSAMode()

	gtm.addReportRow("Phase Noise", header, append([]string{"Phase Noise"}, results...))

	gtm.result.PhaseNoiseMeasurementCompleted = true
	gtm.result.PhaseNoiseAt1Khz, _ = strconv.ParseFloat(results[0], 64)
	gtm.result.PhaseNoiseAt10Khz, _ = strconv.ParseFloat(results[1], 64)
	gtm.result.PhaseNoiseAt100Khz, _ = strconv.ParseFloat(results[2], 64)
	gtm.result.PhaseNoiseAt1Mhz, _ = strconv.ParseFloat(results[3], 64)

	gtm.publishResult()
	gtm.notify("Completed Measurement for phaseNoise", false)
	return nil
}

func (gtm *GroundTransmitterMeasurement) StartMeasurement() {
	defer gtm.stopMeasurement()

	stages := []func() error{
		gtm.startMeasurement,
		gtm.powerMeasurement,
		gtm.frequencyMeasurement,
		func() error { return gtm.spuriousMeasurement(true) },
		func() error { return gtm.spuriousMeasurement(false) },
		gtm.harmonicsMeasurement,
	}

	gtm.order = []string{
		"Power",
		"Frequency",
		"Spurious In-Band",
		"Spurious Out-Band",
		"Harmonics",
	}

	for _, stage := range stages {
		if err := stage(); err != nil {
			return
		}
	}

	// Modulation specific stages
	if strings.EqualFold(gtm.modScheme, "PM") {
		if err := gtm.modIndexMeasurement(); err != nil {
			return
		}
		gtm.order = append(gtm.order, "Phase Modulation")
	} else if strings.EqualFold(gtm.modScheme, "FM") {
		if err := gtm.freqDeviationMeasurement(); err != nil {
			return
		}
		gtm.order = append(gtm.order, "Frequency Deviation")
	}

	if err := gtm.phaseNoiseMeasurement(); err != nil {
		return
	}
	gtm.order = append(gtm.order, "Phase Noise")

	gtm.report.SetOrder(gtm.order)
	gtm.report.SetScreenshots(gtm.images)

	resultDir := filepath.Join(utils.GetTNEResultDirectory(), "GTxMeasurement")
	_ = os.MkdirAll(resultDir, 0755)

	tempFile, err := reports.GenerateResult(gtm.report, true, false, false, false, true)
	if err != nil {
		gtm.fail("Failed to generate PDF")
		return
	}

	fileName := utils.GetTimeStampedFileName("GTxMeasurement") + ".pdf"
	finalPath := filepath.Join(resultDir, fileName)

	if err := os.Rename(tempFile, finalPath); err != nil {
		gtm.fail("Failed to save PDF report")
		return
	}

	gtm.notify("Report Saved", true)
	close(gtm.resultMonitor)
	close(gtm.statusMonitor)
}
