package utils

import "math"

func GetByteArrayForInt(value int) []byte {
	var tbr = make([]byte, 4)
	tbr[3] = byte(value & 0xFF)
	value = value >> 8
	tbr[2] = byte(value & 0xFF)
	value = value >> 8
	tbr[1] = byte(value & 0xFF)
	value = value >> 8
	tbr[0] = byte(value & 0xFF)
	return tbr
}

func GetIntForByte(array []byte) int {
	var tbr int
	tbr = int(array[0]) & 0xFF
	tbr = tbr << 8
	tbr = tbr | (int(array[1]) & 0xFF)
	tbr = tbr << 8
	tbr = tbr | (int(array[2]) & 0xFF)
	tbr = tbr << 8
	tbr = tbr | (int(array[3]) & 0xFF)
	return tbr
}

func GetByteArrayForFloat(value float64) []byte {
	tempUInt := math.Float32bits(float32(value))
	return GetByteArrayForInt(int(tempUInt))
}
