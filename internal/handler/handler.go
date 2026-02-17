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
	config     interfaces.ConfigProvider
	httpClient interfaces.HTTPClient
	logsLimit  int
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(config interfaces.ConfigProvider, logsLimit int) *Handler {
	return &Handler{
		config:     config,
		logsLimit:  logsLimit,
		httpClient: http.DefaultClient,
	}
}

// ProxyStats handles /select/logsql/stats_query endpoint
func (h *Handler) ProxyStats(c *gin.Context) {
	query := NewStats()
	proxyInstance := proxy.NewProxy[StatsResponse](h.config.GetServerGroups(), h.httpClient, c)
	proxyInstance.ProxyRequest(query)
}

// ProxyStatsRange handles /select/logsql/stats_query_range endpoint
func (h *Handler) ProxyStatsRange(c *gin.Context) {
	query := NewStatsRange()
	proxyInstance := proxy.NewProxy[StatsRangeResponse](h.config.GetServerGroups(), h.httpClient, c)
	proxyInstance.ProxyRequest(query)
}

// ProxyHits handles /select/logsql/hits endpoint
func (h *Handler) ProxyHits(c *gin.Context) {
	query := NewHits()
	proxyInstance := proxy.NewProxy[Response](h.config.GetServerGroups(), h.httpClient, c)
	proxyInstance.ProxyRequest(query)
}

// ProxyFieldValues handles /select/logsql/field_values endpoint
func (h *Handler) ProxyFieldValues(c *gin.Context) {
	query := NewFieldValuesQuery()
	proxyInstance := proxy.NewProxy[FieldValuesResponse](h.config.GetServerGroups(), h.httpClient, c)
	proxyInstance.ProxyRequest(query)
}

// StreamQuery handles /select/logsql/query endpoint with streaming
func (h *Handler) StreamQuery(c *gin.Context) {
	query := NewStreamQuery()
	streamProxy := proxy.NewStreamProxy[[]byte](h.config.GetServerGroups(), h.httpClient, c, h.logsLimit)
	streamProxy.ProxyRequest(query)
}

// ReloadConfig handles /reload endpoint
func (h *Handler) ReloadConfig(c *gin.Context) {
	if err := h.config.Reload(); err != nil {
		log.Errorf("failed to reload configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to reload configuration",
			"error":   err.Error(),
		})
		return
	}

	log.Info("Configuration reloaded successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Configuration reloaded successfully",
	})
}

// HealthCheck handles /health endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
