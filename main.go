package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kdjuwidja/kafkalib/kafka"
	"netherrealmstudio.com/aishoppercore/m/APIHandlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Initialize Kafka factory
	if _, err := kafka.GetKafkaFactory(); err != nil {
		fmt.Printf("Warning: Failed to initialize Kafka factory: %v\n", err)
	}

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	if origins := os.Getenv("CORS_ALLOW_ORIGINS"); origins != "" {
		corsConfig.AllowOrigins = strings.Split(origins, ",")
	} else {
		corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	}
	if methods := os.Getenv("CORS_ALLOW_METHODS"); methods != "" {
		corsConfig.AllowMethods = strings.Split(methods, ",")
	}
	if headers := os.Getenv("CORS_ALLOW_HEADERS"); headers != "" {
		corsConfig.AllowHeaders = strings.Split(headers, ",")
	}
	r.Use(cors.New(corsConfig))

	r.GET("/ping", APIHandlers.Ping)
	r.GET("/getLatLngByAddress", APIHandlers.GetLatLngByAddress)
	r.POST("/recommend", APIHandlers.Recommend)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
