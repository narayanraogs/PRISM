package tne

import (
	"fmt"
	"os"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type TSMAttnMeasurement struct {
	deviceProfile    string
	rxName           string
	linkedRxs        []string
	spectrumProfile  string
	tsmConfiguration string
	maxPower         float64
	minPower         float64
	stepSize         float64
	frequency        int64
	tsmConfig        database.TSMConfiguration
	sa               driver.SA
	tsm              driver.TSM
	sg               driver.SG
	currentStatus    [][]string
	statusMonitor    chan AttnMeasurementStatus
	deviations       []CorrectedDeviation
	stop             bool
}

// Internal Helpers

func (tsm *TSMAttnMeasurement) notify(msg string) {
	tsm.statusMonitor <- AttnMeasurementStatus{Message: msg, Error: false}
}

func (tsm *TSMAttnMeasurement) finish(msg string, success bool) {
	tsm.statusMonitor <- AttnMeasurementStatus{
		Message:   msg,
		Error:     !success,
		Completed: true,
	}
	close(tsm.statusMonitor)
}

func (tsm *TSMAttnMeasurement) setError(msg string) {
	logger.Log.Error("TSM Attenuation Measurement Error", "error", msg)
	tsm.finish(msg, false)
}

func (tsm *TSMAttnMeasurement) check(resp utils.CommandResponse, errMsg string) bool {
	if !resp.Success {
		tsm.setError(fmt.Sprintf("%s: %s", errMsg, resp.ErrorMessage))
		return false
	}
	if tsm.stop {
		tsm.setError("Measurement aborted by user")
		return false
	}
	return true
}

// Public API

func (tsm *TSMAttnMeasurement) GetStatusMonitor() chan AttnMeasurementStatus {
	return tsm.statusMonitor
}

func (tsm *TSMAttnMeasurement) Initialize(deviceProfile, rxName, spectrumProfile, tsmConfiguration string,
	maxPower, minPower, stepSize float64) {
	tsm.deviceProfile = deviceProfile
	tsm.rxName = rxName
	tsm.spectrumProfile = spectrumProfile
	tsm.tsmConfiguration = tsmConfiguration
	tsm.maxPower = maxPower
	tsm.minPower = minPower
	tsm.stepSize = stepSize
	tsm.currentStatus = [][]string{{"Sl No", "Set Attn", "Measured Attn", "Deviation"}}
	tsm.statusMonitor = make(chan AttnMeasurementStatus, 20)
	tsm.stop = false

	if !tsm.loadDevices() {
		tsm.setError("Unable to load devices")
		return
	}
	if !tsm.loadDetails() {
		tsm.setError("Unable to load details from database")
	}
}

func (tsm *TSMAttnMeasurement) Stop() {
	tsm.stop = true
}

func (tsm *TSMAttnMeasurement) loadDevices() bool {
	saName, okSa := database.GetSAFromDeviceProfile(tsm.deviceProfile)
	sgName, okSg := database.GetSGFromDeviceProfile(tsm.deviceProfile)
	tsmName, okTsm := database.GetTSMFromDeviceProfile(tsm.deviceProfile)

	if !okSa || !okSg || !okTsm {
		return false
	}

	return tsm.sa.LoadDevice(saName) && tsm.sg.LoadDevice(sgName) && tsm.tsm.LoadDevice(tsmName)
}

func (tsm *TSMAttnMeasurement) loadDetails() bool {
	freq, okFreq := database.GetRxFrequency(tsm.rxName)
	rxs, okRxs := database.GetAllRxWithFrequency(freq)
	paths, okPath := database.GetTSMPathDetails(tsm.tsmConfiguration)

	if !okFreq || !okRxs || !okPath {
		return false
	}

	tsm.frequency = freq
	tsm.linkedRxs = rxs
	tsm.tsmConfig = paths
	return true
}

func (tsm *TSMAttnMeasurement) StartMeasurement() {
	logger.Log.Info("Starting TSM Attenuation Measurement", "rxName", tsm.rxName, "frequency", tsm.frequency, "maxPower", tsm.maxPower, "minPower", tsm.minPower)
	tsm.notify("TSM Attenuation Measurement Started")

	if !tsm.check(tsm.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !tsm.check(tsm.tsm.SetDriverStatus(tsm.tsmConfig.UplinkToSC.String), "TSM: uplink to SC") {
		return
	}
	time.Sleep(500 * time.Millisecond)

	attnNum := int(tsm.tsmConfig.AttnNumber.Int64)
	if !tsm.check(tsm.tsm.SetAttn(attnNum, 0), "TSM: reset attn") {
		return
	}
	if !tsm.check(tsm.sg.SetPower(0), "SG: set power 0") {
		return
	}
	if !tsm.check(tsm.sg.SetFrequency(float64(tsm.frequency)), "SG: set freq") {
		return
	}
	if !tsm.check(tsm.sg.SetModOff(), "SG: mod off") {
		return
	}
	if !tsm.check(tsm.sg.SetRFOn(), "SG: RF on") {
		return
	}

	defer func() {
		tsm.tsm.SetDriverStatus(tsm.tsmConfig.TerminateUplink.String)
		time.Sleep(500 * time.Millisecond)
		tsm.sa.SetAlignmentOn()
		tsm.sa.SystemPreset()
		tsm.sg.SetRFOff()
		tsm.tsm.SetAttn(attnNum, 0)
	}()

	spec, ok := database.GetSpectrumProfile(tsm.spectrumProfile)
	if !ok {
		tsm.setError("Spectrum profile not found")
		return
	}

	if !tsm.check(tsm.sa.SetSpectrum(spec.CenterFrequency, spec.Span, float64(spec.RBW), float64(spec.VBW)), "SA: set spectrum") {
		return
	}

	if tsm.tsmConfig.ExcludePad.Valid {
		if !tsm.check(tsm.tsm.SetDriverStatus(tsm.tsmConfig.ExcludePad.String), "TSM: exclude pad") {
			return
		}
	}

	tsm.notify("Measuring initial power")
	if !tsm.check(tsm.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	resp := tsm.sa.GetMaxMarkerValue()
	if !tsm.check(resp, "SA: read initial power") {
		return
	}
	initialPower := resp.Result["MarkerY"].Value
	slNo := 0

	// Fixed Pad Measurement
	var fixedPower float64
	if tsm.tsmConfig.IncludePad.Valid {
		tsm.notify("Measuring Fixed Pad Attenuation")
		if !tsm.check(tsm.tsm.SetDriverStatus(tsm.tsmConfig.IncludePad.String), "TSM: include pad") {
			return
		}
		time.Sleep(200 * time.Millisecond)

		pResp := tsm.sa.GetMaxMarkerValue()
		if !tsm.check(pResp, "SA: read pad power") {
			return
		}
		fixedPower = pResp.Result["MarkerY"].Value - initialPower
		slNo++

		if !tsm.check(tsm.tsm.SetDriverStatus(tsm.tsmConfig.ExcludePad.String), "TSM: exclude pad restore") {
			return
		}

		measure := AttnMeasurementStatus{Message: "Completed Measuring Fixed Pad Attenuation"}
		measure.AddData(slNo, 0, fixedPower, 0)
		tsm.statusMonitor <- measure
		tsm.currentStatus = append(tsm.currentStatus, []string{strconv.Itoa(slNo), "FixedPad", fmt.Sprintf("%.3f", fixedPower), "-"})
		time.Sleep(200 * time.Millisecond)
	}

	// Attenuation Sweep Loop
	for attn := tsm.minPower; attn <= tsm.maxPower; attn += tsm.stepSize {
		if tsm.stop {
			tsm.setError("User Aborted")
			return
		}

		if !tsm.check(tsm.tsm.SetAttn(attnNum, attn), "TSM: set attn") {
			return
		}
		time.Sleep(200 * time.Millisecond)

		mResp := tsm.sa.GetMaxMarkerValue()
		if !tsm.check(mResp, "SA: read attenuation") {
			return
		}

		power := mResp.Result["MarkerY"].Value
		actualAttn := initialPower - power
		diff := attn - actualAttn
		slNo++

		measure := AttnMeasurementStatus{
			Message:       fmt.Sprintf("Completed Measuring for %.3f dB", attn),
			PlotDeviation: true,
		}
		measure.AddData(slNo, attn, actualAttn, diff)
		logger.Log.Info("TSM Attn Measurement Point", "setAttn", attn, "measuredAttn", actualAttn, "deviation", diff)
		tsm.statusMonitor <- measure
		tsm.currentStatus = append(tsm.currentStatus, []string{strconv.Itoa(slNo), fmt.Sprintf("%.3f", attn), fmt.Sprintf("%.3f", actualAttn), fmt.Sprintf("%.3f", diff)})
		time.Sleep(200 * time.Millisecond)
	}

	tsm.notify("Saving Results")
	tsm.saveAndCalculate(fixedPower)
	logger.Log.Info("Completed TSM Attenuation Measurement", "success", true)
	tsm.finish("Measurement Completed", true)
}

func (tsm *TSMAttnMeasurement) saveAndCalculate(fixedPad float64) {
	var csv strings.Builder
	var reqs, measureds, diffs []float64

	for i, row := range tsm.currentStatus {
		csv.WriteString(strings.Join(row, ","))
		csv.WriteString("\n")

		if i == 0 {
			continue
		}

		r, _ := strconv.ParseFloat(row[1], 64) // r will be 0 for FixedPad row which is fine
		m, _ := strconv.ParseFloat(row[2], 64)
		d, _ := strconv.ParseFloat(row[3], 64)
		reqs = append(reqs, r)
		measureds = append(measureds, m)
		diffs = append(diffs, d)
	}

	// Calculate Corrected Deviations
	provider := utils.TSMAttnProvider{RequiredAttn: reqs, MeasuredAttn: measureds, Difference: diffs}
	corrected := utils.GetCorrectedProfile(provider, fixedPad, tsm.stepSize)

	tsm.deviations = nil
	// Skip the first row if it was FixedPad (reqs[0] == 0)
	startIdx := 0
	if tsm.tsmConfig.IncludePad.Valid {
		startIdx = 1
	}

	for i := startIdx; i < len(reqs); i++ {
		tsm.deviations = append(tsm.deviations, CorrectedDeviation{
			SetValue:           reqs[i],
			MeasuredDeviation:  diffs[i],
			CorrectedDeviation: corrected.GetDeviation(reqs[i]),
		})
	}

	// Write CSV for each linked RX
	for _, rx := range tsm.linkedRxs {
		path := fmt.Sprintf("%s/.resources/tsm-%s.csv", utils.Config.BaseFolder, rx)
		if err := os.WriteFile(path, []byte(csv.String()), 0644); err != nil {
			logger.Log.Error("Failed to write TSM Attn CSV", "rx", rx, "error", err)
		}
	}
}

func (tsm *TSMAttnMeasurement) GetCorrectedDeviations() []CorrectedDeviation {
	return tsm.deviations
}
