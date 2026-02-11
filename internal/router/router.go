package router

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/internal/handler"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/select/logsql/field_values", handler.ProxyFieldValues)

	r.GET("/select/logsql/hits", handler.ProxyHits)
	r.GET("/select/logsql/query", handler.ProxyQuery)
	r.GET("/select/logsql/stats_query", handler.ProxyStats)
	r.GET("/select/logsql/stats_query_range", handler.ProxyStatsRange)

	r.GET("/reload", handler.ReloadConfig)

	r.GET("/health", handler.HealthCheck)
	r.GET("/readyz", handler.ReadyCheck)
}
