package utils

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type PulseParameters struct {
	parameters map[string][]float64
	filename   string
	batchSize  int
}

func (pulse *PulseParameters) LoadFile(filename string, batchSize int) error {
	pulse.filename = filename
	pulse.batchSize = batchSize
	pulse.parameters = make(map[string][]float64)
	var nameMap = make(map[int]string)
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fileData := string(data)
	lines := strings.Split(fileData, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, ",")
		if i == 0 {
			for j, col := range cols {
				col = strings.ReplaceAll(col, "\"", "")
				pulse.parameters[col] = make([]float64, 0)
				nameMap[j] = col
			}
			continue
		}
		if batchSize != -1 {
			if i%batchSize == 0 || i%batchSize == 1 {
				continue
			}
		}
		for j, col := range cols {
			colName := nameMap[j]
			floatValue, err := strconv.ParseFloat(col, 64)
			if err != nil {
				return fmt.Errorf("cannot read file %d %d %s", i, j, err.Error())
			}
			if math.IsNaN(floatValue) {
				length := len(pulse.parameters[colName]) - 1
				if length >= 0 {
					floatValue = pulse.parameters[colName][length]
				}
			}
			pulse.parameters[colName] = append(pulse.parameters[colName], floatValue)
		}
	}
	return nil
}

func (pulse *PulseParameters) LoadPPMFile(filename string) error {
	pulse.parameters = make(map[string][]float64)
	var nameMap = make(map[int]string)
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fileData := string(data)
	lines := strings.Split(fileData, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, ",")
		if i == 0 {
			for j, col := range cols {
				col = strings.ReplaceAll(col, "\"", "")
				pulse.parameters[col] = make([]float64, 0)
				nameMap[j] = col
			}
			continue
		}

		for j, col := range cols {
			colName := nameMap[j]
			floatValue, err := strconv.ParseFloat(col, 64)
			if err != nil {
				return fmt.Errorf("cannot read file %d %d %s", i, j, err.Error())
			}
			if math.IsNaN(floatValue) {
				length := len(pulse.parameters[colName]) - 1
				if length >= 0 {
					floatValue = pulse.parameters[colName][length]
				}
			}
			pulse.parameters[colName] = append(pulse.parameters[colName], floatValue)
		}
	}
	return nil
}

func (pulse *PulseParameters) SegregateHighResolution(parameter string, hrspec float64, lrSpec float64) ([]string, error) {
	var counting = false
	var count = 0
	data, err := os.ReadFile(pulse.filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s", pulse.filename)
	}
	lines := strings.Split(string(data), "\n")
	firstLine := lines[0]
	var bucket1 = make([]string, 0) //hr
	var bucket2 = make([]string, 0) //lr
	bucket1 = append(bucket1, firstLine)
	bucket2 = append(bucket2, "No Of Pulses,Off Time")
	if len(lines) <= 2 {
		return nil, fmt.Errorf("not enough lines")
	}
	cols := strings.Split(firstLine, ",")
	parameter = "\"" + parameter + "\""
	colIndex := slices.Index(cols, parameter)
	if colIndex == -1 {
		return nil, fmt.Errorf("Column not present in data")
	}
	for i := 1; i < len(lines); i++ {
		if strings.EqualFold(strings.TrimSpace(lines[i]), "") {
			continue
		}
		cols := strings.Split(lines[i], ",")
		strValue := cols[colIndex]
		strValue = strings.ReplaceAll(strValue, "\"", "")
		currentValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			counting = false
			count = 0
			continue
		}
		if math.IsNaN(currentValue) {
			counting = false
			count = 0
			continue
		}
		currentValue = currentValue * 1e6
		diffLR := lrSpec - currentValue
		diffHR := hrspec - currentValue
		if math.Abs(diffHR) < math.Abs(diffLR) {
			counting = true
			count = count + 1
			bucket1 = append(bucket1, lines[i])
		} else {
			count = count + 1
			var row string
			if counting {
				row = fmt.Sprintf("%d,%.9f", count, currentValue/1e6)
				bucket2 = append(bucket2, row)
				counting = false
				count = 0
			}
		}
	}
	var hr string
	var lr string

	hr = strings.Join(bucket1, "\n")
	lr = strings.Join(bucket2, "\n")

	var tbr = make([]string, 0)
	tbr = append(tbr, hr, lr)
	return tbr, nil
}

