package driver

import (
	"math"
	"strings"
	"sync"
)

// Transmitter represents a signal source (real or simulated) in the lab.
type Transmitter struct {
	Frequency float64 // Hz
	Power     float64 // dBm
	IsOn      bool    // Status
	LossToSA  float64 // dB loss specifically to the SA port
	LossToPM  float64 // dB loss specifically to the PM port
}

// VirtualLab represents the central state for all simulated hardware in PRISM.
// It allows simulated drivers (SG, PM, SA, TSM, GTx) to interact as if they were physically connected.
type VirtualLab struct {
	mu sync.RWMutex

	// Signal Sources (Dynamic simulation points)
	SpacecraftTransmitters map[string]Transmitter // Key: "S-Band", "X-Band", etc.
	GroundTransmitters     map[string]Transmitter // Key: "SG-1", "GTx-2", etc.

	// TSM Matrix State
	TSMAttenuation map[int]float64 // Attenuator ID -> dB
	TSMPaths       map[int]bool    // Path/Driver ID -> On/Off
	CurrentPath    string          // Logical path string (e.g. "S-Band-Rx")

	// Environment
	NoiseFloor float64 // Default noise floor (e.g. -110 dBm)
}

var (
	instance *VirtualLab
	once     sync.Once
)

// GetVirtualLab returns the singleton instance of the simulation hub.
func GetVirtualLab() *VirtualLab {
	once.Do(func() {
		instance = &VirtualLab{
			SpacecraftTransmitters: map[string]Transmitter{
				"S-Band": {Frequency: 2.2e9, Power: 27.0, IsOn: true, LossToSA: 2.5, LossToPM: 3.1},
				"X-Band": {Frequency: 8.4e9, Power: 47.0, IsOn: true, LossToSA: 5.2, LossToPM: 6.4},
			},

			GroundTransmitters: map[string]Transmitter{
				"SG-1": {Frequency: 2e9, Power: -20.0, IsOn: false, LossToSA: 1.0, LossToPM: 1.0},
			},

			TSMAttenuation: make(map[int]float64),
			TSMPaths:       make(map[int]bool),

			NoiseFloor: -110.0,
		}
	})
	return instance
}

// AddSpacecraftTransmitter allows dynamic registration of any number of spacecraft transmitters with specific path losses.
func (v *VirtualLab) AddSpacecraftTransmitter(name string, freq, power, lossSA, lossPM float64, isOn bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.SpacecraftTransmitters[name] = Transmitter{
		Frequency: freq,
		Power:     power,
		IsOn:      isOn,
		LossToSA:  lossSA,
		LossToPM:  lossPM,
	}
}

// AddGroundTransmitter allows dynamic registration of any number of ground transmitters (SG/GTx).
func (v *VirtualLab) AddGroundTransmitter(name string, freq, power, lossSA, lossPM float64, isOn bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.GroundTransmitters[name] = Transmitter{
		Frequency: freq,
		Power:     power,
		IsOn:      isOn,
		LossToSA:  lossSA,
		LossToPM:  lossPM,
	}
}

// UpdateGroundTransmitter updates the state of an existing ground signal source.
func (v *VirtualLab) UpdateGroundTransmitter(name string, freq, power float64, on bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if tx, ok := v.GroundTransmitters[name]; ok {
		tx.Frequency = freq
		tx.Power = power
		tx.IsOn = on
		v.GroundTransmitters[name] = tx
	}
}

// UpdateTSM updates the matrix/attenuator state.
func (v *VirtualLab) UpdateTSM(attenuators map[int]float64, paths map[int]bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for id, val := range attenuators {
		v.TSMAttenuation[id] = val
	}
	for id, status := range paths {
		v.TSMPaths[id] = status
	}
}

// Measure calculates the resultant power at a specific frequency for a specific device type ("SA" or "PM"),
// accounting for active sources, registered device-specific losses, and TSM losses.
func (v *VirtualLab) Measure(frequency float64, deviceType string) float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	maxPower := v.NoiseFloor

	// 1. Check Spacecraft Transmitters (No TSM Attenuation)
	for _, tx := range v.SpacecraftTransmitters {
		if tx.IsOn && v.isFrequencyMatch(frequency, tx.Frequency) {
			pathPower := v.calculatePathPower(tx.Power, frequency, deviceType, tx, false)
			maxPower = math.Max(maxPower, pathPower)
		}
	}

	// 2. Check Ground Transmitters (TSM Attenuation Applicable)
	for _, tx := range v.GroundTransmitters {
		if tx.IsOn && v.isFrequencyMatch(frequency, tx.Frequency) {
			pathPower := v.calculatePathPower(tx.Power, frequency, deviceType, tx, true)
			maxPower = math.Max(maxPower, pathPower)
		}
	}

	return maxPower
}

// isFrequencyMatch checks if the tuned instrument is close enough to a source frequency.
func (v *VirtualLab) isFrequencyMatch(tuned, source float64) bool {
	return math.Abs(tuned-source) < 10e6
}

// calculatePathPower applies losses to a source signal based on device type, source type, and TSM state.
func (v *VirtualLab) calculatePathPower(sourcePower float64, frequency float64, deviceType string, tx Transmitter, isGroundSource bool) float64 {
	totalLoss := 0.0

	// 1. Apply registered device-specific loss from the transmitter (for SC)
	if !isGroundSource {
		if strings.EqualFold(deviceType, "SA") {
			totalLoss += tx.LossToSA
		} else if strings.EqualFold(deviceType, "PM") {
			totalLoss += tx.LossToPM
		}
	}

	// 2. Add dynamic TSM Attenuation (Only applicable for Ground Transmitters)
	if isGroundSource {
		for _, att := range v.TSMAttenuation {
			totalLoss += att
		}
	}

	// 3. TODO: Integrate frequency-dependent Cable Loss using database-driven business logic

	return sourcePower - totalLoss
}
