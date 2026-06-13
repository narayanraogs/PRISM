package utilities

import (
	"database/sql"
	"fmt"
	"prismServer/database"
	"prismServer/driver"
	"prismServer/logger"
	"time"
)

type ppmStability struct {
	resMode                   sql.NullString
	peakPowerPresentChA       bool
	peakPowerDescriptionChA   string
	avgPowerPresentChA        bool
	avgPowerDescriptionChA    string
	pulseWidthPresentChA      bool
	pulseWidthDescriptionChA  string
	pulsePeriodPresentChA     bool
	pulsePeriodDescriptionChA string
	configurationChA          string
	pulseProfileChA           string
	peakPowerPresentChB       bool
	peakPowerDescriptionChB   string
	avgPowerPresentChB        bool
	avgPowerDescriptionChB    string
	pulseWidthPresentChB      bool
	pulseWidthDescriptionChB  string
	pulsePeriodPresentChB     bool
	pulsePeriodDescriptionChB string
	configurationChB          string
	pulseProfileChB           string
	stop                      bool
}

func startPPMStability(id int64, ppmName string, info *ppmStability, dataChannel chan StabilityUpdate) {
	logger.Log.Info("Starting PPM Stability Measurement", "stabilityId", id, "ppmName", ppmName)
	var ppm driver.PPM
	ok := ppm.LoadDevice(ppmName)
	if !ok {
		logger.Log.Error("Failed to load PPM device for stability", "stabilityId", id, "ppmName", ppmName)
		return
	}
	ppm.PresetPPM()
	preset := true

	if info.peakPowerPresentChA || info.avgPowerPresentChA || info.pulseWidthPresentChA || info.pulsePeriodPresentChA {
		parameters, ok := database.GetPPMRelatedParameters(info.pulseProfileChA)
		if !ok {
			//todo: add an empty point
		}

		specs, ok := database.GetFullSpec(info.configurationChA)
		if !ok {
			//todo: add an empty point
		}

		fmt.Println("CenterFreq..............", specs.CenterFrequency)
		ppm.SetPulseParameters(specs.PulseWidth, specs.PulsePeriod, parameters.PPMTriggerLevel, parameters.PPMReferenceLevel,
			parameters.PPMYDivision, "A", preset)
		preset = false
		time.Sleep(1000 * time.Millisecond)
		ppm.SetChannelFrequency("A", specs.CenterFrequency)
	}

	if info.peakPowerPresentChB || info.avgPowerPresentChB || info.pulseWidthPresentChB || info.pulsePeriodPresentChB {
		parameters, ok := database.GetPPMRelatedParameters(info.pulseProfileChB)
		if !ok {
			//todo: add an empty point
		}

		specs, ok := database.GetFullSpec(info.configurationChB)
		if !ok {
			//todo: add an empty point
		}

		ppm.SetPulseParameters(specs.PulseWidth, specs.PulsePeriod, parameters.PPMTriggerLevel, parameters.PPMReferenceLevel,
			parameters.PPMYDivision, "B", preset)
		time.Sleep(1000 * time.Millisecond)
		ppm.SetChannelFrequency("B", specs.CenterFrequency)
	}

	pointsMeasured := 0
	for !info.stop {

		if info.peakPowerPresentChA {
			resp := ppm.GetPeakPower("A", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "A", "parameter", "Peak Power", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.peakPowerDescriptionChA,
				Value:       resp.Result["PulsePeakPower"].Value,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}
		if info.avgPowerPresentChA {
			resp := ppm.GetPeakPower("A", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "A", "parameter", "Average Power", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.avgPowerDescriptionChA,
				Value:       resp.Result["PulseAveragePower"].Value,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if info.pulseWidthPresentChA {
			resp := ppm.GetPulseWidth("A", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "A", "parameter", "Pulse Width", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.pulseWidthDescriptionChA,
				Value:       resp.Result["PulseOnTime"].Value * 1e6,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if info.pulsePeriodPresentChA {
			resp := ppm.GetPulsePeriod("A", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "A", "parameter", "Pulse Period", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.pulsePeriodDescriptionChA,
				Value:       resp.Result["PulsePeriod"].Value * 1e6,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if info.peakPowerPresentChB {
			resp := ppm.GetPeakPower("B", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "B", "parameter", "Peak Power", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.peakPowerDescriptionChB,
				Value:       resp.Result["PulsePeakPower"].Value,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}
		if info.avgPowerPresentChB {
			resp := ppm.GetPeakPower("B", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "B", "parameter", "Average Power", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.avgPowerDescriptionChB,
				Value:       resp.Result["PulseAveragePower"].Value,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if info.pulseWidthPresentChB {
			resp := ppm.GetPulseWidth("B", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "B", "parameter", "Pulse Width", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.pulseWidthDescriptionChB,
				Value:       resp.Result["PulseOnTime"].Value * 1e6,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if info.pulsePeriodPresentChB {
			resp := ppm.GetPulsePeriod("B", false)
			if !resp.Success {
				logger.Log.Error("PPM Stability Measurement Error", "stabilityId", id, "channel", "B", "parameter", "Pulse Period", "error", resp.ErrorMessage)
				continue
			}
			dataChannel <- StabilityUpdate{
				Description: info.pulsePeriodDescriptionChB,
				Value:       resp.Result["PulsePeriod"].Value * 1e6,
				Timestamp:   time.Now(),
			}
			pointsMeasured++
		}

		if pointsMeasured > 0 && pointsMeasured%1000 == 0 {
			logger.Log.Info("PPM Stability Heartbeat", "stabilityId", id, "pointsMeasured", pointsMeasured)
		}
	}
	logger.Log.Info("Completed PPM Stability Measurement", "stabilityId", id, "totalPointsMeasured", pointsMeasured)
}
