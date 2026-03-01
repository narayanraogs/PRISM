package tne

import (
	"fmt"
	"os"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/utils"
	"strconv"
	"strings"
	"time"
)

type GTxAttnMeasurement struct {
	deviceProfile   string
	rxName          string
	linkedRxs       []string
	spectrumProfile string
	component       string
	maxPower        float64
	minPower        float64
	stepSize        float64
	frequency       int64
	sa              driver.SA
	gtx             driver.GTX
	currentStatus   [][]string
	statusMonitor   chan AttnMeasurementStatus
	deviations      []CorrectedDeviation
	stop            bool
}

// Internal Helpers

func (gtx *GTxAttnMeasurement) notify(msg string) {
	gtx.statusMonitor <- AttnMeasurementStatus{Message: msg, Error: false}
}

func (gtx *GTxAttnMeasurement) finish(msg string, success bool) {
	gtx.statusMonitor <- AttnMeasurementStatus{
		Message:   msg,
		Error:     !success,
		Completed: true,
	}
	close(gtx.statusMonitor)
}

func (gtx *GTxAttnMeasurement) setError(msg string) {
	gtx.finish(msg, false)
}

func (gtx *GTxAttnMeasurement) check(resp utils.CommandResponse, errMsg string) bool {
	if !resp.Success {
		gtx.setError(fmt.Sprintf("%s: %s", errMsg, resp.ErrorMessage))
		return false
	}
	if gtx.stop {
		gtx.setError("Measurement aborted by user")
		return false
	}
	return true
}

// Public API

func (gtx *GTxAttnMeasurement) GetStatusMonitor() chan AttnMeasurementStatus {
	return gtx.statusMonitor
}

func (gtx *GTxAttnMeasurement) Initialize(deviceProfile, rxName, spectrumProfile, component string,
	maxPower, minPower, stepSize float64) {
	gtx.deviceProfile = deviceProfile
	gtx.rxName = rxName
	gtx.spectrumProfile = spectrumProfile
	gtx.component = component
	gtx.maxPower = maxPower
	gtx.minPower = minPower
	gtx.stepSize = stepSize
	gtx.currentStatus = [][]string{{"Sl. No", "Set Power", "Measured Power", "Deviation"}}
	gtx.statusMonitor = make(chan AttnMeasurementStatus, 20)
	gtx.stop = false

	if !gtx.loadDevices() {
		gtx.setError("Unable to load devices")
		return
	}
	if !gtx.loadDetails() {
		gtx.setError("Unable to load details from database")
	}
}

func (gtx *GTxAttnMeasurement) Stop() {
	gtx.stop = true
}

func (gtx *GTxAttnMeasurement) loadDevices() bool {
	saName, okSa := database.GetSAFromDeviceProfile(gtx.deviceProfile)
	gtxName, okGtx := database.GetGTxFromDeviceProfile(gtx.deviceProfile)

	if !okSa || !okGtx {
		return false
	}

	return gtx.sa.LoadDevice(saName) && gtx.gtx.LoadDevice(gtxName)
}

func (gtx *GTxAttnMeasurement) loadDetails() bool {
	freq, okFreq := database.GetRxFrequency(gtx.rxName)
	rxs, okRxs := database.GetAllRxWithFrequency(freq)

	if !okFreq || !okRxs {
		return false
	}

	gtx.frequency = freq
	gtx.linkedRxs = rxs
	return true
}

func (gtx *GTxAttnMeasurement) StartMeasurement() {
	gtx.notify("GTx Power Measurement Started")

	if !gtx.check(gtx.sa.SetAlignmentOff(), "SA: alignment off") {
		return
	}
	if !gtx.check(gtx.gtx.SetPower(gtx.component, 0), "GTx: power 0") {
		return
	}
	time.Sleep(200 * time.Millisecond)

	defer func() {
		gtx.sa.SetAlignmentOn()
		gtx.sa.SystemPreset()
		gtx.gtx.SetCarrierOff(gtx.component)
	}()

	if !gtx.check(gtx.gtx.SetModulationOff(gtx.component), "GTx: modulation off") {
		return
	}
	time.Sleep(200 * time.Millisecond)
	if !gtx.check(gtx.gtx.SetCarrierOn(gtx.component), "GTx: carrier on") {
		return
	}
	time.Sleep(200 * time.Millisecond)

	spec, ok := database.GetSpectrumProfile(gtx.spectrumProfile)
	if !ok {
		gtx.setError("Spectrum profile not found in database")
		return
	}

	if !gtx.check(gtx.sa.SetSpectrum(spec.CenterFrequency, spec.Span, float64(spec.RBW), float64(spec.VBW)), "SA: set spectrum") {
		return
	}
	if !gtx.check(gtx.sa.SetReferenceNominal(), "SA: find carrier") {
		return
	}
	time.Sleep(time.Second)

	resp := gtx.sa.GetMaxMarkerValue()
	if !gtx.check(resp, "SA: read initial power") {
		return
	}
	initialPower := resp.Result["MarkerY"].Value
	slNo := 0

	for power := gtx.minPower; power <= gtx.maxPower; power += gtx.stepSize {
		if gtx.stop {
			gtx.setError("User Aborted")
			return
		}

		if !gtx.check(gtx.gtx.SetPower(gtx.component, power), "GTx: set power") {
			return
		}
		time.Sleep(200 * time.Millisecond)

		mResp := gtx.sa.GetMaxMarkerValue()
		if !gtx.check(mResp, "SA: read power") {
			return
		}

		measuredPower := mResp.Result["MarkerY"].Value - initialPower
		diff := measuredPower - power
		slNo++

		measure := AttnMeasurementStatus{
			Message:       fmt.Sprintf("Completed Measurement for %.3f dBm", power),
			PlotDeviation: true,
		}
		measure.AddData(slNo, power, measuredPower, diff)
		gtx.statusMonitor <- measure

		gtx.currentStatus = append(gtx.currentStatus, []string{
			strconv.Itoa(slNo),
			fmt.Sprintf("%.3f", power),
			fmt.Sprintf("%.3f", measuredPower),
			fmt.Sprintf("%.3f", diff),
		})
	}

	gtx.notify("Saving Results")
	gtx.saveResults()
	gtx.finish("Measurement Completed", true)
}

func (gtx *GTxAttnMeasurement) saveResults() {
	var csv strings.Builder
	var reqs, measureds, diffs []float64

	for i, row := range gtx.currentStatus {
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

	// Calculate Corrected Deviations
	provider := utils.TSMAttnProvider{RequiredAttn: reqs, MeasuredAttn: measureds, Difference: diffs}
	corrected := utils.GetCorrectedProfile(provider, 0, gtx.stepSize)

	gtx.deviations = nil
	for i := range reqs {
		gtx.deviations = append(gtx.deviations, CorrectedDeviation{
			SetValue:           reqs[i],
			MeasuredDeviation:  diffs[i],
			CorrectedDeviation: corrected.GetDeviation(reqs[i]),
		})
	}

	// Write CSV for each linked RX
	for _, rx := range gtx.linkedRxs {
		path := fmt.Sprintf("%s/.resources/gtx-%s.csv", utils.Config.BaseFolder, rx)
		if err := os.WriteFile(path, []byte(csv.String()), 0644); err != nil {
			fmt.Printf("Warning: Failed to write CSV for %s: %v\n", rx, err)
		}
	}
}

func (gtx *GTxAttnMeasurement) GetCorrectedDeviations() []CorrectedDeviation {
	return gtx.deviations
}
