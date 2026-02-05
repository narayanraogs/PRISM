package server

import (
	"fmt"
	"net/http"
	"prismServer/database"
	"prismServer/executeTest"
	"prismServer/logger"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func getAllTests(c *gin.Context) {
	tp, _ := database.GetSelectedTestPhase()
	var resp AllTests
	resp.Categories = make([]string, 0)
	resp.Configurations = make(map[string][]string)
	resp.Tests = make(map[string][]TestDescription)
	resp.Losses = make(map[string]string)
	configs, ok := database.GetAllConfigsForTests()
	var configNames = make([]string, 0)
	if !ok {
		resp.OK = false
		resp.Message = "Not able to get Details from Database"
		c.IndentedJSON(http.StatusOK, resp)
		return
	}

	for _, config := range configs {
		temp := strings.Split(config, ";")
		if slices.Index(resp.Categories, temp[0]) == -1 {
			resp.Categories = append(resp.Categories, temp[0])
			resp.Configurations[temp[0]] = make([]string, 0)
		}
		configNames = append(configNames, temp[1])
		resp.Configurations[temp[0]] = append(resp.Configurations[temp[0]], temp[1])
		if strings.EqualFold(temp[0], "rx") {
			_, sa, _, sc, ok := database.GetCurrentUplinkLoss(temp[1], tp)
			if ok {
				resp.Losses[temp[1]] = fmt.Sprintf("SA: %.2f, SC: %.2f", sa, sc)
			} else {
				resp.Losses[temp[1]] = ""
			}
		} else {
			_, sa, pm, ok := database.GetCurrentDownlinkLoss(temp[1], tp)
			if ok {
				resp.Losses[temp[1]] = fmt.Sprintf("SA: %.2f, PM: %.2f", sa, pm)
			} else {
				resp.Losses[temp[1]] = ""
			}
		}
	}

	for _, config := range configNames {
		resp.Tests[config] = make([]TestDescription, 0)
		tests, ok := database.GetTestsForConfig(config)
		if !ok {
			continue
		}
		for _, test := range tests {
			temp := strings.Split(test, ";")
			var t TestDescription
			if len(temp) == 2 {
				t.TestName = temp[0]
				t.TestCategory = temp[1]
			} else {
				t.TestName = temp[0]
				t.TestCategory = ""
			}
			resp.Tests[config] = append(resp.Tests[config], t)
		}
	}
	c.IndentedJSON(http.StatusOK, resp)
}

func startTests(c *gin.Context) {
	var req StartTestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, Ack{OK: false, Message: "Invalid Request"})
		return
	}

	// For now just acknowledge
	c.IndentedJSON(http.StatusOK, Ack{OK: true, Message: "Tests Started Successfully"})
}

func testProgress(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Error upgrading connection:", "error", err)
		return
	}
	defer conn.Close()

	var req StartTestsRequest
	err = conn.ReadJSON(&req)
	if err != nil {
		logger.Log.Error("Error reading initial client registration:", "error", err)
		return
	}

	var paramMap = make(map[string]interface{})
	var configs = make([]string, 0)
	var testNames = make([]string, 0)
	var testCategories = make([]string, 0)
	var remarks = make([]string, 0)
	for _, t := range req.Tests {
		configs = append(configs, t.Configuration)
		testNames = append(testNames, t.TestName)
		testCategories = append(testCategories, t.TestCategory)
		remarks = append(remarks, t.Remark)
		if len(t.ExtraParameters) > 0 {
			for _, p := range t.ExtraParameters {
				temp := strings.Split(p, ";")
				switch temp[0] {
				case "NominalPower":
					power, _ := strconv.ParseFloat(temp[1], 64)
					paramMap["NominalPower"] = power
				}
			}
		}
	}

	commChannel := make(chan executeTest.TestProgressResponse, 100)
	inputChannel := make(chan string, 10)

	orchestrator := executeTest.NewOrchestrator(configs, testNames, testCategories, remarks, paramMap, commChannel, inputChannel)
	var writeMutex sync.Mutex

	// Channel to receive messages from the websocket connection
	clientInputChan := make(chan ClientInput)

	// Goroutine to read from the websocket and put messages on our channel
	go func() {
		defer close(clientInputChan)
		for {
			var inp ClientInput
			err := conn.ReadJSON(&inp)
			if err != nil {
				// Client disconnected or sent bad data, close the channel to signal this
				return
			}
			clientInputChan <- inp
		}
	}()

	// Goroutine to run the tests
	go orchestrator.RunTests()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "abort" {
				orchestrator.Abort()
			}
		}
	}()

	// This single loop is the ONLY consumer of the CommChannel.
	for update := range orchestrator.CommChannel {
		// 1. Forward the update to the client (with a lock).
		writeMutex.Lock()
		err := conn.WriteJSON(update)
		writeMutex.Unlock()
		if err != nil {
			logger.Log.Error("Websocket write error:", "error", err)
			return // Exit if we can't write to the client.
		}

		// 2. If this update requires a response, handle the input logic.
		if update.UI.UserInput || update.UI.UserConfirmation {
			if update.UI.TimeoutSecs <= 0 {
				clientResponse, ok := <-clientInputChan
				if !ok {
					return // Client disconnected
				}
				orchestrator.InputChannel <- clientResponse.Parameters[0]
			} else {
				timeoutDuration := time.Duration(update.UI.TimeoutSecs) * time.Second
				select {
				case clientResponse, ok := <-clientInputChan:
					if !ok {
						return // Client disconnected
					}
					orchestrator.InputChannel <- clientResponse.Parameters[0]
				case <-time.After(timeoutDuration):
					orchestrator.InputChannel <- "TIMEOUT"
				}
			}
		}
	}
}
