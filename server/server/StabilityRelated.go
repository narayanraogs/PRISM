package server

import (
	"net/http"
	"prismServer/database"

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
