package utils

import (
	"prismServer/logger"
	"strconv"
	"strings"
)

func GetFloatArray(values []string) []float64 {
	var tbr = make([]float64, 0)
	for _, value := range values {
		value = strings.TrimSpace(value)
		fVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			logger.Log.Error("Unable to convert to float")
		}
		tbr = append(tbr, fVal)
	}
	return tbr
}

func GetStringArray(values []float64) []string {
	var tbr = make([]string, 0)
	for _, value := range values {
		sVal := strconv.FormatFloat(value, 'G', 2, 64)
		tbr = append(tbr, sVal)
	}
	return tbr
}

func Transpose(array [][]string) [][]string {
	var retArray = make([][]string, len(array[0]))
	for i := 0; i < len(retArray); i++ {
		var line = make([]string, len(array))
		retArray[i] = line
	}
	for i := 0; i < len(array); i++ {
		for j := 0; j < len(array[i]); j++ {
			retArray[j][i] = array[i][j]
		}
	}
	return retArray
}

func GetRepeatedArray(value string, repeat int) []string {
	var tbr = make([]string, 0)
	for i := 0; i < repeat; i++ {
		tbr = append(tbr, value)
	}
	return tbr
}
