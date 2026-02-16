package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/config"
	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/internal/router"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPort       = "8085"
	defaultConfigPath = "/etc/vlogxy/config.yaml"
	shutdownTimeout   = 30 * time.Second
	readHeaderTimeout = 10 * time.Second
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
	if configPath == "" {
		log.Fatalln("CONFIG_PATH environment variable is not set")
	}

	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	handlerInstance := handler.NewHandler(conf)

	// Setup router
	r := gin.Default()
	router.SetupRoutes(r, handlerInstance)

	// Start server
	log.Info("Starting server on :8085")
	if err := r.Run(":8085"); err != nil {
		log.Panicf("failed to start server: %v", err)
	}
}
