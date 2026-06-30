package server

import (
	"net/http"
	"prismServer/database"
	"prismServer/logger"
	"prismServer/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serverStatus(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("Error upgrading connection", "Error", err)
		return
	}
	defer conn.Close()

	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()
	var ok bool
	for range timer.C {

		var status ServerStatus
		status.SatelliteName = utils.Config.SatelliteName
		status.TestPhase, ok = database.GetSelectedTestPhase()
		if !ok {
			status.TestPhase = "Unknown"
		}
		status.IsOperationRunning = IsOperationRunning

		// Real System Stats
		if v, err := mem.VirtualMemory(); err == nil {
			status.MemoryUsed = v.UsedPercent
		}
		if c, err := cpu.Percent(time.Second, false); err == nil && len(c) > 0 {
			status.CPUUsed = c[0]
		}

		err = conn.WriteJSON(status)
		if err != nil {
			logger.Log.Error("Error writing JSON", "Error", err)
			return
		}
	}
}
