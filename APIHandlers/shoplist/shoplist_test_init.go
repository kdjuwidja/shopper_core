package apiHandlersshoplist

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func setUpShoplistTestEnv(t *testing.T) (*ShoplistHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	shoplistHandler := InitializeShoplistHandler(*testDBConn, apiHandlers.ResponseFactory{})
	return shoplistHandler, testDBConn
}
