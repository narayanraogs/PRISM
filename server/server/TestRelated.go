package server

import (
	"prismServer/executeTest"
	"prismServer/logger"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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
				return
			}
			if len(inp.Parameters) > 0 && inp.Parameters[0] == "abort" {
				orchestrator.Abort()
				continue
			}
			clientInputChan <- inp
		}
	}()

	// Goroutine to run the tests
	go orchestrator.RunTests()

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
