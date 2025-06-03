package apiHandlersshoplist

import (
	"github.com/kdjuwidja/aishoppercommon/db"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizmatch "netherealmstudio.com/m/v2/biz/match"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"
)

type ShoplistHandler struct {
	shoplistBiz     *bizshoplist.ShoplistBiz
	matchBiz        *bizmatch.MatchShoplistItemsWithFlyerBiz
	responseFactory apiHandlers.ResponseFactory
}

// Dependency Injection for ShoplistHandler
func InitializeShoplistHandler(dbPool db.MySQLConnectionPool, shoplistBiz *bizshoplist.ShoplistBiz, matchBiz *bizmatch.MatchShoplistItemsWithFlyerBiz, responseFactory apiHandlers.ResponseFactory) *ShoplistHandler {
	return &ShoplistHandler{
		shoplistBiz:     shoplistBiz,
		matchBiz:        matchBiz,
		responseFactory: responseFactory,
	}
}
