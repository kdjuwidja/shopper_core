package apiHandlersmatch

import (
	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizmatch "netherealmstudio.com/m/v2/biz/match"
)

type MatchHandler struct {
	matchFlyerBiz   *bizmatch.MatchShoplistItemsWithFlyerBiz
	responseFactory apiHandlers.ResponseFactory
}

func InitializeMatchHandler(esc elasticsearch.ElasticsearchClient, dbPool db.MySQLConnectionPool, responseFactory apiHandlers.ResponseFactory) *MatchHandler {
	return &MatchHandler{
		matchFlyerBiz:   bizmatch.NewMatchShoplistItemsWithFlyerBiz(&esc, &dbPool),
		responseFactory: responseFactory,
	}
}
