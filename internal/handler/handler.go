package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/internal/interfaces"
	"github.com/k0rdent/victorialogs-aggregator/internal/service"
	log "github.com/sirupsen/logrus"
)

// Handler holds dependencies for all HTTP handlers
type Handler struct {
	config       interfaces.ConfigProvider
	proxyService *service.ProxyService
}

// NewHandler creates a new handler instance with dependencies
func NewHandler(config interfaces.ConfigProvider, proxyService *service.ProxyService) *Handler {
	return &Handler{
		config:       config,
		proxyService: proxyService,
	}
}

// ProxyQuery handles /select/logsql/query endpoint
func (h *Handler) ProxyQuery(c *gin.Context) {
	query := NewQuery(c.Request.URL.Path, c.Request.URL.RawQuery)
	executeGenericQuery(c, h.proxyService, query)
}

// ProxyStats handles /select/logsql/stats_query endpoint
func (h *Handler) ProxyStats(c *gin.Context) {
	query := NewStats(c.Request.URL.Path, c.Request.URL.RawQuery)
	executeGenericQuery(c, h.proxyService, query)
}

// ProxyStatsRange handles /select/logsql/stats_query_range endpoint
func (h *Handler) ProxyStatsRange(c *gin.Context) {
	query := NewStatsRange(c.Request.URL.Path, c.Request.URL.RawQuery)
	executeGenericQuery(c, h.proxyService, query)
}

// ProxyHits handles /select/logsql/hits endpoint
func (h *Handler) ProxyHits(c *gin.Context) {
	query := NewHits(c.Request.URL.Path, c.Request.URL.RawQuery)
	executeGenericQuery(c, h.proxyService, query)
}

// ProxyFieldValues handles /select/logsql/field_values endpoint
func (h *Handler) ProxyFieldValues(c *gin.Context) {
	query := NewFieldValuesQuery(c.Request.URL.Path, c.Request.URL.RawQuery)
	executeGenericQuery(c, h.proxyService, query)
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

// executeGenericQuery is a unified helper function for executing queries
func executeGenericQuery[T any](c *gin.Context, proxyService *service.ProxyService, querier interfaces.ResponseAggregator[T]) {
	ctx := c.Request.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	response, err := service.Execute(ctx, proxyService, querier)
	if err != nil {
		log.Errorf("failed to execute query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to process query",
		})
		return
	}

	c.Data(http.StatusOK, "application/json", response)
}
