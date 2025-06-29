package main

import (
	"fmt"
	"strings"

	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"github.com/kdjuwidja/aishoppercommon/osutil"
	"netherealmstudio.com/m/v2/apiHandlers"
	apiHandlersHealth "netherealmstudio.com/m/v2/apiHandlers/health"
	apiHandlersmatch "netherealmstudio.com/m/v2/apiHandlers/match"
	apiHandlerssearch "netherealmstudio.com/m/v2/apiHandlers/search"
	apiHandlersshoplist "netherealmstudio.com/m/v2/apiHandlers/shoplist"
	apihandlersuser "netherealmstudio.com/m/v2/apiHandlers/user"
	dbmodel "netherealmstudio.com/m/v2/db"

	bizmatch "netherealmstudio.com/m/v2/biz/match"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"

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
	trustProxiesConf := osutil.GetEnvString("TRUST_PROXIES", "127.0.0.1")
	trustProxies := strings.Split(trustProxiesConf, ",")
	r.SetTrustedProxies(trustProxies)

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

	// IntializeBiz
	shoplistBiz := bizshoplist.InitializeShoplistBiz(*mysqlConn)
	matchBiz := bizmatch.NewMatchShoplistItemsWithFlyerBiz(esc, mysqlConn)

	// Initialize API Handlers
	healthHandler := apiHandlersHealth.InitializeHealthHandler()
	userProfileHandler := apihandlersuser.InitializeUserProfileHandler(*mysqlConn, *rf)
	shoplistHandler := apiHandlersshoplist.InitializeShoplistHandler(*mysqlConn, shoplistBiz, matchBiz, *rf)
	searchHandler := apiHandlerssearch.InitializeSearchHandler(*esc, *rf)
	matchHandler := apiHandlersmatch.InitializeMatchHandler(*esc, *mysqlConn, *rf)

	serviceName := osutil.GetEnvString("SERVICE_NAME", "core")

	r.GET(getRoute(serviceName, "/health"), healthHandler.Health)
	r.GET(getRoute(serviceName, "/v2/user"), tokenVerifier.VerifyToken([]string{"profile"}, userProfileHandler.GetUserProfile))
	r.POST(getRoute(serviceName, "/v2/user"), tokenVerifier.VerifyToken([]string{"profile"}, userProfileHandler.CreateOrUpdateUserProfile))
	r.PUT(getRoute(serviceName, "/v2/shoplist"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.CreateShoplist))
	r.POST(getRoute(serviceName, "/v2/shoplist/:id"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.UpdateShoplist))
	r.POST(getRoute(serviceName, "/v2/shoplist/:id/leave"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.LeaveShopList))
	r.POST(getRoute(serviceName, "/v2/shoplist/:id/share-code"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.RequestShopListShareCode))
	r.POST(getRoute(serviceName, "/v2/shoplist/:id/share-code/revoke"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.RevokeShopListShareCode))
	r.POST(getRoute(serviceName, "/v2/shoplist/join"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.JoinShopList))
	r.PUT(getRoute(serviceName, "/v2/shoplist/:id/item"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.AddItemToShopList))
	r.DELETE(getRoute(serviceName, "/v2/shoplist/:id/item/:itemId"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.RemoveItemFromShopList))
	r.POST(getRoute(serviceName, "/v2/shoplist/:id/item/:itemId"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.UpdateShoplistItem))
	r.GET(getRoute(serviceName, "/v2/search/flyers"), tokenVerifier.VerifyToken([]string{"search"}, searchHandler.SearchFlyers))
	r.GET(getRoute(serviceName, "/v2/match/flyers"), tokenVerifier.VerifyToken([]string{"search"}, matchHandler.MatchShoplistItemsWithFlyer))
	r.GET(getRoute(serviceName, "/v2/shoplist"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.GetAllShoplistAndItemsForUser))
	r.GET(getRoute(serviceName, "/v2/shoplist/:id"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.GetShoplistAndItemsForUserByShoplistID))
	r.GET(getRoute(serviceName, "/v2/shoplist/:id/members"), tokenVerifier.VerifyToken([]string{"shoplist"}, shoplistHandler.GetShoplistMembers))

	logger.Info("Starting server on port 8080")
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getRoute(serviceName string, route string) string {
	return fmt.Sprintf("/%s%s", serviceName, route)
}
