package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/k0rdent/victorialogs-aggregator/interanl/config"
	"github.com/k0rdent/victorialogs-aggregator/interanl/router"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// func init() {
// 	log.SetFormatter(&log.TextFormatter{
// 		FullTimestamp: true,
// 	})
// 	log.SetOutput(gin.DefaultWriter)
// 	log.SetLevel(log.DebugLevel)
// }

func main() {

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	gin.DefaultWriter = logger.Writer()

	gin.DefaultErrorWriter = logger.WriterLevel(logrus.ErrorLevel)
	r := gin.Default()
	router.SetupRoutes(r)
	log.Info("Starting server on :8085")

	configPath := os.Getenv("CONFIG_PATH")
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Panicf("config not loaded: %v", err)
	}
	config.GlobalConfig = conf

	r.Run(":8085")
}
