package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/interfaces"
	"github.com/k0rdent/vlogxy/internal/proxy"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultTailBufferSize = 5
	DefaultBufferSize     = 50
)

// Handler holds dependencies for all HTTP handlers
type Handler struct {
	config     interfaces.ConfigProvider
	httpClient interfaces.HTTPClient
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(config interfaces.ConfigProvider) *Handler {
	return &Handler{
		config:     config,
		httpClient: http.DefaultClient,
	}
}

// ProxyStats handles /select/logsql/stats_query endpoint
func (h *Handler) ProxyStats(c *gin.Context) {
	query := NewStats()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

// ProxyStatsRange handles /select/logsql/stats_query_range endpoint
func (h *Handler) ProxyStatsRange(c *gin.Context) {
	query := NewStatsRange()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

// ProxyHits handles /select/logsql/hits endpoint
func (h *Handler) ProxyHits(c *gin.Context) {
	query := NewHits()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

// ProxyFieldValues handles /select/logsql/field_values endpoint
func (h *Handler) ProxyFieldValues(c *gin.Context) {
	query := NewFieldValuesQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

func (h *Handler) ProxyFieldNames(c *gin.Context) {
	query := NewFieldNamesQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

func (h *Handler) ProxyFacets(c *gin.Context) {
	query := NewFacetsQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

func (h *Handler) StreamIds(c *gin.Context) {
	query := NewStreamIdsQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

func (h *Handler) Streams(c *gin.Context) {
	query := NewStreamsQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

func (h *Handler) StreamFieldNames(c *gin.Context) {
	query := NewStreamFieldNamesQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

func (h *Handler) StreamFieldValues(c *gin.Context) {
	query := NewStreamFieldValuesQuery()
	proxy.NewProxy(h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

// StreamQuery handles /select/logsql/query endpoint with streaming
func (h *Handler) StreamQuery(c *gin.Context) {
	query := NewStreamQuery(h.config.GetMaxLogsLimit(), DefaultBufferSize)
	proxy.NewStreamProxy[[]byte](h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
}

// StreamTail handles /select/logsql/tail endpoint with streaming
func (h *Handler) StreamTail(c *gin.Context) {
	query := NewTail(0, DefaultTailBufferSize)
	proxy.NewStreamProxy[[]byte](h.config.GetServerGroups(), h.httpClient, c, query).ProxyRequest()
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
