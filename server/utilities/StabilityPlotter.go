package utilities

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"os/exec"
	"prismServer/resultsDB"
	"prismServer/utils"
	"strings"
	"sync"
	"time"
)

var colors = []string{"dark-magenta", "brown", "blue", "dark-green", "red", "dark-pink", "web-green", "web-blue",
	"magenta", "goldenrod", "salmon", "orange", "navy", "dark-yellow", "cyan", "green"}

type StabilityPlot struct {
	order    []string
	points   map[string][]resultsDB.StabilityPoint
	dbPoints []resultsDB.StabilityPoint
	stop     []func()
	id       int64
	mutex    sync.Mutex
	once     sync.Once
	xstart   string
}

func (plot *StabilityPlot) CreateNew() {
	plot.order = make([]string, 0)
	plot.points = make(map[string][]resultsDB.StabilityPoint)
	plot.stop = make([]func(), 0)
	plot.dbPoints = make([]resultsDB.StabilityPoint, 0)
	id, _ := resultsDB.StartNewStability()
	plot.id = id
}

func (plot *StabilityPlot) StopStability() {
	for _, stop := range plot.stop {
		stop()
	}
}

func (plot *StabilityPlot) addParameter(paramName string) {
	fmt.Println("adding", paramName)
	plot.order = append(plot.order, paramName)
	plot.points[paramName] = make([]resultsDB.StabilityPoint, 0, 1000)
}

func (plot *StabilityPlot) addPoint(paramName string, value float64) {
	now := time.Now()
	tsInt := now.UnixMilli()
	tsStr := now.Format("02-01-2006_15:04:05.000")
	plot.once.Do(func() {
		plot.xstart = tsStr
	})
	point := resultsDB.StabilityPoint{
		TimeStampInt: tsInt,
		TimeStamp:    tsStr,
		Description:  paramName,
		Value:        value,
	}
	var points = make([]resultsDB.StabilityPoint, 0, 100)
	plot.mutex.Lock()
	plot.dbPoints = append(plot.dbPoints, point)
	plot.points[paramName] = append(plot.points[paramName], point)
	if len(plot.dbPoints) >= 10 {
		points = append(points, plot.dbPoints...)
		plot.dbPoints = make([]resultsDB.StabilityPoint, 0, 100)
	}
	plot.mutex.Unlock()
	if len(points) > 0 {
		go resultsDB.InsertPoints(plot.id, points)
	}
}

func (plot *StabilityPlot) getLayout() string {
	var builder strings.Builder
	builder.WriteString("set multiplot layout ")
	switch len(plot.order) {
	case 1:
		builder.WriteString("1,1")
	case 2:
		builder.WriteString("2,1")
	case 3:
		builder.WriteString("3,1")
	case 4:
		builder.WriteString("2,2")
	case 5:
		fallthrough
	case 6:
		builder.WriteString("3,2")
	case 7:
		fallthrough
	case 8:
		fallthrough
	case 9:
		builder.WriteString("3,3")
	case 10:
		fallthrough
	case 11:
		fallthrough
	case 12:
		builder.WriteString("4,3")
	default:
		builder.WriteString("4,4")
	}
	builder.WriteString(" rowsfirst downwards spacing 0,0\n")
	return builder.String()
}

func (plot *StabilityPlot) getRange(name string) (string, string, string) {
	points := plot.points[name]
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)

	for _, p := range points {
		if minValue > p.Value {
			minValue = p.Value
		}
		if maxValue < p.Value {
			maxValue = p.Value
		}
	}
	maxStr := fmt.Sprintf("%.2f", maxValue)
	minStr := fmt.Sprintf("%.2f", minValue)
	maxValue = maxValue + 0.5
	minValue = minValue - 0.5
	if math.IsInf(minValue, 1) || math.IsInf(maxValue, -1) {
		return "", "", ""
	}
	return fmt.Sprintf("set yrange [%.2f:%.2f]\n", minValue, maxValue), maxStr, minStr
}

func (plot *StabilityPlot) Plot() (string, bool) {
	output := utils.Config.BaseFolder + "/temp/temp.svg"
	var builder strings.Builder
	for _, name := range plot.order {
		builder.WriteString(getData(name, plot.points[name]))
	}
	builder.WriteString("set terminal svg dynamic enhanced size 1920,1080\n")
	builder.WriteString(fmt.Sprintf("set output \"%s\"\n", output))
	builder.WriteString("set datafile missing \"?\"\n")
	builder.WriteString(plot.getLayout())
	builder.WriteString("set xdata time\n")
	builder.WriteString("set timefmt \"%d-%m-%Y_%H:%M:%S\"\n")
	builder.WriteString("set format y \"%.2f\"\n")
	builder.WriteString("set format x \"%H:%M:%S\"\n")
	builder.WriteString(fmt.Sprintf("set xrange [\"%s\":]\n", plot.xstart))
	builder.WriteString("set grid x y\n")

	for i, name := range plot.order {
		rangeStr, maxStr, minStr := plot.getRange(name)
		builder.WriteString(rangeStr)
		tempName := strings.ReplaceAll(name, "_", "\\\\_")
		builder.WriteString(fmt.Sprintf("set ylabel \"%s\"\n", tempName))
		if len(plot.points[name]) <= 1 {
			builder.WriteString("set yrange [0:1]\n")
			builder.WriteString("set label 1 \"Waiting for Data\" at  graph 0.5,0.5 center\n")
			builder.WriteString("plot NaN notitle\n")
			builder.WriteString("unset yrange\n")
		} else {
			builder.WriteString("set xtics rotate by 270\n")
			builder.WriteString("set xlabel \"Time\"\n")
			builder.WriteString(fmt.Sprintf("set label 1 \"Max: %s Min: %s\" at graph 0.95, 0.95 right\n", maxStr, minStr))
			builder.WriteString(fmt.Sprintf("plot $%s using 2:3 with lines linetype rgb \"%s\" lw 2 notitle\n", name, colors[i]))
		}
	}
	builder.WriteString("unset multiplot\n")
	builder.WriteString("unset output\n\n")
	gnuPlotFile := utils.Config.BaseFolder + "/temp/gnuPlot.plot"
	err := os.WriteFile(gnuPlotFile, []byte(builder.String()), 0666)
	if err != nil {
		fmt.Println("Unable to write GNUPlot file")
		return "Unable to write GNUPlot file", false
	}
	cmd := exec.Command("gnuplot", utils.Config.BaseFolder+"/temp/gnuPlot.plot")
	data, err := cmd.CombinedOutput()
	if err != nil {
		return string(data), false
	}
	file, err := os.ReadFile(utils.Config.BaseFolder + "/temp/temp.svg")
	if err != nil {
		fmt.Println("Unable to read Plot File")
		return "Unable to read Plot File", false
	}
	codedSVG := base64.StdEncoding.EncodeToString(file)
	return codedSVG, true
}

func getData(header string, data []resultsDB.StabilityPoint) string {
	var builder strings.Builder
	builder.WriteString("$")
	builder.WriteString(header)
	builder.WriteString(" << EOD\n")
	for _, pt := range data {
		builder.WriteString(fmt.Sprintf("%d %s %.2f\n", pt.TimeStampInt, pt.TimeStamp, pt.Value))
	}
	builder.WriteString("EOD\n")
	return builder.String()
}
