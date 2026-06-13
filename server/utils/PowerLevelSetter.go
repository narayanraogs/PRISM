package utils

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"prismServer/logger"
)

type attenuationEntry struct {
	setValue      float64
	achievedValue float64
}

type AttenuationTable struct {
	values             []attenuationEntry
	fixedPadValue      float64
	fixedPadApplicable bool
}

func (t *AttenuationTable) AddAttn(set float64, achieved float64) {
	entry := attenuationEntry{
		setValue:      set,
		achievedValue: achieved,
	}
	i := sort.Search(len(t.values), func(i int) bool {
		return t.values[i].achievedValue >= achieved
	})

	t.values = append(t.values, attenuationEntry{})
	copy(t.values[i+1:], t.values[i:])
	t.values[i] = entry
}

func (t *AttenuationTable) SetFixedPadValue(fixed float64) {
	t.fixedPadValue = fixed * -1
	t.fixedPadApplicable = true
}

func (t *AttenuationTable) GetFixedPadValue() float64 {
	return t.fixedPadValue
}

func (t *AttenuationTable) GetValueToBeSet(required float64) (float64, bool) {
	fixedPadRequired := false
	if len(t.values) == 0 {
		logger.Log.Error("error when getting required value", "required", required)
		return 0.0, fixedPadRequired
	}
	if t.fixedPadApplicable {
		if required > t.fixedPadValue {
			fixedPadRequired = true
			required = required - t.fixedPadValue
		}
	}
	i := sort.Search(len(t.values), func(i int) bool {
		return t.values[i].achievedValue >= required
	})
	if i == len(t.values) {
		logger.Log.Debug("Required Attn", "required", required, "toBeSet", t.values[i-1].setValue)
		return t.values[i-1].setValue, fixedPadRequired
	}
	if i == 0 {
		logger.Log.Debug("Required Attn", "required", required, "toBeSet", t.values[0].setValue)
		return t.values[0].setValue, fixedPadRequired
	}
	val1 := math.Abs(t.values[i-1].achievedValue - required)
	val2 := math.Abs(t.values[i].achievedValue - required)
	if val1 < val2 {
		logger.Log.Debug("Required Attn", "required", required, "toBeSet", t.values[i-1].setValue)
		return t.values[i-1].setValue, fixedPadRequired
	}
	logger.Log.Debug("Required Attn", "required", required, "toBeSet", t.values[i].setValue)
	return t.values[i].setValue, fixedPadRequired
}

func GetAttenuationTable(file string) AttenuationTable {
	var t AttenuationTable
	var lines = strings.Split(file, "\n")
	for _, line := range lines {
		if strings.EqualFold("", strings.TrimSpace(line)) {
			continue
		}
		fields := strings.Split(line, ",")
		if strings.Contains(fields[1], "Fixed") {
			val, _ := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
			t.SetFixedPadValue(val)
			continue
		}
		set, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		acheived, _ := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
		t.AddAttn(set, acheived)
	}
	return t
}
