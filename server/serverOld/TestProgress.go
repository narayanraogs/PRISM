package server

import (
	"net/http"
	"prismServer/logger"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func testProgressHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	var req emptyRequest
	err = conn.ReadJSON(&req)
	if err != nil {
		logger.Log.Error("Error reading initial client registration:", err)
		return
	}

	gbl := sessions.getServer(req.ID)
	if gbl == nil || gbl.orchestrator == nil {
		_ = conn.WriteJSON(gin.H{"error": "Orchestrator not ready"})
		return
	}

	orchestrator := gbl.orchestrator
	var writeMutex sync.Mutex

	// Channel to receive messages from the websocket connection
	clientInputChan := make(chan getRequest)

	// Goroutine to read from the websocket and put messages on our channel
	go func() {
		defer close(clientInputChan)
		for {
			var inp getRequest
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

	// This single loop is the ONLY consumer of the CommChannel.
	for update := range orchestrator.CommChannel {
		// 1. Forward the update to the client (with a lock).
		writeMutex.Lock()
		err := conn.WriteJSON(update)
		writeMutex.Unlock()
		if err != nil {
			logger.Log.Error("Websocket write error:", err)
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
