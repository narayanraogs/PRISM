package server

import (
	"net/http"
	"prismServer/remote"

	"github.com/gin-gonic/gin"
)

func remoteInfo(c *gin.Context) {
	var response remote.SoftwareResponse
	response = remote.GetInfo()

	c.IndentedJSON(http.StatusOK, response)
}
