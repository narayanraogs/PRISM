package tne

import (
	"fmt"
	"prismServer/reports"
)

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
	SpectrumDump              []string
}

type GainResults struct {
	SetPower    []float64
	OutputPower []float64
	Gain        []float64
	AverageGain float64
}

func (res *GainResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Sl.No", "Power Set", "Power Measured", "Gain"}
	for i := 0; i < len(res.Gain); i++ {
		row := make([]reports.DataCell, 0)
		row = append(row, reports.GetDataCell(fmt.Sprintf("%d", i+1)))
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.SetPower[i])))
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.OutputPower[i])))
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.Gain[i])))
		result.Data = append(result.Data, row)
	}
	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell(""))
	row = append(row, reports.GetDataCell("Avg"))
	row = append(row, reports.GetDataCell(""))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.AverageGain)))
	result.Data = append(result.Data, row)
	return result
}

type FrequencyResults struct {
	ExpectedFrequency float64
	MeasuredFrequency float64
	Deviation         float64
}

func (res *FrequencyResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Expected Frequency", "Measured Frequency", "Deviation"}

	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.ExpectedFrequency)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.MeasuredFrequency)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.Deviation)))
	result.Data = append(result.Data, row)

	return result
}

type HarmonicResults struct {
	HarmonicNo        []int
	HarmonicFrequency []string
	CarrierLevel      []string
	NoiseFloor        []float64
}

func (res *HarmonicResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Harmonic No", "Harmonic Frequency", "Carrier Level", "Noise Floor"}
	for i := 0; i < len(res.HarmonicNo); i++ {
		row := make([]reports.DataCell, 0)
		row = append(row, reports.GetDataCell(fmt.Sprintf("%d", res.HarmonicNo[i])))
		row = append(row, reports.GetDataCell(res.HarmonicFrequency[i]))
		row = append(row, reports.GetDataCell(res.CarrierLevel[i]))
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.NoiseFloor[i])))
		result.Data = append(result.Data, row)
	}
	return result
}

type SpuriousResults struct {
	Frequency        []string
	MeasuredPowerdBm []string
	SpuriousLeveldBC []string
}

func (res *SpuriousResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Sl. No", "Frequency", "Measured Power (dBm)", "Spurious Level (dBC)"}
	for i := 0; i < len(res.Frequency); i++ {
		row := make([]reports.DataCell, 0)
		row = append(row, reports.GetDataCell(fmt.Sprintf("%d", i+1)))
		row = append(row, reports.GetDataCell(res.Frequency[i]))
		row = append(row, reports.GetDataCell(res.MeasuredPowerdBm[i]))
		row = append(row, reports.GetDataCell(res.SpuriousLeveldBC[i]))
		result.Data = append(result.Data, row)
	}
	return result
}

type PowerOrLeakageResults struct {
	Frequency float64
	Power     float64
}

func (res *PowerOrLeakageResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Frequency", "Measured Power (dBm)"}

	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.Frequency)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.Power)))
	result.Data = append(result.Data, row)

	return result
}

type PhaseNoiseResults struct {
	Frequency  []float64
	PhaseNoise []float64
}

func (res *PhaseNoiseResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Frequency", "Phase Noise (dB/Hz)"}

	for i := 0; i < len(res.Frequency); i++ {
		row := make([]reports.DataCell, 0)
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.Frequency[i])))
		row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.PhaseNoise[i])))
		result.Data = append(result.Data, row)
	}

	return result
}

type PowerMatchingResults struct {
	InternalLOPowerMeasured float64
	ExternalLOPowerMeasured float64
	ExternalSGPowerSet      float64
}

func (res *PowerMatchingResults) getResultTable() reports.Result {
	var result reports.Result
	result.Data = make([][]reports.DataCell, 0)
	result.Header = []string{"Internal LO Power Measured", "External LO Power Measured", "SG Power Set"}

	row := make([]reports.DataCell, 0)
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.InternalLOPowerMeasured)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.ExternalLOPowerMeasured)))
	row = append(row, reports.GetDataCell(fmt.Sprintf("%.2f", res.ExternalSGPowerSet)))
	result.Data = append(result.Data, row)

	return result
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
