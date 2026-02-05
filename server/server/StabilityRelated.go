package server

import (
	"net/http"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

func getStatbilityMetadata(c *gin.Context) {
	var stb StabilityMetadata
	stb.InstrumentTypes = []string{"SA", "PM", "PPM", "TM"}
	stb.Instruments = make(map[string][]string)
	stb.Parameters = make(map[string][]string)
	sa, ok := database.GetSAAndVSAList()
	if !ok {
		stb.OK = false
		stb.Message = "SA's not present in Database"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	stb.Instruments["SA"] = sa
	pm, ok := database.GetPMAndPPMList()
	if !ok {
		stb.OK = false
		stb.Message = "PM's not present in Database"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	stb.Instruments["PM"] = pm
	ppm, ok := database.GetPPMList()
	if ok {
		//PPM is optional
		stb.Instruments["PPM"] = ppm
	}
	stb.Instruments["TM"] = []string{"TM1", "TM2", "ANY"}
	stb.Parameters["SA"] = []string{"Power", "Frequency", "Trace"}
	stb.Parameters["PM"] = []string{"Channel A", "Channel B"}
	stb.Parameters["PPM"] = []string{"Peak Power", "Average Power", "Pulse Width", "Pulse Period"}
	stb.Parameters["TM"] = []string{"Processed", "Raw"}
	stb.PPMChannels = []string{"A", "B"}
	stb.Profiles = make([]SpectrumProfile, 0)
	sps, ok := database.GetAllSpectrumProfiles()
	if !ok {
		stb.OK = false
		stb.Message = "Cannot get Spectrum Profiles from Database"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	for _, profile := range sps {
		spec, ok := database.GetSpectrumProfile(profile)
		if !ok {
			continue
		}
		var prof SpectrumProfile
		prof.ProfileName = spec.Name
		prof.CenterFrequency = spec.CenterFrequency
		prof.Span = spec.Span
		prof.RBW = float64(spec.RBW)
		prof.VBW = float64(spec.VBW)
		stb.Profiles = append(stb.Profiles, prof)
	}
	stb.PLConfigs, ok = database.GetPLConfigurations()
	if !ok {
		stb.OK = false
		stb.Message = "PL Configurations not found"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	stb.PulseProfiles, ok = database.GetPulseProfileNames()
	if !ok {
		stb.OK = false
		stb.Message = "Pulse Profiles not found"
		c.IndentedJSON(http.StatusOK, stb)
		return
	}
	stb.OK = true
	stb.Message = "Success"
	c.IndentedJSON(http.StatusOK, stb)
}

/*
{
    "centerFrequency": float64, // In MHz
    "span":            float64, // In MHz
    "rbw":             float64, // In MHz
    "vbw":             float64, // In MHz
    "autoRef":         bool,
    "refLevel":        float64, // In dBm
    "profile":         string,  // e.g., "CW Capture"
}

{
    "frequency": float64,
    "unit":      string, // "MHz", "GHz", or "kHz"
}

{
    "plConfig":     string, // The Payload Config name
    "pulseProfile": string, // The Pulse Profile name
}

{
    "mnemonic": string, // e.g., "BAT_VOLT_1"
}

*/

func stability(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Error upgrading connection:", "error", err)
		return
	}
	defer conn.Close()

	var req StabilityRequest
	var resp StabilityResponse
	err = conn.ReadJSON(&req)
	if err != nil {
		logger.Log.Error("Error reading initial client registration:", "error", err)
		resp.OK = false
		resp.Message = "Invalid Request"
		conn.WriteJSON(resp)
		return
	}

	var stab utilities.Stability
	stab = *utilities.NewStability()
	inputChan := stab.GetDataChannel()
	for _, param := range req.Parameters {
		switch param.InstrumentType {
		case "SA":
			centerFreq := param.ExtraDetails["centerFrequency"].(float64)
			span := param.ExtraDetails["span"].(float64)
			rbw := param.ExtraDetails["rbw"].(float64)
			vbw := param.ExtraDetails["vbw"].(float64)
			autoRef := param.ExtraDetails["autoRefLevel"].(bool)
			refLevel := param.ExtraDetails["refLevel"].(float64)
			stab.AddSA(param.Description, param.Instrument, param.Parameter, centerFreq, span, rbw, vbw, refLevel, autoRef)
		case "PM":
			freq := param.ExtraDetails["frequencyHz"].(float64)
			stab.AddPM(param.Description, param.Instrument, param.Parameter, freq)
		case "PPM":
			channel := param.ExtraDetails["channel"].(string)
			config := param.ExtraDetails["plConfig"].(string)
			pulseProfile := param.ExtraDetails["pulseProfile"].(string)
			stab.AddPPM(param.Description, param.Instrument, param.Parameter, channel, config, pulseProfile)
		case "TM":
			go func(label string) {
				ticker := time.NewTicker(500 * time.Millisecond) // 2Hz sampling
				defer ticker.Stop()

				val := 25.0 // Starting point (e.g. 25 degC)
				seed := float64(time.Now().UnixNano() % 100)

				for {
					select {
					case <-ticker.C:
						// Simple random walk for simulation
						val += (seed/100.0 - 0.5) * 0.2
						seed = float64(int(seed+7) % 100) // Deterministic random-looking shift

						inputChan <- utilities.StabilityUpdate{
							Description: label,
							Value:       val,
							Timestamp:   time.Now(),
						}
					case <-c.Done(): // Clean up when connection closes
						return
					}
				}
			}(param.Description)
			//fmt.Println("To be implemented")
		}
	}

	go func() {
		for {
			var action StabilityAction
			err := conn.ReadJSON(&action)
			if err != nil || action.Action == "abort" {
				stab.StopStability()
				return
			}
		}
	}()

	go stab.StartStability()
	resp.Updates = make([]utilities.StabilityUpdate, 0)
	resp.OK = true
	resp.Message = "Success"
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case update := <-inputChan:
			resp.Updates = append(resp.Updates, update)
		case <-ticker.C:

			if len(resp.Updates) > 0 {
				err := conn.WriteJSON(resp)
				if err != nil {
					logger.Log.Error("Error writing to client:", "error", err)
					stab.StopStability()
					return
				}
				resp.Updates = make([]utilities.StabilityUpdate, 0)
			}
		case <-c.Done():
			stab.StopStability()
			return
		}
	}

}
