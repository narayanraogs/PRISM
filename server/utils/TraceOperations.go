package utils

import (
	_ "embed"
	"math"
	"sort"
	"strconv"
	"strings"
)

type trace struct {
	freq  float64
	power float64
}

type modIndex struct {
	powerIndBc float64
	modIndex   float64
	difference float64
}

//go:embed BesselChart.csv
var bessel string

func PowerToModIndexConverter(power float64) (float64, error) {
	if power > 0 {
		power = power * -1
	}
	var mi = make([]modIndex, 0)
	lines := strings.Split(bessel, "\n")
	for _, line := range lines {
		fileds := getFields(line)
		if len(fileds) != 3 {
			continue
		}
		var m modIndex
		m.powerIndBc, _ = strconv.ParseFloat(fileds[1], 64)
		m.modIndex, _ = strconv.ParseFloat(fileds[2], 64)
		m.difference = math.Abs(power - m.powerIndBc)
		mi = append(mi, m)
	}

	sort.Slice(mi, func(i int, j int) bool {
		return mi[i].difference < mi[j].difference
	})

	return mi[0].modIndex, nil
}

func MeasureFreqDeviation(fileData string, skipRows int) (float64, float64) {
	var traces = make([]trace, 0)
	traces = getTracePoints(fileData, skipRows)
	sort.Slice(traces, func(i int, j int) bool {
		return traces[i].power < traces[j].power
	})

	return traces[0].freq, traces[1].freq
}

func getTracePoints(fileData string, skipRows int) []trace {
	var traces = make([]trace, 0)

	var lines = strings.Split(fileData, "\n")
	for i, line := range lines {
		if i < skipRows {
			continue
		}
		rec := getFields(line)
		freq, err1 := strconv.ParseFloat(rec[0], 64)
		power, err2 := strconv.ParseFloat(rec[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		var t trace
		t.freq = freq
		t.power = power
		traces = append(traces, t)
	}
	return traces
}

func getFields(line string) []string {
	var values = make([]string, 0)
	var temp = ""
	var quote = false
	for _, char := range line {
		if char == '"' {
			quote = !quote
			continue
		}
		if char == ',' && !quote {
			values = append(values, temp)
			temp = ""
			continue
		}
		temp = temp + string(char)
	}
	values = append(values, temp)
	return values
}
