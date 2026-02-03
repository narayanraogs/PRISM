package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setValues(c *gin.Context) {
	var request setRequest
	var response ackResponse
	response.OK = false
	response.Message = ""

	if err := c.BindJSON(&request); err != nil {
		response.OK = false
		response.Message = "Bad Request"
		c.IndentedJSON(http.StatusOK, response)
		return
	}

	s := sessions.getServer(request.ID)
	if s == nil {
		response.OK = false
		response.Message = "Client Not registered"
		c.IndentedJSON(http.StatusOK, response)
		return
	}
	response = processSetRequest(s, request)
	c.IndentedJSON(http.StatusOK, response)
}

func processSetRequest(c *client, request setRequest) ackResponse {
	var ack ackResponse
	ack.OK = true
	ack.Message = ""
	for _, value := range request.Values {
		fmt.Println(value.Name, value.Values)
		c.global.SetParameters(value.Name, value.Values)
	}
	return ack
}
