package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/config"
)

func EmptyConfigMiddleware(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow health and reload requests to pass through even when config is empty
		if isHealthRequest(c) || isReloadRequest(c) {
			c.Next()
			return
		}

		if conf.IsEmpty() {
			c.JSON(503, gin.H{
				"error": "No server groups configured",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func isHealthRequest(c *gin.Context) bool {
	return c.Request.URL.Path == "/health"
}

func isReloadRequest(c *gin.Context) bool {
	return c.Request.URL.Path == "/reload"
}
