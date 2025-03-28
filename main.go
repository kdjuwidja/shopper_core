package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
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
	// Initialize database connection
	err := db.Initialize(&db.Config{
		Host:         getEnvString("AI_SHOPPER_CORE_DB_HOST", "localhost"),
		Port:         getEnvString("AI_SHOPPER_CORE_DB_PORT", "3306"),
		DBName:       getEnvString("AI_SHOPPER_CORE_DB_NAME", "ai_shopper_core"),
		User:         getEnvString("AI_SHOPPER_CORE_DB_USER", "ai_shopper_dev"),
		Password:     getEnvString("AI_SHOPPER_CORE_DB_PASSWORD", "password"),
		MaxOpenConns: getEnvInt("AI_SHOPPER_CORE_DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns: getEnvInt("AI_SHOPPER_CORE_DB_MAX_IDLE_CONNS", 10),
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("Migrating database...")
	if err := db.GetDB().AutoMigrate(&model.User{}, &model.Shoplist{}, &model.ShoplistMember{}, &model.ShoplistItem{}, &model.ShoplistShareCode{}); err != nil {
		panic(err)
	}
	fmt.Println("Database migrated successfully")

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
	r.GET("/getLatLngByAddress", apiHandlers.GetLatLngByAddress)
	r.POST("/recommend", apiHandlers.Recommend)
	r.GET("/user", oauth.VerifyToken([]string{"profile"}, apiHandlers.GetUserProfile))
	r.POST("/user", oauth.VerifyToken([]string{"profile"}, apiHandlers.CreateOrUpdateUserProfile))
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
