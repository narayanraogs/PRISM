package server

import (
	"net/http"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/resultsDB"
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
	id, _ := resultsDB.StartNewStability()
	resp.Updates = make([]utilities.StabilityUpdate, 0)
	resp.OK = true
	resp.Message = "Success"
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	var points = make([]resultsDB.StabilityPoint, 0, 100)
outerFor:
	for {
		select {
		case update := <-inputChan:
			now := time.Now()
			tsInt := now.UnixMilli()
			tsStr := now.Format("02-01-2006_15:04:05.000")
			point := resultsDB.StabilityPoint{
				TimeStampInt: tsInt,
				TimeStamp:    tsStr,
				Description:  update.Description,
				Value:        update.Value,
			}
			points = append(points, point)
			if len(points) >= 100 {
				go resultsDB.InsertPoints(id, points)
				points = make([]resultsDB.StabilityPoint, 0, 100)
			}
			resp.Updates = append(resp.Updates, update)
		case <-ticker.C:
			if len(resp.Updates) > 0 {
				err := conn.WriteJSON(resp)
				if err != nil {
					logger.Log.Error("Error writing to client:", "error", err)
					stab.StopStability()
					break outerFor
				}
				resp.Updates = make([]utilities.StabilityUpdate, 0)
			}
		case <-c.Done():
			stab.StopStability()
			break outerFor
		}
	}
	if len(points) > 0 {
		go resultsDB.InsertPoints(id, points)
	}
}

func getStabilityReportsMetadata(c *gin.Context) {
	rows, err := resultsDB.GetStabilitySessions()
	var resp StabiltiyReportsMetadata
	if err != nil {
		resp.OK = false
		resp.Message = "Error getting stability reports metadata"
		c.IndentedJSON(http.StatusOK, resp)
		return
	}
	resp.ID = make([]int64, len(rows))
	resp.Date = make([]string, len(rows))
	resp.Time = make([]string, len(rows))
	resp.Parameters = make([][]string, len(rows))
	for i, row := range rows {
		resp.ID[i] = row.ID
		resp.Date[i] = row.Date
		resp.Time[i] = row.Time
		params, _ := resultsDB.GetStabilityParameters(row.ID)
		resp.Parameters[i] = params
	}
	resp.OK = true
	resp.Message = "Success"
	c.IndentedJSON(http.StatusOK, resp)
}

func getStabilityPoints(c *gin.Context) {
	var req StabilityPointsRequest
	var resp StabilityPointsResponse
	err := c.BindJSON(&req)
	if err != nil {
		resp.OK = false
		resp.Message = "Invalid Request"
		c.IndentedJSON(http.StatusOK, resp)
		return
	}
	rows, err := resultsDB.GetStabilityPoints(req.ID, req.Parameter)
	if err != nil {
		resp.OK = false
		resp.Message = "Error getting stability points"
		c.IndentedJSON(http.StatusOK, resp)
		return
	}
	resp.Points = rows
	resp.OK = true
	resp.Message = "Success"
	c.IndentedJSON(http.StatusOK, resp)
}
