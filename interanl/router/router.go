package router

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/interanl/handler"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/select/logsql/field_values", handler.ProxyRequest)

	r.GET("/select/logsql/hits", handler.ProxyRequest)
	r.GET("/select/logsql/query", handler.ProxyRequest)
	r.GET("/select/logsql/stats_query", handler.ProxyRequest)
	r.GET("/select/logsql/stats_query_range", handler.ProxyRequest)

	r.GET("/reload", handler.ReloadConfig)

	r.GET("/health", handler.HealthCheck)
	r.GET("/readyz", handler.ReadyCheck)
}
