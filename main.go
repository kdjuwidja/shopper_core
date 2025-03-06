package main

import (
	"fmt"
	"os"
	"strings"

	"netherrealmstudio.com/aishoppercore/m/APIHandlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Configure CORS from environment variables
	config := cors.DefaultConfig()

	// Parse comma-separated origins from environment
	origins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	if len(origins) == 0 {
		fmt.Fprintf(os.Stderr, "Error: CORS_ALLOW_ORIGINS environment variable is required\n")
		os.Exit(1)
	}
	config.AllowOrigins = origins

	// Parse comma-separated methods from environment
	methods := strings.Split(os.Getenv("CORS_ALLOW_METHODS"), ",")
	if len(methods) == 0 {
		fmt.Fprintf(os.Stderr, "Error: CORS_ALLOW_METHODS environment variable is required\n")
		os.Exit(1)
	}
	config.AllowMethods = methods

	// Parse comma-separated headers from environment
	headers := strings.Split(os.Getenv("CORS_ALLOW_HEADERS"), ",")
	if len(headers) == 0 {
		fmt.Fprintf(os.Stderr, "Error: CORS_ALLOW_HEADERS environment variable is required\n")
		os.Exit(1)
	}
	config.AllowHeaders = headers

	r.Use(cors.New(config))

	r.GET("/ping", APIHandlers.Ping)
	r.GET("/getLatLngByAddress", APIHandlers.GetLatLngByAddress)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
