package utilities

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func StartStability(values []string, plot *StabilityPlot) {
	plot.CreateNew()
	var pms = make(map[string]*pmStability)
	var sas = make(map[string]*saStability)
	var ppms = make(map[string]*ppmStability)
	for _, line := range values {
		params := strings.Split(line, ";")
		params[0] = strings.TrimSpace(params[0])
		params[0] = strings.ReplaceAll(params[0],"-","_")
		params[0] = strings.ReplaceAll(params[0]," ","_")
		plot.addParameter(params[0])
		if strings.EqualFold(params[1], "PM") {
			pm, ok := pms[params[2]]
			if !ok {
				pm = &pmStability{}
				pms[params[2]] = pm
			}
			if strings.EqualFold(params[3], "Channel A") {
				pm.channelAPresent = true
				pm.channelADescription = params[0]
				freqStr := strings.ReplaceAll(params[4], "Frequency: ", "")
				freq, _ := strconv.ParseFloat(freqStr, 64)
				pm.channelAFrequency = freq
			}
			if strings.EqualFold(params[3], "Channel B") {
				pm.channelBPresent = true
				pm.channelBDescription = params[0]
				freqStr := strings.ReplaceAll(params[4], "Frequency: ", "")
				freq, _ := strconv.ParseFloat(freqStr, 64)
				pm.channelBFrequency = freq
			}
		}

		if strings.EqualFold(params[1], "SA") {
			sa, ok := sas[params[2]]
			if !ok {
				sa = &saStability{}
				sas[params[2]] = sa
			}
			param := params[3]
			desc := params[0]
			details := strings.Split(params[4], ",")
			centerFreqStr := strings.ReplaceAll(details[0], "Freq:", "")
			centerFreq, _ := strconv.ParseFloat(centerFreqStr, 64)
			spanStr := strings.ReplaceAll(details[1], " Span: ", "")
			span, _ := strconv.ParseFloat(spanStr, 64)
			rbwStr := strings.ReplaceAll(details[2], " RBW: ", "")
			rbw, _ := strconv.ParseFloat(rbwStr, 64)
			vbwStr := strings.ReplaceAll(details[3], " VBW: ", "")
			vbw, _ := strconv.ParseFloat(vbwStr, 64)
			autoStr := strings.ReplaceAll(details[4], " Auto: ", "")
			auto := strings.EqualFold(autoStr, "true")
			refStr := strings.ReplaceAll(details[5], " Ref: ", "")
			ref, _ := strconv.ParseFloat(refStr, 64)
			sa.addMonitor(param, desc, centerFreq, span, rbw, vbw, ref, auto)
		}

		if strings.EqualFold(params[1], "PPM") {
			ppm, ok := ppms[params[2]]
			if !ok {
				ppm = &ppmStability{}
				ppms[params[2]] = ppm
			}
			details := strings.Split(params[4], ",")
			channel := strings.ReplaceAll(details[0], "Channel: ", "")
			config := strings.ReplaceAll(details[1], " Config: ", "")
			profile := strings.ReplaceAll(details[2], " Profile: ", "")
			if strings.EqualFold(channel, "A"){
				if strings.EqualFold(params[3], "Peak Power") {
					ppm.peakPowerPresentChA = true
					ppm.peakPowerDescriptionChA = params[0]
				}
				if strings.EqualFold(params[3], "Average Power") {
					ppm.avgPowerPresentChA = true
					ppm.avgPowerDescriptionChA = params[0]
				}
				if strings.EqualFold(params[3], "Pulse Width") {
					ppm.pulseWidthPresentChA = true
					ppm.pulseWidthDescriptionChA = params[0]
				}
				if strings.EqualFold(params[3], "Pulse Period") {
					ppm.pulsePeriodPresentChA = true
					ppm.pulsePeriodDescriptionChA = params[0]
				}
				ppm.configurationChA = config
				ppm.pulseProfileChA = profile
			}
			if strings.EqualFold(channel, "B"){
				if strings.EqualFold(params[3], "Peak Power") {
					ppm.peakPowerPresentChB = true
					ppm.peakPowerDescriptionChB = params[0]
				}
				if strings.EqualFold(params[3], "Average Power") {
					ppm.avgPowerPresentChB = true
					ppm.avgPowerDescriptionChB = params[0]
				}
				if strings.EqualFold(params[3], "Pulse Width") {
					ppm.pulseWidthPresentChB = true
					ppm.pulseWidthDescriptionChB = params[0]
				}
				if strings.EqualFold(params[3], "Pulse Period") {
					ppm.pulsePeriodPresentChB = true
					ppm.pulsePeriodDescriptionChB = params[0]
				}
				ppm.configurationChB = config
				ppm.pulseProfileChB = profile
			}
		}
	}

	for pm := range pms {
		stab := pms[pm]
		var stop = func() {
			stab.stop = true
		}
		plot.stop = append(plot.stop, stop)
		go startPMStability(pm, stab, plot)
	}

	for sa := range sas {
		stab := sas[sa]
		var stop = func() {
			stab.stop = true
		}
		plot.stop = append(plot.stop, stop)
		go startSAStability(sa, stab, plot)
	}
	for ppm := range ppms {
		stab := ppms[ppm]
		var stop = func() {
			stab.stop = true
		}
		plot.stop = append(plot.stop, stop)
		go startPPMStability(ppm, stab, plot)
	}

}

func StartMockStability() {
	var plot StabilityPlot
	for i := 0; i < 16; i++ {
		plot.addParameter("Param" + strconv.Itoa(i+1))
	}

	for i := 0; i < 100; i++ {
		for i := 0; i < 16; i++ {
			plot.addPoint("Param"+strconv.Itoa(i+1), -20+rand.Float64()*-1)
		}
		time.Sleep(50 * time.Millisecond)
		if i%10 == 0 {
			fmt.Println("Added point", i)
		}
	}
	_, ok := plot.Plot()
	if !ok {
		fmt.Println("File Cannot be created")
	} else {
		fmt.Println("File Created")
	}
}
