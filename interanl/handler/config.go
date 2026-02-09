package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/interanl/config"
	log "github.com/sirupsen/logrus"
)

func ReloadConfig(c *gin.Context) {
	if err := config.GlobalConfig.Reload(); err != nil {
		log.Errorf("failed to reload configuration: %v", err)
		c.JSON(500, gin.H{
			"message": "Failed to reload configuration",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Configuration reloaded successfully",
	})
}
