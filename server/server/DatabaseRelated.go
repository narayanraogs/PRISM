package server

import (
	"net/http"
	"prismServer/database"
	"prismServer/utils"

	"github.com/gin-gonic/gin"
)

func getConfigsForUplink(c *gin.Context) {
	var req ConfigsForLossRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	cfgs, ok := database.GetAllConfigsForUplinkLoss(req.TestPhase)
	if !ok {
		c.IndentedJSON(http.StatusOK, ConfigsForLossResponse{OK: false, Message: "Failed to fetch configs"})
		return
	}
	c.IndentedJSON(http.StatusOK, ConfigsForLossResponse{
		Configs: cfgs,
		OK:      true,
	})
}

func getConfigsForDownlink(c *gin.Context) {
	var req ConfigsForLossRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	cfgs, ok := database.GetAllConfigsForDownlink(req.TestPhase)
	if !ok {
		c.IndentedJSON(http.StatusOK, ConfigsForLossResponse{OK: false, Message: "Failed to fetch configs"})
		return
	}
	c.IndentedJSON(http.StatusOK, ConfigsForLossResponse{
		Configs: cfgs,
		OK:      true,
	})
}

func getUplinkLossProfile(c *gin.Context) {
	var req LossProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	profile, ok := database.GetUplinkLossProfile(req.Config, req.TestPhase)
	if !ok {
		c.IndentedJSON(http.StatusOK, LossProfileResponse{OK: false, Message: "Failed to fetch profile"})
		return
	}
	c.IndentedJSON(http.StatusOK, LossProfileResponse{
		Profile: profile,
		OK:      true,
	})
}

func getDownlinkLossProfile(c *gin.Context) {
	var req LossProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	profile, ok := database.GetDownlinkLossProfile(req.Config, req.TestPhase)
	if !ok {
		c.IndentedJSON(http.StatusOK, LossProfileResponse{OK: false, Message: "Failed to fetch profile"})
		return
	}
	c.IndentedJSON(http.StatusOK, LossProfileResponse{
		Profile: profile,
		OK:      true,
	})
}

func saveUplinkLossProfile(c *gin.Context) {
	var req SaveLossProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	ok := database.UpdateUplinkLossProfile(req.Config, req.TestPhase, req.Profile)
	if !ok {
		c.IndentedJSON(http.StatusOK, Ack{OK: false, Message: "Failed to save profile"})
		return
	}
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: "Profile saved successfully"})
}

func saveDownlinkLossProfile(c *gin.Context) {
	var req SaveLossProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	ok := database.UpdateDownlinkLossProfile(req.Config, req.TestPhase, req.Profile)
	if !ok {
		c.IndentedJSON(http.StatusOK, Ack{OK: false, Message: "Failed to save profile"})
		return
	}
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: "Profile saved successfully"})
}

func selectTestPhase(c *gin.Context) {
	var req SelectTestPhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	msg, ok := database.SelectExisitingTestPhase(req.TestPhase)
	if !ok {
		c.IndentedJSON(http.StatusOK, Ack{OK: false, Message: msg})
		return
	}
	utils.SetTestPhase(req.TestPhase)
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: "Test phase selected"})
}

func addNewTestPhase(c *gin.Context) {
	var req AddNewTestPhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}
	msg, ok := database.InsertNewTestPhase(req.NewPhase, req.CopyFrom)
	if !ok {
		c.IndentedJSON(http.StatusOK, Ack{OK: false, Message: msg})
		return
	}
	utils.SetTestPhase(req.NewPhase)
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: "New test phase added"})
}
