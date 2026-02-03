package utilities

import (
	"database/sql"
	"fmt"
	"prismServer/database"
	"prismServer/driver"
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

func startPPMStability(ppmName string, info *ppmStability, stab *StabilityPlot) {
	var ppm driver.PPM
	ok := ppm.LoadDevice(ppmName)
	if !ok {
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

	for !info.stop {

		if info.peakPowerPresentChA {
			resp := ppm.GetPeakPower("A", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.peakPowerDescriptionChA, resp.Result["PulsePeakPower"].Value)
		}
		if info.avgPowerPresentChA {
			resp := ppm.GetPeakPower("A", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.avgPowerDescriptionChA, resp.Result["PulseAveragePower"].Value)
		}

		if info.pulseWidthPresentChA {
			resp := ppm.GetPulseWidth("A", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.pulseWidthDescriptionChA, resp.Result["PulseOnTime"].Value*1e6)
		}

		if info.pulsePeriodPresentChA {
			resp := ppm.GetPulsePeriod("A", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.pulsePeriodDescriptionChA, resp.Result["PulsePeriod"].Value*1e6)
		}

		if info.peakPowerPresentChB {
			resp := ppm.GetPeakPower("B", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.peakPowerDescriptionChB, resp.Result["PulsePeakPower"].Value)
		}
		if info.avgPowerPresentChB {
			resp := ppm.GetPeakPower("B", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.avgPowerDescriptionChB, resp.Result["PulseAveragePower"].Value)
		}

		if info.pulseWidthPresentChB {
			resp := ppm.GetPulseWidth("B", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.pulseWidthDescriptionChB, resp.Result["PulseOnTime"].Value*1e6)
		}

		if info.pulsePeriodPresentChB {
			resp := ppm.GetPulsePeriod("B", false)
			if !resp.Success {
				continue
			}
			stab.addPoint(info.pulsePeriodDescriptionChB, resp.Result["PulsePeriod"].Value*1e6)
		}
	}
}
