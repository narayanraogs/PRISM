package measurements

import (
	"fmt"
	"strings"
)

type resultValue struct {
	Value  interface{}
	Status string // "Success", "Error", or ""
}

func (rv resultValue) String(format string) string {
	var base string
	if s, ok := rv.Value.(string); ok {
		base = s
	} else {
		base = fmt.Sprintf(format, rv.Value)
	}

	if rv.Status != "" {
		return fmt.Sprintf("%s;%s", base, rv.Status)
	}
	return base
}

type txFrequencyResult struct {
	SpecificationMHz    float64
	MeasuredMHz         float64
	AllowedDeviationKHz float64
	DeviationKHz        resultValue
	DeviationPPM        resultValue
}

func (r *txFrequencyResult) ToHeader() []string {
	return []string{"Specification [MHz]", "Measured [MHz]", "Allowed Deviation [kHz]", "Deviation [kHz]", "Deviation [PPM]"}
}

func (r *txFrequencyResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.6f", r.SpecificationMHz),
		fmt.Sprintf("%.6f", r.MeasuredMHz),
		fmt.Sprintf("%.3f", r.AllowedDeviationKHz),
		r.DeviationKHz.String("%.3f"),
		r.DeviationPPM.String("%.2f"),
	}
}

type txPowerResult struct {
	SpecifiedDBm  float64
	MeasuredDBm   resultValue
	DeviationDB   resultValue
	PMReadingDBm  float64
	SAOBWPowerDBm float64
}

func (r *txPowerResult) ToHeader() []string {
	return []string{"Specified [dBm]", "Measured [dBm]", "Deviation [dB]", "PM Reading [dBm]", "SA OBW Power [dBm]"}
}

func (r *txPowerResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.2f", r.SpecifiedDBm),
		r.MeasuredDBm.String("%.2f"),
		r.DeviationDB.String("%.2f"),
		fmt.Sprintf("%.2f", r.PMReadingDBm),
		fmt.Sprintf("%.2f", r.SAOBWPowerDBm),
	}
}

type txHarmonicsResult struct {
	SpecifiedFreqMHz   float64
	MeasuredFreqMHz    float64
	SpecificationDBc   float64
	LevelDBc           resultValue
	NoiseFloorLevelDBm float64
	IsNilResult        bool
}

func (r *txHarmonicsResult) ToHeader() []string {
	return []string{"Frequency [MHz]", "Measured Frequency [MHz]", "Specification [dBc]", "Level [dBc]", "Noise Floor Level [dBm]"}
}

func (r *txHarmonicsResult) ToRow() []string {
	if r.IsNilResult {
		return []string{
			fmt.Sprintf("%06f", r.SpecifiedFreqMHz),
			fmt.Sprintf("%06f", r.SpecifiedFreqMHz),
			fmt.Sprintf("%.2f", r.SpecificationDBc),
			"NIL;Success",
			fmt.Sprintf("%.2f", r.NoiseFloorLevelDBm),
		}
	}

	return []string{
		fmt.Sprintf("%06f", r.SpecifiedFreqMHz),
		fmt.Sprintf("%06f", r.MeasuredFreqMHz),
		fmt.Sprintf("%.2f", r.SpecificationDBc),
		r.LevelDBc.String("%.2f"),
		fmt.Sprintf("%.2f", r.NoiseFloorLevelDBm),
	}
}

// txSpuriousResult models a single row of results for the Spurious measurement test.
type txSpuriousResult struct {
	FrequencyKHz float64
	LevelDBc     float64
	Spec         float64
}

func (r *txSpuriousResult) ToHeader() []string {
	return []string{"Frequency [kHz]", "Spurious Level [dBc]", "Spec [dBc]"}
}

func (r *txSpuriousResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.2f", r.FrequencyKHz),
		fmt.Sprintf("%.2f", r.LevelDBc),
		fmt.Sprintf("%.2f", r.Spec),
	}
}

type txModIndexResult struct {
	SubCarrier          string
	SubCarrierFrequency float64
	SpecifiedModIndex   float64
	MeasuredModIndex    resultValue
	Deviation           resultValue
}

func (r *txModIndexResult) ToHeader() []string {
	return []string{"SubCarrier Name", "SubCarrier Frequency [kHz]", "Mod Index Spec [rad]", "Measured Mod Index [rad]", "Deviation [%]"}
}

func (r *txModIndexResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%s", r.SubCarrier),
		fmt.Sprintf("%.2f", r.SubCarrierFrequency),
		fmt.Sprintf("%.3f", r.SpecifiedModIndex),
		r.MeasuredModIndex.String("%.3f"),
		r.Deviation.String("%.2f"),
	}
}

type tpModIndexResult struct {
	SubCarrier        string
	ToneFrequency     float64
	SpecifiedModIndex float64
	MeasuredModIndex  resultValue
	Deviation         resultValue
}

