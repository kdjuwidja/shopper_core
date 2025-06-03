package bizmatch

import (
	"context"

	bizmodels "netherealmstudio.com/m/v2/biz"
	dbmodels "netherealmstudio.com/m/v2/db"
)

func (b *MatchShoplistItemsWithFlyerBiz) GetShoplistItems(ctx context.Context, userId string, shoplistItemIds []int) ([]bizmodels.ShoplistItem, *MatchError) {
	if len(shoplistItemIds) == 0 {
		return nil, NewMatchError(MatchItemIdNotProvided, "No item IDs provided.")
	}

	var shoplistItems []*dbmodels.ShoplistItem
	err := b.dbPool.GetDB().WithContext(ctx).Raw(`
		SELECT shoplist_items.id as id, 
			   shoplist_items.item_name as item_name, 
			   shoplist_items.brand_name as brand_name
		FROM (SELECT * FROM shoplist_members WHERE member_id = ?) as tbl1
		LEFT JOIN shoplist_items
		ON shoplist_items.shop_list_id = tbl1.shop_list_id
		WHERE shoplist_items.id IN (?)
	`, userId, shoplistItemIds).Scan(&shoplistItems).Error
	if err != nil {
		return nil, NewMatchError(MatchFailedToProcess, "Failed to get shoplist items.")
	}

	if len(shoplistItems) == 0 {
		return nil, NewMatchError(MatchShoplistItemNotFound, "No shoplist items found.")
	}

	shoplistItemsBiz := make([]bizmodels.ShoplistItem, 0)
	for _, shoplistItem := range shoplistItems {
		shoplistItemsBiz = append(shoplistItemsBiz, bizmodels.ShoplistItem{
			ID:        shoplistItem.ID,
			ItemName:  shoplistItem.ItemName,
			BrandName: shoplistItem.BrandName,
		})
	}

	return shoplistItemsBiz, nil
}