func (pulse *PulseParameters) GetMeanValues(parameter string) ([]float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return nil, false
	}

	var sums = make([]float64, pulse.batchSize)
	for i, value := range data {
		index := i % pulse.batchSize
		sums[index] = sums[index] + value
	}
	noOfRepetitions := len(data) / pulse.batchSize

	var mean = make([]float64, 0)
	for i := 1; i < pulse.batchSize-1; i++ {
		temp := sums[i] / float64(noOfRepetitions)
		mean = append(mean, temp)
	}
	return mean, true
}

func (pulse *PulseParameters) GetValues(parameter string) ([]float64, bool) {
	data, ok := pulse.parameters[parameter]
	return data, ok
}

func (pulse *PulseParameters) GetMaxValues(parameter string) ([]float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return nil, false
	}

	var values = make([][]float64, pulse.batchSize)
	for i := 0; i < pulse.batchSize; i++ {
		values[i] = make([]float64, 0)
	}

	for i, value := range data {
		index := i % pulse.batchSize
		values[index] = append(values[index], value)
	}

	var max = make([]float64, 0)
	for i := 1; i < pulse.batchSize-1; i++ {
		temp := slices.Max(values[i])
		max = append(max, temp)
	}
	return max, true
}

func (pulse *PulseParameters) GetMinValues(parameter string) ([]float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return nil, false
	}

	var values = make([][]float64, pulse.batchSize)
	for i := 0; i < pulse.batchSize; i++ {
		values[i] = make([]float64, 0)
	}
	for i, value := range data {
		index := i % pulse.batchSize
		values[index] = append(values[index], value)
	}

	var min = make([]float64, 0)
	for i := 1; i < pulse.batchSize-1; i++ {
		temp := slices.Min(values[i])
		min = append(min, temp)
	}
	return min, true
}

func (pulse *PulseParameters) GetMeanSDValue(parameter string) (float64, float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return 0.0, 0.0, false
	}
	var length float64 = float64(len(data))

	var sum = 0.0
	for _, value := range data {
		sum = sum + value
	}
	var mean = sum / length

	var squares = 0.0
	for _, value := range data {
		diff := value - mean
		squares = squares + (diff * diff)
	}
	squares = math.Sqrt(squares)
	var sd = squares / length

	return mean, sd, true
}

func (pulse *PulseParameters) GetMaxValue(parameter string) (float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return 0.0, false
	}
	max := slices.Max(data)
	return max, true
}

func (pulse *PulseParameters) GetMinValue(parameter string) (float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return 0.0, false
	}
	min := slices.Min(data)
	return min, true
}

func (pulse *PulseParameters) GetFirstValue(parameter string) (float64, bool) {
	data, ok := pulse.parameters[parameter]
	if !ok {
		return 0.0, false
	}
	return data[0], true
}

func (pulse *PulseParameters) GetMiddleValues(decisionParam string, requiredParam string, division float64) []float64 {
	decisionValues := pulse.parameters[decisionParam]
	var singleTRMs = make([]PulseParameters, 0)
	referenceTime := time.UnixMilli(int64(decisionValues[0]))
	division = division * 1000
	requiredTimeDiff := time.Duration(division) * time.Millisecond
	var currentTRM PulseParameters
	currentTRM.parameters = make(map[string][]float64)
	for key := range pulse.parameters {
		currentTRM.parameters[key] = make([]float64, 0)
	}
	for i := 0; i < len(decisionValues); i++ {
		currentTime := time.UnixMilli(int64(decisionValues[i]))
		if currentTime.Sub(referenceTime) > requiredTimeDiff {
			singleTRMs = append(singleTRMs, currentTRM)
			currentTRM = PulseParameters{}
			currentTRM.parameters = make(map[string][]float64)
			for key := range pulse.parameters {
				currentTRM.parameters[key] = make([]float64, 0)
				currentTRM.parameters[key] = append(currentTRM.parameters[key], pulse.parameters[key][i])
			}
			referenceTime = time.UnixMilli(int64(decisionValues[i]))
		} else {
			for key := range pulse.parameters {
				currentTRM.parameters[key] = append(currentTRM.parameters[key], pulse.parameters[key][i])
			}
		}
	}
	singleTRMs = append(singleTRMs, currentTRM)

	var middleValues = make([]float64, 0)
	for _, trm := range singleTRMs {
		values := trm.parameters[requiredParam]
		mid := getMiddleValue(values)
		middleValues = append(middleValues, mid)
	}
	return middleValues
}

func getMiddleValue(array []float64) float64 {
	length := len(array)
	middle := int(length / 2)
	return array[middle]
}
