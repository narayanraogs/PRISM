package utils

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetTracePlot(fileData string, noOfRowsToSkip int) (string, bool) {
	fileName := filepath.Join(Config.BaseFolder + "/temp/trace.png")

	var builder strings.Builder
	builder.WriteString("$data << EOD\nFrequency,Power\n")
	lines := strings.Split(fileData, "\n")
	lines = lines[noOfRowsToSkip:]
	for _, line := range lines {
		builder.WriteString(strings.TrimSpace(line))
		builder.WriteString("\n")
	}
	builder.WriteString("EOD\n")
	builder.WriteString("set terminal pngcairo enhanced size 1200,800\n")
	builder.WriteString("set datafile separator \",\"\n")
	builder.WriteString("set output \"")
	builder.WriteString(fileName + "\"\n")
	builder.WriteString("set datafile missing \"?\"\n")
	builder.WriteString("set format x \"%.0f\"\n")
	builder.WriteString("set format y \"%.2f\"\n")
	builder.WriteString("set grid x y\n")
	builder.WriteString("set ylabel \"Power\"\n")
	builder.WriteString("set xtics\n")
	builder.WriteString("plot $data using 1:($2) with lines linetype rgb \"blue\" lw 2 title \"\"\n")
	builder.WriteString("unset output\n")
	gnuPlotFile := filepath.Join(Config.BaseFolder + "/temp/trace.gnuPlot")
	err := os.WriteFile(gnuPlotFile, []byte(builder.String()), 0666)
	if err != nil {
		return "Unable to create plot file", false
	}
	cmd := exec.Command("gnuplot", gnuPlotFile)
	val, err := cmd.CombinedOutput()
	if err != nil {
		return string(val), false
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return "Unable to read Plot", false
	}

	return base64.StdEncoding.EncodeToString(data), true
}
