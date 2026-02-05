package utilities

import (
	"strings"
	"time"
)

type Stability struct {
	pms         map[string]*pmStability
	sas         map[string]*saStability
	ppms        map[string]*ppmStability
	dataChannel chan StabilityUpdate
}

type StabilityUpdate struct {
	Description string    `json:"Description"` // Must match the Description in the request
	Value       float64   `json:"Value"`
	Timestamp   time.Time `json:"Timestamp"`
}

func NewStability() *Stability {
	return &Stability{
		pms:         make(map[string]*pmStability),
		sas:         make(map[string]*saStability),
		ppms:        make(map[string]*ppmStability),
		dataChannel: make(chan StabilityUpdate, 100),
	}
}

func (stab *Stability) AddPM(description string, pmName string, channel string, frequency float64) {
	pm, ok := stab.pms[pmName]
	if !ok {
		pm = &pmStability{}
		stab.pms[pmName] = pm
	}
	if strings.EqualFold(channel, "Channel A") {
		pm.channelAPresent = true
		pm.channelADescription = description
		pm.channelAFrequency = frequency
	}
	if strings.EqualFold(channel, "Channel B") {
		pm.channelBPresent = true
		pm.channelBDescription = description
		pm.channelBFrequency = frequency
	}
}

func (stab *Stability) AddSA(description string, saName string, param string, centerFreq float64, span float64,
	rbw float64, vbw float64, refLevel float64, autoRef bool) {
	sa, ok := stab.sas[saName]
	if !ok {
		sa = &saStability{}
		stab.sas[saName] = sa
	}
	sa.addMonitor(param, description, centerFreq, span, rbw, vbw, refLevel, autoRef)
}

func (stab *Stability) AddPPM(description string, ppmName string, param string, channel string, config string, pulseProfile string) {

	ppm, ok := stab.ppms[ppmName]
	if !ok {
		ppm = &ppmStability{}
		stab.ppms[ppmName] = ppm
	}
	if strings.EqualFold(channel, "A") {
		if strings.EqualFold(param, "Peak Power") {
			ppm.peakPowerPresentChA = true
			ppm.peakPowerDescriptionChA = description
		}
		if strings.EqualFold(param, "Average Power") {
			ppm.avgPowerPresentChA = true
			ppm.avgPowerDescriptionChA = description
		}
		if strings.EqualFold(param, "Pulse Width") {
			ppm.pulseWidthPresentChA = true
			ppm.pulseWidthDescriptionChA = description
		}
		if strings.EqualFold(param, "Pulse Period") {
			ppm.pulsePeriodPresentChA = true
			ppm.pulsePeriodDescriptionChA = description
		}
		ppm.configurationChA = config
		ppm.pulseProfileChA = pulseProfile
	}
	if strings.EqualFold(channel, "B") {
		if strings.EqualFold(param, "Peak Power") {
			ppm.peakPowerPresentChB = true
			ppm.peakPowerDescriptionChB = description
		}
		if strings.EqualFold(param, "Average Power") {
			ppm.avgPowerPresentChB = true
			ppm.avgPowerDescriptionChB = description
		}
		if strings.EqualFold(param, "Pulse Width") {
			ppm.pulseWidthPresentChB = true
			ppm.pulseWidthDescriptionChB = description
		}
		if strings.EqualFold(param, "Pulse Period") {
			ppm.pulsePeriodPresentChB = true
			ppm.pulsePeriodDescriptionChB = description
		}
		ppm.configurationChB = config
		ppm.pulseProfileChB = pulseProfile
	}
}

func (stab *Stability) StopStability() {
	for _, pm := range stab.pms {
		pm.stop = true
	}
	for _, sa := range stab.sas {
		sa.stop = true
	}
	for _, ppm := range stab.ppms {
		ppm.stop = true
	}
}

func (stab *Stability) GetDataChannel() chan StabilityUpdate {
	return stab.dataChannel
}

func (stab *Stability) StartStability() {
	for name, pm := range stab.pms {
		go startPMStability(name, pm, stab.dataChannel)
	}
	for name, sa := range stab.sas {
		go startSAStability(name, sa, stab.dataChannel)
	}
	for name, ppm := range stab.ppms {
		go startPPMStability(name, ppm, stab.dataChannel)
	}
}
