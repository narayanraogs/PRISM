package utilities

import (
	"prismServer/driver"
	"prismServer/logger"
	"strings"
	"time"
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

func startSAStability(id int64, saName string, info *saStability, dataChannel chan StabilityUpdate) {
	logger.Log.Info("Starting SA Stability Measurement", "stabilityId", id, "saName", saName)
	var sa driver.SA
	ok := sa.LoadDevice(saName)
	if !ok {
		logger.Log.Error("Failed to load SA device for stability", "stabilityId", id, "saName", saName)
		return
	}
	response := sa.SystemPreset()
	if !response.Success {
		logger.Log.Error("SA Stability Setup Error", "stabilityId", id, "step", "Preset", "error", response.ErrorMessage)
	}
	response = sa.SetAlignmentOff()
	if !response.Success {
		logger.Log.Error("SA Stability Setup Error", "stabilityId", id, "step", "Alignment Off", "error", response.ErrorMessage)
	}

	for i, span := range info.spans {
		if span.autoReferece {
			sa.SetSpectrum(span.centerFrequency, span.span, span.rbw, span.vbw)
			response = sa.SetReferenceNominal()
			info.spans[i].referenceLevel = response.Result["ReferenceLevel"].Value
		}
	}

	pointsMeasured := 0
	for !info.stop {
		for _, span := range info.spans {
			sa.SetSpectrum(span.centerFrequency, span.span, span.rbw, span.vbw)
			sa.SetReferenceLevel(span.referenceLevel)
			sa.PeakSearch(false, 1)
			if span.frequencyPresent {
				response = sa.GetFrequencyInCounterMode(1)
				if !response.Success {
					logger.Log.Error("SA Stability Measurement Error", "stabilityId", id, "parameter", "Frequency", "error", response.ErrorMessage)
					continue
				}
				frequency := response.Result["Frequency"].Value
				dataChannel <- StabilityUpdate{
					Description: span.frequencyDescription,
					Value:       frequency,
					Timestamp:   time.Now(),
				}
				pointsMeasured++
			}
			if span.powerPresent {
				response = sa.GetMaxMarkerValue()
				if !response.Success {
					logger.Log.Error("SA Stability Measurement Error", "stabilityId", id, "parameter", "Power", "error", response.ErrorMessage)
					continue
				}
				power := response.Result["MarkerY"].Value
				dataChannel <- StabilityUpdate{
					Description: span.powerDescription,
					Value:       power,
					Timestamp:   time.Now(),
				}
				pointsMeasured++
			}
		}

		if pointsMeasured > 0 && pointsMeasured%1000 == 0 {
			logger.Log.Info("SA Stability Heartbeat", "stabilityId", id, "pointsMeasured", pointsMeasured)
		}
	}
	logger.Log.Info("Completed SA Stability Measurement", "stabilityId", id, "totalPointsMeasured", pointsMeasured)
}
