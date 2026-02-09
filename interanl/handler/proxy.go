package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/interanl/victorialogs"
	log "github.com/sirupsen/logrus"
)

func ProxyRequest(c *gin.Context) {
	request := victorialogs.NewProxyRequest(c.Request.URL.Path, c.Request.URL.RawQuery)

	response, err := request.Query()
	if err != nil {
		log.Errorf("failed to make query: %v", err)
		c.JSON(500, gin.H{
			"message": "Failed to process proxy request",
		})
		return
	}

	c.Data(200, "application/json", response)
}
