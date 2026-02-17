package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/vlogxy/internal/config"
	"github.com/k0rdent/vlogxy/internal/handler"
	"github.com/k0rdent/vlogxy/internal/middleware"
	"github.com/k0rdent/vlogxy/internal/router"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPort      = "8085"
	defaultLogsLimit = 1000
)

func main() {
	port := flag.String("port", defaultPort, "Port to run the server on")
	debug := flag.Bool("debug", false, "Enable debug logging")
	configPath := flag.String("config", "", "Path to configuration file (overrides CONFIG_PATH env variable)")
	logsLimit := flag.Int("logs-limit", defaultLogsLimit, "Maximum number of logs to return in `query` response")
	flag.Parse()

	// Setup logger
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})

	// Set log level based on debug flag
	if *debug {
		logger.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		logger.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = logger.Writer()
	gin.DefaultErrorWriter = logger.WriterLevel(log.ErrorLevel)

	// Load configuration
	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}
	if cfgPath == "" {
		log.Fatalln("CONFIG_PATH environment variable is not set and --config flag is not provided")
	}

	conf, err := config.NewConfig(cfgPath, *logsLimit)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	handlerInstance := handler.NewHandler(conf)

	r := gin.Default()

	// Add middleware to check for empty configuration
	r.Use(middleware.EmptyConfigMiddleware(conf))

	// Setup router
	router.SetupRoutes(r, handlerInstance)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Infof("Starting server on %s (debug=%v)", addr, *debug)
	if err := r.Run(addr); err != nil {
		log.Panicf("failed to start server: %v", err)
	}
}
