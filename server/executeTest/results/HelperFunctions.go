package results

import (
	"os"
	"strings"
)

func getColumnMap(ppm bool) map[string]string {
	var tbr = make(map[string]string)
	if ppm {
		tbr["AverageTxPowerPPM"] = "AverageTxPower"
		tbr["PulseAveragePowerPPM"] = "AveragePower"
		tbr["PeakPowerPPM"] = "PeakPower"
		tbr["DutyCycle"] = "DutyCycle"
		tbr["RiseTime"] = "RiseTime"
		tbr["FallTime"] = "FallTime"
		tbr["PulsePeriod"] = "PulsePeriod"
		tbr["PulseWidth"] = "PulseWidth"
		tbr["PulseSeparation"] = "PulseSeparation"
	} else {
		tbr["AverageTxPowerVSA"] = "Pulse_Mean_Level1"
		tbr["PulseAveragePowerVSA"] = "Pulse_On_Level1"
		tbr["PeakPowerVSA"] = "Pulse_Top_Level1"
		tbr["DutyCycle"] = "Pulse_Duty_Cycle1"
		tbr["RiseTime"] = "Pulse_Rise_Time1"
		tbr["FallTime"] = "Pulse_Fall_Time1"
		tbr["PulsePeriod"] = "Pulse_PRI1"
		tbr["PulseWidth"] = "Pulse_Width1"
		tbr["PulseSeparation"] = "Pulse_Off_Time1"

	}
	tbr["RepetitionRate"] = "Pulse_PRF1"
	tbr["FrequencyShift"] = "Pulse_Frequency_Deviation1"
	tbr["Phase"] = "Pulse_Phase_Mean1"
	tbr["Droop"] = "Pulse_DroopDb1"
	tbr["Overshoot"] = "Pulse_OvershootDb1"
	tbr["Ripple"] = "Pulse_RippleDb1"
	tbr["PRF"] = "Pulse_PRF1"
	tbr["RepilcaPeriod"] = "Pulse_PRI1"
	tbr["ReplicaRate"] = "Pulse_PRF1"
	tbr["ChirpBandwidth"] = "Pulse_Best-Fit_FM_Pk-Pk_Dev1"
	tbr["ChirpRate"] = "Pulse_Best-Fit_FM_Slope1"
	tbr["ChirpRateDeviation"] = "Pulse_Best-Fit_FM_INL1"
	tbr["Frequency"] = "Frequency"
	tbr["Bandwidth"] = "Bandwidth"
	return tbr
}

type details struct {
	filedata string
}

func (d *details) load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	d.filedata = string(data)
	return nil
}

func (d *details) getValue(paramName string) string {
	lines := strings.Split(d.filedata, "\n")
	for _, line := range lines {
		temp := strings.Split(line, ",")
		if strings.EqualFold(temp[0], paramName) {
			return strings.TrimSpace(temp[1])
		}
	}
	return ""
}

func (d *details) getTRMDetails() string {
	lines := strings.Split(d.filedata, "\n")
	for _, line := range lines {
		temp := strings.Split(line, ",")
			return strings.TrimSpace(temp[0])
		}
	
	return ""
}

