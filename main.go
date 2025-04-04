package main

import (
	"os"
	"strconv"
	"strings"

	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/logger"
	"netherrealmstudio.com/aishoppercore/m/oauth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getEnvInt(key string, defaultValue int) int {
	if val, err := strconv.Atoi(os.Getenv(key)); err == nil && val > 0 {
		return val
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	// Initialize logger
	err := logger.Init()
	if err != nil {
		panic(err)
	}

	// Initialize database connection
	err = db.Initialize(&db.Config{
		Host:         getEnvString("AI_SHOPPER_CORE_DB_HOST", "localhost"),
		Port:         getEnvString("AI_SHOPPER_CORE_DB_PORT", "3306"),
		DBName:       getEnvString("AI_SHOPPER_CORE_DB_NAME", "ai_shopper_core"),
		User:         getEnvString("AI_SHOPPER_CORE_DB_USER", "ai_shopper_dev"),
		Password:     getEnvString("AI_SHOPPER_CORE_DB_PASSWORD", "password"),
		MaxOpenConns: getEnvInt("AI_SHOPPER_CORE_DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: getEnvInt("AI_SHOPPER_CORE_DB_MAX_IDLE_CONNS", 10),
	})
	if err != nil {
		logger.Errorf("Failed to initialize database connection: %v", err)
		panic(err)
	}
	defer db.Close()

	// Migrate database
	logger.Info("Migrating database...")
	db.AutoMigrate()
	logger.Info("Database migrated successfully")

	r := gin.Default()

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

	r.GET("/ping", apiHandlers.Ping)
	r.GET("/v1/user", oauth.VerifyToken([]string{"profile"}, apiHandlers.GetUserProfile))
	r.POST("/v1/user", oauth.VerifyToken([]string{"profile"}, apiHandlers.CreateOrUpdateUserProfile))
	r.PUT("/v1/shoplist", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.CreateShoplist))
	r.GET("/v1/shoplist", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.GetAllShoplists))
	r.GET("/v1/shoplist/:id", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.GetShoplist))
	r.POST("/v1/shoplist/:id", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.UpdateShoplist))
	r.POST("/v1/shoplist/:id/leave", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.LeaveShopList))
	r.POST("/v1/shoplist/:id/share-code", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.RequestShopListShareCode))
	r.POST("/v1/shoplist/:id/share-code/revoke", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.RevokeShopListShareCode))
	r.POST("/v1/shoplist/join", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.JoinShopList))
	r.PUT("/v1/shoplist/:id/item", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.AddItemToShopList))
	r.DELETE("/v1/shoplist/:id/item/:itemId", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.RemoveItemFromShopList))
	r.POST("/v1/shoplist/:id/item/:itemId", oauth.VerifyToken([]string{"shoplist"}, apiHandlers.UpdateShoplistItem))

	logger.Info("Starting server on port 8080")
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
