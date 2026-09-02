package server

import (
	"fmt"
	"net/http"
	"prismServer/logger"
	"prismServer/resultsDB"
	"prismServer/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

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
		fmt.Printf("%+v\n", param)
		switch param.InstrumentType {
		case "SA":
			centerFreq := param.ExtraDetails["centerFrequency"].(float64)
			span := param.ExtraDetails["span"].(float64)
			rbw := param.ExtraDetails["rbw"].(float64)
			vbw := param.ExtraDetails["vbw"].(float64)
			autoRef := param.ExtraDetails["autoRefLevel"].(bool)
			var refLevel = 0.0
			if !autoRef {
				refLevel = param.ExtraDetails["refLevel"].(float64)
			}
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

	if !TryLockOperation() {
		logger.Log.Warn("Cannot start stability test: operation lock busy")
		resp.OK = false
		resp.Message = "System Busy"
		conn.WriteJSON(resp)
		return
	}
	defer UnlockOperation()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go func() {
		defer stab.StopStability()
		for {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			var action StabilityAction
			err := conn.ReadJSON(&action)
			if err != nil || action.Action == "abort" {
				if err != nil {
					logger.Log.Warn("Stability connection error or timeout", "error", err)
				}
				return
			}
		}
	}()

	id, _ := resultsDB.StartNewStability()
	go stab.StartStability(id)
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
