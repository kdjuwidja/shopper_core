package bizshoplist

import "github.com/kdjuwidja/aishoppercommon/db"

// ShoplistBiz dependencies
type ShoplistBiz struct {
	dbPool db.MySQLConnectionPool
}

// Dependency Injection for ShoplistBiz
func InitializeShoplistBiz(dbPool db.MySQLConnectionPool) *ShoplistBiz {
	return &ShoplistBiz{
		dbPool: dbPool,
	}
}
