package router

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/handler"
)

// SetupRoutes configures all application routes with the provided handler
func SetupRoutes(r *gin.Engine, h *handler.Handler) {
	r.POST("/select/logsql/field_values", h.ProxyFieldValues)

	r.GET("/select/logsql/hits", h.ProxyHits)
	r.GET("/select/logsql/query", h.StreamQuery)
	r.GET("/select/logsql/stats_query", h.ProxyStats)
	r.GET("/select/logsql/stats_query_range", h.ProxyStatsRange)

	r.GET("/reload", h.ReloadConfig)

	r.GET("/health", h.HealthCheck)
}
