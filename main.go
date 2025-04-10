package main

import (
	"strings"

	"github.com/kdjuwidja/aishoppercommon/logger"
	commonOs "github.com/kdjuwidja/aishoppercommon/os"
	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/oauth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	err := logger.Init("core")
	if err != nil {
		panic(err)
	}

	mysqlConn := &db.MySQLConnectionPool{}
	mysqlConn.Configure(commonOs.GetEnvString("AI_SHOPPER_CORE_DB_USER", "ai_shopper_dev"),
		commonOs.GetEnvString("AI_SHOPPER_CORE_DB_PASSWORD", "password"),
		commonOs.GetEnvString("AI_SHOPPER_CORE_DB_HOST", "localhost"),
		commonOs.GetEnvString("AI_SHOPPER_CORE_DB_PORT", "3306"),
		commonOs.GetEnvString("AI_SHOPPER_CORE_DB_NAME", "ai_shopper_core"),
		commonOs.GetEnvInt("AI_SHOPPER_CORE_DB_MAX_OPEN_CONNS", 25),
		commonOs.GetEnvInt("AI_SHOPPER_CORE_DB_MAX_IDLE_CONNS", 10))

	// Initialize database connection
	_, err = db.InitializeMySQLConnPoolSingleton(mysqlConn)
	if err != nil {
		logger.Errorf("Failed to initialize database connection: %v", err)
		panic(err)
	}
	defer mysqlConn.Close()

	// Migrate database
	logger.Info("Migrating database...")
	mysqlConn.AutoMigrate()
	logger.Info("Database migrated successfully")

	r := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()

	origins := commonOs.GetEnvString("CORS_ALLOW_ORIGINS", "http://localhost:5173")
	corsConfig.AllowOrigins = strings.Split(origins, ",")

	methods := commonOs.GetEnvString("CORS_ALLOW_METHODS", "GET, POST, PUT, DELETE, OPTIONS")
	corsConfig.AllowMethods = strings.Split(methods, ",")

	headers := commonOs.GetEnvString("CORS_ALLOW_HEADERS", "Content-Type, Authorization")
	corsConfig.AllowHeaders = strings.Split(headers, ",")

	r.Use(cors.New(corsConfig))

	// Initialize API Handlers
	userProfileHandler := apiHandlers.InitializeUserProfileHandler(*mysqlConn)
	shoplistHandler := apiHandlers.InitializeShoplistHandler(*mysqlConn)

	r.GET("/ping", apiHandlers.Ping)
	r.GET("/v1/user", oauth.VerifyToken([]string{"profile"}, userProfileHandler.GetUserProfile))
	r.POST("/v1/user", oauth.VerifyToken([]string{"profile"}, userProfileHandler.CreateOrUpdateUserProfile))
	r.PUT("/v1/shoplist", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.CreateShoplist))
	r.GET("/v1/shoplist", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.GetAllShoplists))
	r.GET("/v1/shoplist/:id", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.GetShoplist))
	r.POST("/v1/shoplist/:id", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.UpdateShoplist))
	r.POST("/v1/shoplist/:id/leave", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.LeaveShopList))
	r.POST("/v1/shoplist/:id/share-code", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.RequestShopListShareCode))
	r.POST("/v1/shoplist/:id/share-code/revoke", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.RevokeShopListShareCode))
	r.POST("/v1/shoplist/join", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.JoinShopList))
	r.PUT("/v1/shoplist/:id/item", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.AddItemToShopList))
	r.DELETE("/v1/shoplist/:id/item/:itemId", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.RemoveItemFromShopList))
	r.POST("/v1/shoplist/:id/item/:itemId", oauth.VerifyToken([]string{"shoplist"}, shoplistHandler.UpdateShoplistItem))

	logger.Info("Starting server on port 8080")
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
