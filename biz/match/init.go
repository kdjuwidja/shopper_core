package bizmatch

import (
	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
)

type MatchShoplistItemsWithFlyerBiz struct {
	esc    *elasticsearch.ElasticsearchClient
	dbPool *db.MySQLConnectionPool
}

func NewMatchShoplistItemsWithFlyerBiz(esc *elasticsearch.ElasticsearchClient, dbPool *db.MySQLConnectionPool) *MatchShoplistItemsWithFlyerBiz {
	return &MatchShoplistItemsWithFlyerBiz{esc: esc, dbPool: dbPool}
}
