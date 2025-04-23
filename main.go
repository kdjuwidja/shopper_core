package main

import (
	"strings"

	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"github.com/kdjuwidja/aishoppercommon/osutil"
	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	apiHandlersping "netherrealmstudio.com/aishoppercore/m/apiHandlers/ping"
	apiHandlerssearch "netherrealmstudio.com/aishoppercore/m/apiHandlers/search"
	apiHandlersshoplist "netherrealmstudio.com/aishoppercore/m/apiHandlers/shoplist"
	apihandlersuser "netherrealmstudio.com/aishoppercore/m/apiHandlers/user"
	dbmodel "netherrealmstudio.com/aishoppercore/m/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// Initialize database connection

	models := []interface{}{
		&dbmodel.Shoplist{},
		&dbmodel.ShoplistItem{},
		&dbmodel.ShoplistMember{},
		&dbmodel.ShoplistShareCode{},
		&dbmodel.User{},
	}
	mysqlConn, err := db.InitializeMySQLConnectionPool(osutil.GetEnvString("AI_SHOPPER_CORE_DB_USER", "ai_shopper_dev"),
		osutil.GetEnvString("AI_SHOPPER_CORE_DB_PASSWORD", "password"),
		osutil.GetEnvString("AI_SHOPPER_CORE_DB_HOST", "localhost"),
		osutil.GetEnvString("AI_SHOPPER_CORE_DB_PORT", "3306"),
		osutil.GetEnvString("AI_SHOPPER_CORE_DB_NAME", "ai_shopper_core"),
		osutil.GetEnvInt("AI_SHOPPER_CORE_DB_MAX_OPEN_CONNS", 25),
		osutil.GetEnvInt("AI_SHOPPER_CORE_DB_MAX_IDLE_CONNS", 10),
		models,
	)
	if err != nil {
		logger.Fatalf("Failed to initialize MySQL connection pool: %v", err)
	}
	defer mysqlConn.Close()

	esc, err := elasticsearch.NewElasticsearchClient(osutil.GetEnvString("ELASTICSEARCH_HOST", "localhost"), osutil.GetEnvString("ELASTICSEARCH_PORT", "9200"))
	if err != nil {
		logger.Fatalf("Failed to initialize Elasticsearch client: %v", err)
	}

	// Migrate database
	logger.Info("Migrating database...")
	mysqlConn.AutoMigrate()
	logger.Info("Database migrated successfully")

	r := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()

	origins := osutil.GetEnvString("CORS_ALLOW_ORIGINS", "http://localhost:5173")
	corsConfig.AllowOrigins = strings.Split(origins, ",")

	methods := osutil.GetEnvString("CORS_ALLOW_METHODS", "GET, POST, PUT, DELETE, OPTIONS")
	corsConfig.AllowMethods = strings.Split(methods, ",")

	headers := osutil.GetEnvString("CORS_ALLOW_HEADERS", "Content-Type, Authorization")
	corsConfig.AllowHeaders = strings.Split(headers, ",")

	r.Use(cors.New(corsConfig))

	// Initialize Response Factory
	rf := apiHandlers.Initialize()

	// Initialize Token Verifier
	tokenVerifier := apiHandlers.InitializeTokenVerifier(*rf)

	// Initialize API Handlers
	userProfileHandler := apihandlersuser.InitializeUserProfileHandler(*mysqlConn, *rf)
	shoplistHandler := apiHandlersshoplist.InitializeShoplistHandler(*mysqlConn, *rf)
	searchHandler := apiHandlerssearch.InitializeSearchHandler(esc, *rf)

	r.GET("/ping", apiHandlersping.Ping)
	r.GET("/v1/user", tokenVerifier.VerifyToken([]string{"profile"}, userProfileHandler.GetUserProfile))
	r.POST("/v1/user", tokenVerifier.VerifyToken([]string{"profile"}, userProfileHandler.CreateOrUpdateUserProfile))
	r.PUT("/v1/shoplist", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.CreateShoplist))
	r.GET("/v1/shoplist", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.GetAllShoplists))
	r.GET("/v1/shoplist/:id", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.GetShoplist))
	r.POST("/v1/shoplist/:id", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.UpdateShoplist))
	r.POST("/v1/shoplist/:id/leave", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.LeaveShopList))
	r.POST("/v1/shoplist/:id/share-code", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.RequestShopListShareCode))
	r.POST("/v1/shoplist/:id/share-code/revoke", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.RevokeShopListShareCode))
	r.POST("/v1/shoplist/join", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.JoinShopList))
	r.PUT("/v1/shoplist/:id/item", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.AddItemToShopList))
	r.DELETE("/v1/shoplist/:id/item/:itemId", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.RemoveItemFromShopList))
	r.POST("/v1/shoplist/:id/item/:itemId", tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.UpdateShoplistItem))
	r.GET("/v1/search/flyers", tokenVerifier.VerifyToken([]string{"search"}, searchHandler.SearchFlyers))

	logger.Info("Starting server on port 8080")
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