func (r *tpModIndexResult) ToHeader() []string {
	return []string{"SubCarrier Name", "SubCarrier Frequency [kHz]", "Mod Index Spec [rad]", "Measured Mod Index [rad]", "Deviation [%]"}
}

func (r *tpModIndexResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%s", r.SubCarrier),
		fmt.Sprintf("%.2f", r.ToneFrequency),
		fmt.Sprintf("%.3f", r.SpecifiedModIndex),
		r.MeasuredModIndex.String("%.3f"),
		r.Deviation.String("%.2f"),
	}
}

type txBandwidthResult struct {
	CentreFrequencyMHz float64
	MeasuredBW         resultValue
}

func (r *txBandwidthResult) ToHeader() []string {
	return []string{"Centre Frequency [MHz]", "Measured BW [kHz]"}
}

func (r *txBandwidthResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.6f", r.CentreFrequencyMHz),
		r.MeasuredBW.String("%.4f"),
	}
}

type rxLockDynamicResult struct {
	receiverPower float64
	actualPower   float64
	lockStatus    string
	agcValue      float64
}

func (r *rxLockDynamicResult) ToHeader() []string {
	return []string{"Receiver Power (dBm)", "Actual Power (dBm)", "Lock Status", "AGC"}
}

func (r *rxLockDynamicResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.2f", r.receiverPower),
		fmt.Sprintf("%.2f", r.actualPower),
		r.lockStatus,
		fmt.Sprintf("%.2f", r.agcValue),
	}
}

type rxCommandDynamicResult struct {
	receiverPower float64
	actualPower   float64
	lockStatus    string
	agcValue      float64
	cmdsSent      int
	cmdsExecuted  int
}

func (r *rxCommandDynamicResult) ToHeader() []string {
	return []string{"Receiver Power (dBm)", "Actual Power (dBm)", "Lock Status", "AGC", "Commands Sent", "Commands Executed"}
}

func (r *rxCommandDynamicResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.2f", r.receiverPower),
		fmt.Sprintf("%.2f", r.actualPower),
		r.lockStatus,
		fmt.Sprintf("%.2f", r.agcValue),
		fmt.Sprintf("%d", r.cmdsSent),
		fmt.Sprintf("%d", r.cmdsExecuted),
	}
}

type rxLoopStressResult struct {
	frequencyOffset float64
	loopStressValue float64
}

func (r *rxLoopStressResult) ToHeader() []string {
	return []string{"Frequency Offset (kHz)", "Loop Stress TM Value"}
}

func (r *rxLoopStressResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.2f", r.frequencyOffset),
		fmt.Sprintf("%.2f", r.loopStressValue),
	}
}

type rxCarrierAcquisitionResult struct {
	setFrequency         float64
	offsetFrequency      float64
	agcValue             float64
	noOfCommandsSent     int
	noOfCommandsExecuted int
	lockStatus           string
}

func (r *rxCarrierAcquisitionResult) ToHeader() []string {
	return []string{"Set Frequency (MHz)", "Offset Frequency (kHz)", "AGC", "Commands Sent", "Commands Executed", "Lock Status"}
}

func (r *rxCarrierAcquisitionResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.3f", r.setFrequency/1e6),
		fmt.Sprintf("%.2f", r.offsetFrequency/1e3),
		fmt.Sprintf("%.2f", r.agcValue),
		fmt.Sprintf("%d", r.noOfCommandsSent),
		fmt.Sprintf("%d", r.noOfCommandsExecuted),
		r.lockStatus,
	}
}

type rxRFUplinkResult struct {
	setPower               float64
	actualPower            float64
	lockStatus             bool
	agc                    float64
	modulation             string
	modIndexSpec           float64
	frequencyDeviationSpec float64
	measured               string
}

func (r *rxRFUplinkResult) ToHeader() []string {
	switch strings.ToUpper(r.modulation) {
	case "PM":
		return []string{"Rx I/P Power Required (dBm)", "Rx I/P Power Measured (dBm)", "Rx Lock Status",
			"AGC", "Spec TC MI (rad)", "Measured MI"}
	case "FM":
		return []string{"Rx I/P Power Required (dBm)", "Rx I/P Power Measured (dBm)", "Rx Lock Status",
			"AGC", "Spec Freq Dev (kHz)", "Measured Freq Dev"}
	default:
		return []string{"Rx I/P Power Required (dBm)", "Rx I/P Power Measured (dBm)", "Rx Lock Status",
			"AGC"}
	}
}

