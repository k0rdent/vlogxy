package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/config"
)

func EmptyConfigMiddleware(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
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
