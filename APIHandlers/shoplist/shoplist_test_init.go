package apiHandlersshoplist

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"netherealmstudio.com/m/v2/apiHandlers"
	testutil "netherealmstudio.com/m/v2/testUtil"
)

func setUpShoplistTestEnv(t *testing.T) (*ShoplistHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	shoplistHandler := InitializeShoplistHandler(*testDBConn, apiHandlers.ResponseFactory{})
	return shoplistHandler, testDBConn
}
