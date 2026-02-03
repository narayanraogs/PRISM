package utilities

import (
	"fmt"
	"prismServer/driver"
	"strings"
)

type saStability struct {
	spans []singleSpan
	stop  bool
}

type singleSpan struct {
	powerPresent         bool
	frequencyPresent     bool
	powerDescription     string
	frequencyDescription string
	centerFrequency      float64
	span                 float64
	rbw                  float64
	vbw                  float64
	referenceLevel       float64
	autoReferece         bool
}

func (info *saStability) addMonitor(mode string, desc string, center float64, span float64, rbw float64, vbw float64, reference float64, auto bool) {
	var add = true
	var tbc = singleSpan{}
	for _, s := range info.spans {
		if s.centerFrequency == center && s.span == span && s.rbw == rbw && s.vbw == vbw && s.referenceLevel == reference {
			add = false
			tbc = s
			break
		}
	}
	if add {
		tbc.centerFrequency = center
		tbc.span = span
		tbc.rbw = rbw
		tbc.vbw = vbw
		tbc.referenceLevel = reference
		tbc.autoReferece = auto
	}
	if strings.EqualFold(mode, "Power") {
		tbc.powerPresent = true
		tbc.powerDescription = desc
	}
	if strings.EqualFold(mode, "Frequency") {
		tbc.frequencyPresent = true
		tbc.frequencyDescription = desc
	}
	info.spans = append(info.spans, tbc)
}

func startSAStability(saName string, info *saStability, stab *StabilityPlot) {
	var sa driver.SA
	ok := sa.LoadDevice(saName)
	if !ok {
		return
	}
	response := sa.SystemPreset()
	if !response.Success {
		//todo: add an empty point
	}
	response = sa.SetAlignmentOff()
	if !response.Success {
		//todo: add an empty point
	}

	for i, span := range info.spans {
		if span.autoReferece {
			sa.SetSpectrum(span.centerFrequency, span.span, span.rbw, span.vbw)
			response = sa.SetReferenceNominal()
			info.spans[i].referenceLevel = response.Result["ReferenceLevel"].Value
		}
	}

	for !info.stop {
		for _, span := range info.spans {
			sa.SetSpectrum(span.centerFrequency, span.span, span.rbw, span.vbw)
			sa.SetReferenceLevel(span.referenceLevel)
			sa.PeakSearch(false, 1)
			if span.frequencyPresent {
				response = sa.GetFrequencyInCounterMode(1)
				frequency := response.Result["Frequency"].Value
				fmt.Println(frequency)
				stab.addPoint(span.frequencyDescription, frequency)
			}
			if span.powerPresent {
				response = sa.GetMaxMarkerValue()
				power := response.Result["MarkerY"].Value
				stab.addPoint(span.powerDescription, power)
			}
		}
	}
}
