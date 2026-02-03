package utils

import (
	"bytes"
	"image"
	"image/png"
	"time"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func GetTimeStampedFileName(prefix string) string {
	now := time.Now()
	formattedTime := now.Format("02-Jan-2006_15-04-05")
	filename := prefix + "_" + formattedTime
	return filename
}

func GetOldTimeStampedFileName(prefix string, t time.Time) string {
	formattedTime := t.Format("02-Jan-2006_15-04-05")
	filename := prefix + "_" + formattedTime
	return filename
}

func CropImage(x0 int, y0 int, x1 int, y1 int, data []byte) []byte {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	bounds := img.Bounds()
	maxX := bounds.Dx()
	maxY := bounds.Dy()
	rect := image.Rect(x0, y0, maxX-x1, maxY-y1)
	crop := img.(SubImager).SubImage(rect)
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, crop)
	if err != nil {
		return nil
	}
	return buffer.Bytes()
}
