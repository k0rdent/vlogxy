package handler

import "github.com/gin-gonic/gin"

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func ReadyCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ready",
	})
}
