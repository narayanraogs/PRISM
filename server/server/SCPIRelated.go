package server

import (
	"prismServer/driver"
	"time"

	"github.com/gin-gonic/gin"
)

func scpi(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var request SCPICommandRequest
	err = conn.ReadJSON(&request)
	if err != nil {
		conn.WriteJSON(SCPICommandResponse{
			OK:      false,
			Message: "Unable to read request",
		})
		return
	}
	var stop bool
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if string(msg) == "abort" {
				stop = true
			}
		}
	}()
	var device driver.ArbitraryDevice
	device = *driver.NewArbitraryDevice(request.Instrument, request.PortNo)
	for i, command := range request.Commands {
		if stop {
			conn.WriteJSON(SCPICommandResponse{
				OK:      false,
				Message: "Aborted",
			})
			return
		}
		if i < len(request.Delays) && request.Delays[i] > 0 {
			time.Sleep(time.Duration(request.Delays[i]) * time.Millisecond)
		}
		var response SCPICommandResponse
		response.Command = command
		response.Response, err = device.SendCommand(command, request.Read[i])
		if err != nil {
			response.OK = false
			response.Message = err.Error()
			conn.WriteJSON(response)
			return
		}
		response.OK = true
		response.Message = "Success"
		conn.WriteJSON(response)
	}
}
