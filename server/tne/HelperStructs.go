package tne

// RTStatus is the universal status reporting object sent to the client
type RTStatus struct {
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
	Success   bool   `json:"success"`
	Error     bool   `json:"error"`
}

// Spectrum Settings used across various measurements
type spectrumSettings struct {
	span float64
	rbw  float64
	vbw  float64
}

// --- Cable Loss Shared Types ---

type CableLossRecord struct {
	SlNo         int                `json:"slNo"`
	CableName    string             `json:"cableName"`
	Length       float64            `json:"length"`
	Date         string             `json:"date"`
	Time         string             `json:"time"`
	Measurements []MeasurementPoint `json:"measurements"`
}

type MeasurementPoint struct {
	Frequency float64 `json:"frequency"` // GHz
	Loss      float64 `json:"loss"`      // dB
}

type cableLossMeasured struct {
	Frequency []string
	Measured  []string
}

// --- Up/Down Converter Shared Types ---

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

// --- Attenuation Shared Types ---

type CorrectedDeviation struct {
	SetValue           float64
	MeasuredDeviation  float64
	CorrectedDeviation float64
}

type AttnMeasurementStatus struct {
	SlNo          int
	SetAttn       float64
	MeasuredAttn  float64
	Deviation     float64
	HasData       bool
	Completed     bool
	Error         bool
	Message       string
	PlotDeviation bool
}

func (t *AttnMeasurementStatus) AddData(slNo int, setAttn float64, measured float64, deviation float64) {
	t.SlNo = slNo
	t.SetAttn = setAttn
	t.MeasuredAttn = measured
	t.Deviation = deviation
	t.HasData = true
}
