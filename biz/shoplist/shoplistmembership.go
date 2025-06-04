package bizshoplist

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	bizmodels "netherealmstudio.com/m/v2/biz"
	"netherealmstudio.com/m/v2/db"
)

func (b *ShoplistBiz) GetShoplistMembers(ctx context.Context, userID string, shoplistID int) ([]bizmodels.ShoplistMember, *ShoplistError) {
	shopListData, shopListErr := b.GetShoplistWithMembers(ctx, shoplistID)
	if shopListErr != nil {
		return nil, shopListErr
	}

	// check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		return nil, NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist.")
	}

	result := make([]bizmodels.ShoplistMember, 0)
	for _, member := range shopListData.Members {
		shoplistMember := bizmodels.ShoplistMember{
			ID:       member.MemberID,
			Nickname: member.Nickname,
		}
		result = append(result, shoplistMember)
	}

	return result, nil
}

func (b *ShoplistBiz) LeaveShopList(ctx context.Context, userID string, shoplistID int) *ShoplistError {
	shopListData, shopListErr := b.GetShoplistWithMembers(ctx, shoplistID)
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
		if err := b.dbPool.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// First remove the member
			if err := tx.Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&db.ShoplistMember{}).Error; err != nil {
				return err
			}

			// Delete any share code record for the shoplist
			if err := tx.Where("shop_list_id = ?", shoplistID).Unscoped().Delete(&db.ShoplistShareCode{}).Error; err != nil {
				return err
			}

			// Delete any items for the shoplist
			if err := tx.Where("shop_list_id = ?", shoplistID).Unscoped().Delete(&db.ShoplistItem{}).Error; err != nil {
				return err
			}

			// Then delete the shoplist
			if err := tx.Unscoped().Delete(&db.Shoplist{}, shoplistID).Error; err != nil {
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
		if err := b.dbPool.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// Transfer ownership
			if err := tx.Model(&db.Shoplist{}).Where("id = ?", shoplistID).Update("owner_id", newOwnerID).Error; err != nil {
				return err
			}

			// Remove member
			if err := tx.Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&db.ShoplistMember{}).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return NewShoplistError(ShoplistFailedToProcess, "Failed to transfer ownership and remove member")
		}
		return nil
	}

	//If user is not owner, remove user from shoplist
	if err := b.dbPool.GetDB().WithContext(ctx).Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&db.ShoplistMember{}).Error; err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to remove member")
	}

	return nil
}

func (b *ShoplistBiz) RequestShopListShareCode(ctx context.Context, userID string, shoplistID int) (*db.ShoplistShareCode, *ShoplistError) {
	shopListData, shopListErr := b.GetShoplistWithMembers(ctx, shoplistID)
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

	tx := b.dbPool.GetDB().WithContext(ctx).Begin()

	// Generate a share code that is unique among all active share codes (6 characters, alphanumeric)
	var shareCode string
	for {
		shareCode = GenerateShareCode(6)
		if VerifyShareCodeFromDB(tx, shareCode) {
			break
		}
	}
	expiresAt := time.Now().Add(24 * time.Hour) // Share code expires in 24 hours

	// Create or update share code record
	shareCodeRecord := db.ShoplistShareCode{
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

func (b *ShoplistBiz) RevokeShopListShareCode(ctx context.Context, userID string, shoplistID int) *ShoplistError {
	shopListData, shopListErr := b.GetShoplistWithMembers(ctx, shoplistID)
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
	var shareCode db.ShoplistShareCode
	err := b.dbPool.GetDB().WithContext(ctx).Where("shop_list_id = ? AND expiry > ?", shoplistID, time.Now()).First(&shareCode).Error
	if err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to find active share code")
	}

	// Update the expiry to current time to revoke the code
	if err := b.dbPool.GetDB().WithContext(ctx).Model(&shareCode).Update("expiry", time.Now()).Error; err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to revoke share code")
	}

	return nil
}

func (b *ShoplistBiz) JoinShopList(ctx context.Context, userID string, shareCode string) *ShoplistError {
	var dbShareCode db.ShoplistShareCode
	err := b.dbPool.GetDB().WithContext(ctx).Where("code = ? AND expiry > ?", shareCode, time.Now()).First(&dbShareCode).Error
	if err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Invalid share code")
	}

	// Check if user is already a member
	var existingMember db.ShoplistMember
	err = b.dbPool.GetDB().WithContext(ctx).Where("shop_list_id = ? AND member_id = ?", dbShareCode.ShopListID, userID).First(&existingMember).Error
	if err == nil {
		return NewShoplistError(ShoplistFailedToProcess, "User is already a member of the shoplist")
	}

	// Add user as member
	newMember := db.ShoplistMember{
		ShopListID: dbShareCode.ShopListID,
		MemberID:   userID,
	}

	if err := b.dbPool.GetDB().WithContext(ctx).Create(&newMember).Error; err != nil {
		return NewShoplistError(ShoplistFailedToProcess, "Failed to join shoplist")
	}

	return nil
}
