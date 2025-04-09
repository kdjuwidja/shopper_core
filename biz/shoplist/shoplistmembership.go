package bizshoplist

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"netherrealmstudio.com/aishoppercore/m/model"
	"netherrealmstudio.com/aishoppercore/m/util"
)

func (b *ShoplistBiz) LeaveShopList(userID string, shoplistID int) *ShoplistError {
	shopListData, shopListErr := b.GetShoplistWithMembers(shoplistID)
	if shopListErr != nil {
		return shopListErr
	}

	// check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		return NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist.")
	}

	//If no other members, delete the shoplist
	if len(shopListData.Members) == 1 {
		// Use transaction to batch remove member and delete shoplist
		if err := b.dbPool.GetDB().Transaction(func(tx *gorm.DB) error {
			// First remove the member
			if err := tx.Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
				return err
			}

			// Then delete the shoplist
			if err := tx.Unscoped().Delete(&model.Shoplist{}, shoplistID).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return NewShoplistError(ShoplistFailedToProcess, "Failed to remove member and delete shoplist")
		}
		return nil
	}

	//If user is owner, transfer ownership to another member
	if shopListData.OwnerID == userID {
		// Find another member to transfer ownership to
		var newOwnerID string
		for memberID := range shopListData.Members {
			if memberID != userID {
				newOwnerID = memberID
				break
			}
		}

		// Use transaction to batch transfer ownership and remove member
		if err := b.dbPool.GetDB().Transaction(func(tx *gorm.DB) error {
			// Transfer ownership
			if err := tx.Model(&model.Shoplist{}).Where("id = ?", shoplistID).Update("owner_id", newOwnerID).Error; err != nil {
				return err
			}

			// Remove member
			if err := tx.Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return NewShoplistError(ShoplistFailedToProcess, "Failed to transfer ownership and remove member")
		}
		return nil
	}

	//If user is not owner, remove user from shoplist
	if err := b.dbPool.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to remove member")
	}

	return nil
}

func (b *ShoplistBiz) RequestShopListShareCode(userID string, shoplistID int) (*model.ShoplistShareCode, *ShoplistError) {
	shopListData, shopListErr := b.GetShoplistWithMembers(shoplistID)
	if shopListErr != nil {
		return nil, shopListErr
	}

	// check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		return nil, NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist.")
	}

	// check if user is the owner
	if shopListData.OwnerID != userID {
		return nil, NewShoplistError(ShoplistNotOwner, "Only the owner can generate share codes.")
	}

	tx := b.dbPool.GetDB().Begin()

	// Generate a share code that is unique among all active share codes (6 characters, alphanumeric)
	var shareCode string
	for {
		shareCode = util.GenerateShareCode(6)
		if util.VerifyShareCodeFromDB(tx, shareCode) {
			break
		}
	}
	expiresAt := time.Now().Add(24 * time.Hour) // Share code expires in 24 hours

	// Create or update share code record
	shareCodeRecord := model.ShoplistShareCode{
		ShopListID: shoplistID,
		Code:       shareCode,
		Expiry:     expiresAt,
	}

	// Upsert the share code record
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "shop_list_id"}},
		UpdateAll: true,
	}).Create(&shareCodeRecord).Error; err != nil {
		tx.Rollback()
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to generate share code")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, NewShoplistError(ShoplistFailedToProcess, "Failed to commit transaction")
	}

	return &shareCodeRecord, nil
}

func (b *ShoplistBiz) RevokeShopListShareCode(userID string, shoplistID int) *ShoplistError {
	shopListData, shopListErr := b.GetShoplistWithMembers(shoplistID)
	if shopListErr != nil {
		return shopListErr
	}

	// check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		return NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist.")
	}

	// check if user is the owner
	if shopListData.OwnerID != userID {
		return NewShoplistError(ShoplistNotOwner, "Only the owner can generate share codes.")
	}

	// Find the active share code
	var shareCode model.ShoplistShareCode
	err := b.dbPool.GetDB().Where("shop_list_id = ? AND expiry > ?", shoplistID, time.Now()).First(&shareCode).Error
	if err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to find active share code")
	}

	// Update the expiry to current time to revoke the code
	if err := b.dbPool.GetDB().Model(&shareCode).Update("expiry", time.Now()).Error; err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to revoke share code")
	}

	return nil
}

func (b *ShoplistBiz) JoinShopList(userID string, shareCode string) *ShoplistError {
	var dbShareCode model.ShoplistShareCode
	err := b.dbPool.GetDB().Where("code = ? AND expiry > ?", shareCode, time.Now()).First(&dbShareCode).Error
	if err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Invalid share code")
	}

	// Check if user is already a member
	var existingMember model.ShoplistMember
	err = b.dbPool.GetDB().Where("shop_list_id = ? AND member_id = ?", dbShareCode.ShopListID, userID).First(&existingMember).Error
	if err == nil {
		return NewShoplistError(ShoplistFailedToProcess, "User is already a member of the shoplist")
	}

	// Add user as member
	newMember := model.ShoplistMember{
		ShopListID: dbShareCode.ShopListID,
		MemberID:   userID,
	}

	if err := b.dbPool.GetDB().Create(&newMember).Error; err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to join shoplist")
	}

	return nil
}
