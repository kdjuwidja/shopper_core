package bizshoplist

import "netherrealmstudio.com/aishoppercore/m/db"

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

// getShoplistWithMembers retrieves shoplist data including owner and members
type ShoplistData struct {
	ShopListID int
	OwnerID    string
	Members    map[string]struct{ MemberID string }
}

// GetAllShoplistData is the response data structure for GetAllShoplists
type GetAllShoplistData struct {
	ID            int    `json:"id" gorm:"column:shop_list_id"`
	Name          string `json:"name" gorm:"column:shop_list_name"`
	OwnerID       string `json:"owner_id" gorm:"column:owner_id"`
	OwnerNickname string `json:"owner_nickname" gorm:"column:owner_nickname"`
}

// GetShoplistData is the response data structure for GetShoplist
type GetShoplistData struct {
	ID                     int     `json:"id" gorm:"column:shop_list_id"`
	Name                   string  `json:"name" gorm:"column:shop_list_name"`
	OwnerId                string  `json:"owner_id" gorm:"column:owner_id"`
	ShopListItemID         *int    `json:"shop_list_item_id" gorm:"column:shop_list_item_id"`
	ShopListItemName       *string `json:"item_name" gorm:"column:item_name"`
	ShopListItemBrandName  *string `json:"brand_name" gorm:"column:brand_name"`
	ShopListItemExtraInfo  *string `json:"extra_info" gorm:"column:extra_info"`
	ShopListItemIsBought   *bool   `json:"is_bought" gorm:"column:is_bought"`
	ShopListMemberID       string  `json:"member_id" gorm:"column:member_id"`
	ShopListMemberNickname string  `json:"member_nickname" gorm:"column:member_nickname"`
}
