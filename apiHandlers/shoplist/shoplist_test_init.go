package apiHandlersshoplist

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/elasticsearch"
	"github.com/stretchr/testify/require"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizmatch "netherealmstudio.com/m/v2/biz/match"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"
	testutil "netherealmstudio.com/m/v2/testUtil"
)

const (
	elasticsearchHost = "localhost"
	elasticsearchPort = "10200"
)

func setUpShoplistTestEnv(t *testing.T) (*ShoplistHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	esc, err := elasticsearch.NewElasticsearchClient(elasticsearchHost, elasticsearchPort)
	require.NoError(t, err)
	shoplistBiz := bizshoplist.InitializeShoplistBiz(*testDBConn)
	matchBiz := bizmatch.NewMatchShoplistItemsWithFlyerBiz(esc, testDBConn)
	shoplistHandler := InitializeShoplistHandler(*testDBConn, shoplistBiz, matchBiz, apiHandlers.ResponseFactory{})
	return shoplistHandler, testDBConn
}
