package router

import (
	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/handler"
)

// SetupRoutes configures all application routes with the provided handler
func SetupRoutes(r *gin.Engine, h *handler.Handler) {
	r.POST("/select/logsql/field_values", h.ProxyFieldValues)
	r.POST("/select/logsql/field_names", h.ProxyFieldNames)
	r.GET("/select/logsql/field_values", h.ProxyFieldValues)
	r.GET("/select/logsql/field_names", h.ProxyFieldNames)

	r.GET("/select/logsql/hits", h.ProxyHits)
	r.GET("/select/logsql/query", h.StreamQuery)
	r.GET("/select/logsql/stats_query", h.ProxyStats)
	r.GET("/select/logsql/stats_query_range", h.ProxyStatsRange)
	r.GET("/select/logsql/facets", h.ProxyFacets)
	r.GET("/select/logsql/tail", h.StreamTail)
	r.GET("/select/logsql/stream_ids", h.StreamIds)
	r.GET("/select/logsql/streams", h.Streams)
	r.GET("/select/logsql/stream_field_names", h.StreamFieldNames)
	r.GET("/select/logsql/stream_field_values", h.StreamFieldValues)

	r.GET("/reload", h.ReloadConfig)
	r.GET("/health", h.HealthCheck)
}
