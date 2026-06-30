package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"prismServer/database"
	"prismServer/driver"
	"prismServer/executeTest"
	"prismServer/remote"
	"prismServer/utils"

	"github.com/gin-gonic/gin"
)

func info(c *gin.Context) {
	var resp remote.SoftwareResponse
	resp = remote.GetInfo()
	c.IndentedJSON(http.StatusOK, resp)
}

func remoteGetter(c *gin.Context) {
	var req remote.GetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, remote.GetResponse{
			Ack: remote.Acknowledgement{Status: false, Msg: "Bad Request"},
		})
		return
	}

	var values = make([]string, len(req.Params))
	for i, param := range req.Params {
		// Check in-memory remote parameter cache first
		if val, ok := remote.GetParam(param); ok {
			values[i] = val
			continue
		}
		if vals, ok := remote.GetParams(param); ok {
			values[i] = strings.Join(vals, ";")
			continue
		}

		// Fallback to current system global status
		switch param {
		case "TestPhase":
			values[i] = utils.GetTestPhase()
		default:
			values[i] = ""
		}
	}

	c.IndentedJSON(http.StatusOK, remote.GetResponse{
		Ack:    remote.Acknowledgement{Status: true},
		Params: req.Params,
		Values: values,
	})
}

func remoteSetter(c *gin.Context) {
	var req remote.SetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, remote.Acknowledgement{Status: false, Msg: "Bad Request"})
		return
	}

	if len(req.Params) != len(req.Values) {
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Params and Values length mismatch"})
		return
	}

	for i, param := range req.Params {
		val := req.Values[i]
		remote.SetParameter(param, val)

		// Synchronize specific configurations with local database/globals if needed
		if param == "TestPhase" {
			_, ok := database.SelectExisitingTestPhase(val)
			if ok {
				utils.SetTestPhase(val)
			}
		}
	}

	c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: true})
}

func runOrchestratorAsync(configs, testTypes, testCategories, remarks []string, extraParams map[string]interface{}) {
	commChannel := make(chan executeTest.TestProgressResponse, 100)
	inputChannel := make(chan string, 10)

	orchestrator := executeTest.NewOrchestrator(configs, testTypes, testCategories, remarks, extraParams, commChannel, inputChannel)
	go func() {
		defer UnlockOperation()

		// Drain the orchestrator progress responses asynchronously
		go func() {
			for update := range commChannel {
				broadcastToMonitors(update)
			}
		}()

		orchestrator.RunTests()
	}()
}

