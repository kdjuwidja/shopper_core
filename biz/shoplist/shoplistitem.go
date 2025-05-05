package bizshoplist

import (
	"gorm.io/gorm"
	"netherealmstudio.com/m/v2/db"
)

func (b *ShoplistBiz) AddItemToShopList(userID string, shoplistID int, itemName string, brandName string, extraInfo string, thumbnail string) (*db.ShoplistItem, *ShoplistError) {
	if !b.checkShoplistMembershipFromDB(userID, shoplistID) {
		return nil, NewShoplistError(ShoplistNotFound, "Shoplist not found.")
	}

	// Create new item
	newItem := db.ShoplistItem{
		ShopListID: shoplistID,
		ItemName:   itemName,
		BrandName:  brandName,
		ExtraInfo:  extraInfo,
		IsBought:   false,
		Thumbnail:  thumbnail,
	}

	// check if item name is empty
	if itemName == "" {
		return nil, NewShoplistError(ShoplistItemNameEmpty, "Item name is required.")
	}

	if err := b.dbPool.GetDB().Create(&newItem).Error; err != nil {
		return nil, NewShoplistError(ShoplistFailedToCreate, "Failed to add item.")
	}

	return &newItem, nil
}

func (b *ShoplistBiz) RemoveItemFromShopList(userID string, shoplistID int, itemID int) *ShoplistError {
	shopListData, shopListErr := b.GetShoplistWithMembers(shoplistID)
	if shopListErr != nil {
		return shopListErr
	}

	// check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		return NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist.")
	}

	// check if item exists and belongs to the shoplist
	var item db.ShoplistItem
	err := b.dbPool.GetDB().Where("id = ? AND shop_list_id = ?", itemID, shoplistID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewShoplistError(ShoplistItemNotFound, "Item not found.")
		}
		return NewShoplistError(ShoplistFailedToProcess, "Failed to check item.")
	}

	// Delete the item
	err = b.dbPool.GetDB().Unscoped().Delete(&item).Error
	if err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to remove item.")
	}

	return nil
}

func (b *ShoplistBiz) UpdateShoplistItem(userID string, shoplistID int, itemID int, itemName *string, brandName *string, extraInfo *string, isBought *bool) (*db.ShoplistItem, *ShoplistError) {
	shopListData, shopListErr := b.GetShoplistWithMembers(shoplistID)
	if shopListErr != nil {
		return nil, shopListErr
	}

	// check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		return nil, NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist.")
	}

	//check if item exists and belongs to the shoplist
	var item db.ShoplistItem
	err := b.dbPool.GetDB().Where("id = ? AND shop_list_id = ?", itemID, shoplistID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewShoplistError(ShoplistItemNotFound, "Item not found.")
		}
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to check item.")
	}

	// Update only the fields that are provided in the request
	updates := make(map[string]interface{})
	if itemName != nil {
		updates["item_name"] = *itemName
	}
	if brandName != nil {
		updates["brand_name"] = *brandName
	}
	if extraInfo != nil {
		updates["extra_info"] = *extraInfo
	}
	if isBought != nil {
		updates["is_bought"] = *isBought
	}

	// Update the item
	err = b.dbPool.GetDB().Model(&item).Updates(updates).Error
	if err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to update item.")
	}

	return &item, nil
}
