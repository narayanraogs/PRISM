package utils

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Chart struct {
	title   string
	columns []string
	values  [][]float64
	max     []float64
	min     []float64
}

func (c *Chart) MakeChart(title string, columns []string, values [][]float64) {
	c.title = title
	c.columns = make([]string, 0)
	c.columns = append(c.columns, columns...)
	c.values = make([][]float64, 0)
	c.values = append(c.values, values...)
	c.max = make([]float64, 0)
	c.min = make([]float64, 0)
	for _, value := range c.values {
		var max = math.Inf(-1)
		var min = math.Inf(1)
		for _, v := range value {
			if math.IsNaN(v) {
				continue
			}
			if v > max {
				max = v
			}
			if v < min {
				min = v
			}
		}
		c.max = append(c.max, max+1.0)
		c.min = append(c.min, min-1.0)
	}
}

func (c *Chart) getPointsAsData() string {
	var builder strings.Builder
	builder.WriteString("$data << EOD\n")
	builder.WriteString("PulseNo ")
	for i := 0; i < len(c.columns); i++ {
		builder.WriteString(c.columns[i])
		builder.WriteString(" ")
	}
	builder.WriteString("\n")
	for i := 0; i < len(c.values[0]); i++ {
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString(" ")
		for j := 0; j < len(c.values); j++ {
			builder.WriteString(fmt.Sprintf("%.2f ", c.values[j][i]))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("EOD\n\n")
	return builder.String()
}

func (c *Chart) Save(fileName string) bool {
	var rowsCols string
	if len(c.columns) == 1 {
		rowsCols = "1,1"
	} else if len(c.columns) == 2 {
		rowsCols = "2,1"
	} else if len(c.columns) <= 4 {
		rowsCols = "2,2"
	} else {
		rowsCols = "3,2"
	}
	var colors = []string{"orange", "blue", "green", "medium-blue", "dark-green", "cyan"}
	var builder strings.Builder
	data := c.getPointsAsData()
	builder.WriteString(data)
	builder.WriteString("set terminal pngcairo enhanced size 1200,800 \n")
	builder.WriteString("set output \"")
	builder.WriteString(fileName + "\"\n")
	builder.WriteString("set datafile missing \"?\"\n")
	builder.WriteString("set multiplot layout ")
	builder.WriteString(rowsCols)
	builder.WriteString(" rowsfirst downwards spacing 0,0 title \"Pulse Parameters\"\n")
	builder.WriteString("set format x \"%.0f\"\n")
	builder.WriteString("set format y \"%.2f\"\n")
	builder.WriteString("set grid x y\n")

	for i := 0; i < len(c.columns); i++ {
		builder.WriteString("set yrange [")
		builder.WriteString(fmt.Sprintf("%.2f:%.2f]\n", c.min[i], c.max[i]))
		builder.WriteString("set ylabel \"")
		builder.WriteString(c.columns[i])
		builder.WriteString("\"\n")
		builder.WriteString("set xtics\n")
		if i < 4 {
			builder.WriteString("unset xlabel\n")
		} else {
			builder.WriteString("set xlabel \"Pulse Number\"\n")
		}
		builder.WriteString("plot $data using 1:($")
		builder.WriteString(strconv.Itoa(i + 2))
		builder.WriteString(") with linespoints linetype rgb \"")
		builder.WriteString(colors[i])
		builder.WriteString("\" lw 2 title \"\"\n")
	}
	builder.WriteString("unset multiplot\n")
	builder.WriteString("unset output\n\n")
	gnuPlotFile := Config.BaseFolder + "/temp/gnuPlot.plot"
	os.WriteFile(gnuPlotFile, []byte(builder.String()), 0666)

	cmd := exec.Command("gnuplot", gnuPlotFile)
	val, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error when executing gnuplot", string(val))
	}

	return true
}
