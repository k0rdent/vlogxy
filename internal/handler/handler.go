package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/internal/proxy"
	log "github.com/sirupsen/logrus"
)

// Handler holds dependencies for all HTTP handlers
type Handler struct {
	config interfaces.ConfigProvider
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(config interfaces.ConfigProvider) *Handler {
	return &Handler{
		config: config,
	}
}

// ProxyQuery handles /select/logsql/query endpoint
func (h *Handler) ProxyQuery(c *gin.Context) {
	query := NewQuery()
	proxy := proxy.NewProxy[Logs](h.config.GetServerGroups(), http.DefaultClient, c)
	proxy.ProxyRequest(query)
}

// ProxyStats handles /select/logsql/stats_query endpoint
func (h *Handler) ProxyStats(c *gin.Context) {
	query := NewStats()
	proxy := proxy.NewProxy[StatsResponse](h.config.GetServerGroups(), http.DefaultClient, c)
	proxy.ProxyRequest(query)
}

// ProxyStatsRange handles /select/logsql/stats_query_range endpoint
func (h *Handler) ProxyStatsRange(c *gin.Context) {
	query := NewStatsRange()
	proxy := proxy.NewProxy[StatsRangeResponse](h.config.GetServerGroups(), http.DefaultClient, c)
	proxy.ProxyRequest(query)
}

// ProxyHits handles /select/logsql/hits endpoint
func (h *Handler) ProxyHits(c *gin.Context) {
	query := NewHits()
	proxy := proxy.NewProxy[Response](h.config.GetServerGroups(), http.DefaultClient, c)
	proxy.ProxyRequest(query)
}

// ProxyFieldValues handles /select/logsql/field_values endpoint
func (h *Handler) ProxyFieldValues(c *gin.Context) {
	query := NewFieldValuesQuery()
	proxy := proxy.NewProxy[FieldValuesResponse](h.config.GetServerGroups(), http.DefaultClient, c)
	proxy.ProxyRequest(query)
}

func (h *Handler) StreamQuery(c *gin.Context) {
	query := NewStreamQuery()
	streamProxy := proxy.NewStreamProxy[[]byte](h.config.GetServerGroups(), http.DefaultClient, c)
	streamProxy.ProxyRequest(query)
}

// ReloadConfig handles /reload endpoint
func (h *Handler) ReloadConfig(c *gin.Context) {
	if err := h.config.Reload(); err != nil {
		log.Errorf("failed to reload configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to reload configuration",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration reloaded successfully",
	})
}

// HealthCheck handles /health endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "I'm alive",
	})
}
