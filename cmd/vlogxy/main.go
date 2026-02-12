package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/config"
	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/internal/router"
	"github.com/k0rdent/vlogxy/internal/service"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Setup logger
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetLevel(log.DebugLevel)
	gin.DefaultWriter = logger.Writer()
	gin.DefaultErrorWriter = logger.WriterLevel(log.ErrorLevel)

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Panicf("config not loaded: %v", err)
	}

	// Initialize dependencies
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	proxyService := service.NewProxyService(conf, httpClient)
	handlerInstance := handler.NewHandler(conf, proxyService)

	// Setup router
	r := gin.Default()
	router.SetupRoutes(r, handlerInstance)

	// Start server
	log.Info("Starting server on :8085")
	if err := r.Run(":8085"); err != nil {
		log.Panicf("failed to start server: %v", err)
	}
}