func (r *rxRFUplinkResult) ToRow() []string {
	var lockSts = "LOCK"
	if !r.lockStatus {
		lockSts = "UNLOCK"
	}
	switch strings.ToUpper(r.modulation) {
	case "PM":
		return []string{
			fmt.Sprintf("%.2f", r.setPower),
			fmt.Sprintf("%.2f", r.actualPower),
			lockSts,
			fmt.Sprintf("%.2f", r.agc),
			fmt.Sprintf("%.2f", r.modIndexSpec),
			r.measured,
		}
	case "FM":
		return []string{
			fmt.Sprintf("%.2f", r.setPower),
			fmt.Sprintf("%.2f", r.actualPower),
			lockSts,
			fmt.Sprintf("%.2f", r.agc),
			fmt.Sprintf("%.2f", r.frequencyDeviationSpec),
			r.measured,
		}
	default:
		return []string{
			fmt.Sprintf("%.2f", r.setPower),
			fmt.Sprintf("%.2f", r.actualPower),
			lockSts,
			fmt.Sprintf("%.2f", r.agc),
		}
	}
}

type txFrequencyDeviationResult struct {
	Freq1Spec float64
	Freq2spec float64
	Freq1Meas resultValue
	Freq2Meas resultValue
	Freq1Dev  float64
	Freq2Dev  float64
}

func (r *txFrequencyDeviationResult) ToHeader() []string {
	return []string{"Spec Frequency1 [MHz]", "Spec Frequency2 [MHz]", "Measured Frequency1 [MHz]", "Measured Frequency2 [MHz]",
		"Deviation for Frequency1 [kHz]", "Deviation for Frequency2 [kHz]"}
}

func (r *txFrequencyDeviationResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.6f", r.Freq1Spec),
		fmt.Sprintf("%.6f", r.Freq2spec),
		r.Freq1Meas.String("%.6f"),
		r.Freq2Meas.String("%.6f"),
		fmt.Sprintf("%.3f", r.Freq1Dev),
		fmt.Sprintf("%.3f", r.Freq2Dev),
	}
}

type pulsePPMResult struct {
	PulsePeakPower resultValue
	PulseAvgPower  resultValue
	PulseWidth     resultValue
	PulsePeriod    resultValue
	PulseOffTime   resultValue
	RiseTime       resultValue
	FallTime       resultValue
	DutyCycle      resultValue
}

func (r *pulsePPMResult) ToHeader() []string {
	return []string{"PeakPower", "AveragePower", "PulseWidth", "PulsePeriod", "PulseOffTime", "RiseTime", "FallTime", "DutyCycle"}
}

func (r *pulsePPMResult) ToRow() []string {
	return []string{
		r.PulsePeakPower.String("%.2f"),
		r.PulseAvgPower.String("%.2f"),
		r.PulseWidth.String("%.2f"),
		r.PulsePeriod.String("%.2f"),
		r.PulseOffTime.String("%.2f"),
		r.RiseTime.String("%.2f"),
		r.FallTime.String("%.2f"),
		r.DutyCycle.String("%.2f"),
	}
}

type pulseBandwidthResult struct {
	CentreFrequencyMHz float64
	SpecBW             float64
	MeasuredBW         resultValue
	Deviation          resultValue
}

func (r *pulseBandwidthResult) ToHeader() []string {
	return []string{"Centre Frequency [MHz]", "Specification BW [kHz]", "Measured BW [kHz]", "Deviation [kHz]"}
}

func (r *pulseBandwidthResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.2f", r.CentreFrequencyMHz),
		fmt.Sprintf("%.2f", r.SpecBW),
		r.MeasuredBW.String("%.2f"),
		r.Deviation.String("%.2f"),
	}
}

type pulseFrequencyResult struct {
	SpecificationMHz    float64
	MeasuredMHz         float64
	AllowedDeviationKHz float64
	DeviationKHz        resultValue
	DeviationPPM        resultValue
}

func (r *pulseFrequencyResult) ToHeader() []string {
	return []string{"Specification [MHz]", "Measured [MHz]", "Allowed Deviation [kHz]", "Deviation [kHz]", "Deviation [PPM]"}
}

func (r *pulseFrequencyResult) ToRow() []string {
	return []string{
		fmt.Sprintf("%.6f", r.SpecificationMHz),
		fmt.Sprintf("%.6f", r.MeasuredMHz),
		fmt.Sprintf("%.3f", r.AllowedDeviationKHz),
		r.DeviationKHz.String("%.4f"),
		r.DeviationPPM.String("%.4f"),
	}
}

type pulseUplink struct {
	SetFrequency  float64
	ExpectedPower float64
	MeasuredPower float64
}

func (p *pulseUplink) ToHeader() []string {
	return []string{"Frequency [MHz]", "Set Power [dBm]", "Measured Power [dBm]"}
}

func (p *pulseUplink) ToRow() []string {
	return []string{
		fmt.Sprintf("%.3f", p.SetFrequency),
		fmt.Sprintf("%.2f", p.ExpectedPower),
		fmt.Sprintf("%.2f", p.MeasuredPower),
	}
}