func remoteAction(c *gin.Context) {
	var req remote.ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, remote.Acknowledgement{Status: false, Msg: "Bad Request"})
		return
	}

	switch req.Type {
	case "RFUplink":
		config, _ := remote.GetParam("Configuration")
		if config == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing Configuration parameter"})
			return
		}
		powerStr, _ := remote.GetParam("Power")
		dopplerStr, _ := remote.GetParam("Doppler")
		fastUplinkStr, _ := remote.GetParam("FastUplink")

		var power float64
		if powerStr != "" {
			power, _ = strconv.ParseFloat(powerStr, 64)
		}

		isDoppler := strings.EqualFold(dopplerStr, "true")
		isFast := strings.EqualFold(fastUplinkStr, "true")

		var testCategory = "Full"
		if isFast && isDoppler {
			testCategory = "Fast-Doppler"
		} else if isFast {
			testCategory = "Fast"
		} else if isDoppler {
			testCategory = "Full-Doppler"
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}

		runOrchestratorAsync(
			[]string{config},
			[]string{"RFUplink"},
			[]string{testCategory},
			[]string{"Triggered via Remote API"},
			map[string]interface{}{"NominalPower": power},
		)
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: true, Msg: "RFUplink action started"})

	case "RemoveRFUplink":
		config, _ := remote.GetParam("Configuration")
		if config == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing Configuration parameter"})
			return
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}

		runOrchestratorAsync(
			[]string{config},
			[]string{"RFUplinkRemoval"},
			[]string{""},
			[]string{"Triggered via Remote API"},
			nil,
		)
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: true, Msg: "RemoveRFUplink action started"})

	case "ConductTest":
		config, _ := remote.GetParam("Configuration")
		testType, _ := remote.GetParam("TestType")
		testCategory, _ := remote.GetParam("TestCategory")
		remark, _ := remote.GetParam("Remark")

		if config == "" || testType == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing Configuration or TestType parameter"})
			return
		}

		var extraParams = make(map[string]interface{})
		if powerStr, ok := remote.GetParam("Power"); ok && powerStr != "" {
			if pVal, err := strconv.ParseFloat(powerStr, 64); err == nil {
				extraParams["NominalPower"] = pVal
			}
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}

		runOrchestratorAsync(
			[]string{config},
			[]string{testType},
			[]string{testCategory},
			[]string{remark},
			extraParams,
		)
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: true, Msg: fmt.Sprintf("%s action started", testType)})

	case "RouteTSM":
		tsmSelected, _ := remote.GetParam("TSM")
		tsmPath, _ := remote.GetParam("TSMPath")
		if tsmSelected == "" || tsmPath == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing TSM or TSMPath parameter"})
			return
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}
		defer UnlockOperation()

		var tsm driver.TSM
		if !tsm.LoadDevice(tsmSelected) {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Failed to load TSM device"})
			return
		}
		resp := tsm.SetDriverStatus(tsmPath)
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: resp.Success, Msg: resp.ErrorMessage})

	case "TSMAttnSet":
		tsmSelected, _ := remote.GetParam("TSM")
		attnNoStr, _ := remote.GetParam("AttnNo")
		attnValueStr, _ := remote.GetParam("AttnValue")

		if tsmSelected == "" || attnNoStr == "" || attnValueStr == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing TSM, AttnNo, or AttnValue parameter"})
			return
		}

		attnNo, err1 := strconv.Atoi(attnNoStr)
		attnValue, err2 := strconv.ParseFloat(attnValueStr, 64)
		if err1 != nil || err2 != nil {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Invalid AttnNo or AttnValue format"})
			return
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}
		defer UnlockOperation()

		var tsm driver.TSM
		if !tsm.LoadDevice(tsmSelected) {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Failed to load TSM device"})
			return
		}
		resp := tsm.SetAttn(attnNo, attnValue)
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: resp.Success, Msg: resp.ErrorMessage})

	case "SpectrumDump":
		saSelected, _ := remote.GetParam("SA")
		if saSelected == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing SA parameter"})
			return
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}
		defer UnlockOperation()

		var dev driver.SA
		if !dev.LoadDevice(saSelected) {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Unable to Load Device [Database Error]"})
			return
		}
		resp := dev.GetSpectrumDump()
		if resp.Success {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: true, Msg: resp.Result["SpectrumDump"].String})
		} else {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: resp.ErrorMessage})
		}

	case "Screenshot":
		vsaSelected, _ := remote.GetParam("VSA")
		profile, _ := remote.GetParam("Profile")
		if vsaSelected == "" || profile == "" {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Missing VSA or Profile parameter"})
			return
		}

		if !TryLockOperation() {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "System busy with another active session"})
			return
		}
		defer UnlockOperation()

		var dev driver.VSA
		if !dev.LoadDevice(vsaSelected) {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: "Unable to Load Device [Database Error]"})
			return
		}
		resp := dev.GetScreenshot(profile)
		if resp.Success {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: true, Msg: resp.Result["Screenshot"].String})
		} else {
			c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: resp.ErrorMessage})
		}

	default:
		c.IndentedJSON(http.StatusOK, remote.Acknowledgement{Status: false, Msg: fmt.Sprintf("Unknown action type: %s", req.Type)})
	}
}
