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

type SGPowerMeasurement struct {
	deviceProfile   string
	rxName          string
	linkedRxs       []string
	spectrumProfile string
	maxPower        float64
	minPower        float64
	stepSize        float64
	frequency       int64
	sa              driver.SA
	sg              driver.SG
	currentStatus   [][]string
	statusMonitor   chan AttnMeasurementStatus
	deviations      []CorrectedDeviation
	stop            bool
}

// Internal Helpers

func (sg *SGPowerMeasurement) notify(msg string) {
	sg.statusMonitor <- AttnMeasurementStatus{Message: msg, Error: false}
}

func (sg *SGPowerMeasurement) finish(msg string, success bool) {
	sg.statusMonitor <- AttnMeasurementStatus{
		Message:   msg,
		Error:     !success,
		Completed: true,
	}
	close(sg.statusMonitor)
}

func (sg *SGPowerMeasurement) setError(msg string) {
	logger.Log.Error("SG Power Measurement Error", "error", msg)
	sg.finish(msg, false)
}

func (sg *SGPowerMeasurement) check(resp utils.CommandResponse, errMsg string) bool {
	if !resp.Success {
		sg.setError(fmt.Sprintf("%s: %s", errMsg, resp.ErrorMessage))
		return false
	}
	if sg.stop {
		sg.setError("Measurement aborted by user")
		return false
	}
	return true
}

// Public API

func (sg *SGPowerMeasurement) GetStatusMonitor() chan AttnMeasurementStatus {
	return sg.statusMonitor
}

func (sg *SGPowerMeasurement) Initialize(deviceProfile, rxName, spectrumProfile string,
	maxPower, minPower, stepSize float64) {
	sg.deviceProfile = deviceProfile
	sg.rxName = rxName
	sg.spectrumProfile = spectrumProfile
	sg.maxPower = maxPower
	sg.minPower = minPower
	sg.stepSize = stepSize
	sg.currentStatus = [][]string{{"Sl. No", "Set Power", "Measured Power", "Deviation"}}
	sg.statusMonitor = make(chan AttnMeasurementStatus, 20)
	sg.stop = false

	if !sg.loadDevices() {
		sg.setError("Unable to load devices")
		return
	}
	if !sg.loadDetails() {
		sg.setError("Unable to load details from database")
	}
}

func (sg *SGPowerMeasurement) Stop() {
	sg.stop = true
}

func (sg *SGPowerMeasurement) loadDevices() bool {
	saName, okSa := database.GetSAFromDeviceProfile(sg.deviceProfile)
	sgName, okSg := database.GetSGFromDeviceProfile(sg.deviceProfile)

	if !okSa || !okSg {
		return false
	}

	return sg.sa.LoadDevice(saName) && sg.sg.LoadDevice(sgName)
}

func (sg *SGPowerMeasurement) loadDetails() bool {
	freq, okFreq := database.GetRxFrequency(sg.rxName)
	rxs, okRxs := database.GetAllRxWithFrequency(freq)

	if !okFreq || !okRxs {
		return false
	}

	sg.frequency = freq
	sg.linkedRxs = rxs
	return true
}

func (sg *SGPowerMeasurement) StartMeasurement() {
	logger.Log.Info("Starting SG Power Measurement", "rxName", sg.rxName, "frequency", sg.frequency, "maxPower", sg.maxPower, "minPower", sg.minPower)
	sg.notify("SG Power Measurement Started")

	if !sg.check(sg.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !sg.check(sg.sg.SetFrequency(float64(sg.frequency)), "SG: set freq") {
		return
	}
	if !sg.check(sg.sg.SetModOff(), "SG: mod off") {
		return
	}
	if !sg.check(sg.sg.SetPower(0), "SG: power 0") {
		return
	}
	if !sg.check(sg.sg.SetRFOn(), "SG: RF on") {
		return
	}

	defer func() {
		sg.sa.SetAlignmentOn()
		sg.sa.SystemPreset()
		sg.sg.SetRFOff()
	}()

	spec, ok := database.GetSpectrumProfile(sg.spectrumProfile)
	if !ok {
		sg.setError("Spectrum profile not found in database")
		return
	}

	if !sg.check(sg.sa.SetSpectrum(spec.CenterFrequency, spec.Span, float64(spec.RBW), float64(spec.VBW)), "SA: set spectrum") {
		return
	}

	sg.notify("Measuring initial baseline power")
	if !sg.check(sg.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	resp := sg.sa.GetMaxMarkerValue()
	if !sg.check(resp, "SA: read baseline power") {
		return
	}
	baseline := resp.Result["MarkerY"].Value
	slNo := 0

	for power := sg.minPower; power <= sg.maxPower; power += sg.stepSize {
		if sg.stop {
			sg.setError("User Aborted")
			return
		}

		if !sg.check(sg.sg.SetPower(power), "SG: set power") {
			return
		}
		time.Sleep(time.Second)

		mResp := sg.sa.GetMaxMarkerValue()
		if !sg.check(mResp, "SA: read power") {
			return
		}

		measuredPower := mResp.Result["MarkerY"].Value - baseline
		diff := measuredPower - power
		slNo++

		measure := AttnMeasurementStatus{
			Message:       fmt.Sprintf("Completed Measurement for %.3f dBm", power),
			PlotDeviation: true,
		}
		measure.AddData(slNo, power, measuredPower, diff)
		logger.Log.Info("SG Power Measurement Point", "setPower", power, "measuredPower", measuredPower, "deviation", diff)
		sg.statusMonitor <- measure

		sg.currentStatus = append(sg.currentStatus, []string{
			strconv.Itoa(slNo),
			fmt.Sprintf("%.3f", power),
			fmt.Sprintf("%.3f", measuredPower),
			fmt.Sprintf("%.3f", diff),
		})
	}

	sg.notify("Saving Results")
	sg.saveResults()
	logger.Log.Info("Completed SG Power Measurement", "success", true)
	sg.finish("Measurement Completed", true)
}

func (sg *SGPowerMeasurement) saveResults() {
	var csv strings.Builder
	var reqs, measureds, diffs []float64

	for i, row := range sg.currentStatus {
		csv.WriteString(strings.Join(row, ","))
		csv.WriteString("\n")

		if i == 0 {
			continue
		}

		r, _ := strconv.ParseFloat(row[1], 64)
		m, _ := strconv.ParseFloat(row[2], 64)
		d, _ := strconv.ParseFloat(row[3], 64)
		reqs = append(reqs, r)
		measureds = append(measureds, m)
		diffs = append(diffs, d)
	}

	// Calculate Corrected Profile
	provider := utils.TSMAttnProvider{RequiredAttn: reqs, MeasuredAttn: measureds, Difference: diffs}
	corrected := utils.GetCorrectedProfile(provider, 0, sg.stepSize)

	sg.deviations = nil
	for i := range reqs {
		sg.deviations = append(sg.deviations, CorrectedDeviation{
			SetValue:           reqs[i],
			MeasuredDeviation:  diffs[i],
			CorrectedDeviation: corrected.GetDeviation(reqs[i]),
		})
	}

	// Write CSV for each linked RX
	for _, rx := range sg.linkedRxs {
		path := fmt.Sprintf("%s/.resources/sg-%s.csv", utils.Config.BaseFolder, rx)
		if err := os.WriteFile(path, []byte(csv.String()), 0644); err != nil {
			logger.Log.Error("Failed to write SG Power CSV", "rx", rx, "error", err)
		}
	}
}

func (sg *SGPowerMeasurement) GetCorrectedDeviations() []CorrectedDeviation {
	return sg.deviations
}
