package apiHandlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
	"netherrealmstudio.com/aishoppercore/m/util"
)

// LeaveShopList allows a user to leave a shoplist
// @Summary Leave a shoplist
// @Description Allows a user to leave a shoplist. If the user is the owner, ownership will be transferred to another member. If the user is the last member, the shoplist will be deleted.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} map[string]interface{} "Successfully left the shoplist"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shoplist/{id}/leave [POST]
func LeaveShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// If no other members, delete the shoplist
	if len(shopListData.Members) == 1 {
		// Use transaction to batch remove member and delete shoplist
		if err := db.GetDB().Transaction(func(tx *gorm.DB) error {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member and delete shoplist"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully left the shoplist"})
		return
	}

	// If user is owner, transfer ownership to another member
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
		if err := db.GetDB().Transaction(func(tx *gorm.DB) error {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transfer ownership and remove member"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully left the shoplist"})
		return
	}

	// Remove member
	if err := db.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).Unscoped().Delete(&model.ShoplistMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the shoplist"})
}

// RequestShopListShareCode generates a share code for a shoplist
// @Summary Generate share code
// @Description Generates a unique share code for a shoplist that can be used by other users to join. Only the owner can generate share codes. The code expires in 24 hours.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} map[string]interface{} "Successfully generated share code"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Only the owner can generate share codes"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shoplist/{id}/share-code [POST]
func RequestShopListShareCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if shopListData.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can generate share codes"})
		return
	}

	tx := db.GetDB().Begin()

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate share code"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share_code": shareCode,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// RevokeShopListShareCode revokes the active share code for a shoplist
// @Summary Revoke a share code
// @Description Revokes the active share code for a specific shoplist. Only the owner can revoke share codes.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} gin.H "Successfully revoked share code"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 403 {object} map[string]string "Only the owner can revoke share codes"
// @Failure 404 {object} map[string]string "Not found"
// @Router /shoplist/{id}/share-code/revoke [post]
func RevokeShopListShareCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	// Get shoplist data including members
	shopListData, err := getShoplistWithMembers(shoplistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process shoplist data"})
		return
	}

	// Check if user is a member
	if _, exists := shopListData.Members[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Check if user is the owner
	if shopListData.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can revoke share codes"})
		return
	}

	// Find the active share code
	var shareCode model.ShoplistShareCode
	err = db.GetDB().Where("shop_list_id = ? AND expiry > ?", shoplistID, time.Now()).First(&shareCode).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Update the expiry to current time to revoke the code
	if err := db.GetDB().Model(&shareCode).Update("expiry", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke share code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// JoinShopList allows a user to join a shoplist using a share code
// @Summary Join a shoplist using a share code
// @Description Allows a user to join a shoplist by providing a valid share code. The share code must be active and not expired.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param share_code body string true "Share code to join the shoplist"
// @Success 200 {object} gin.H "Successfully joined the shoplist"
// @Failure 400 {object} map[string]string "Share code is required"
// @Failure 400 {object} map[string]string "Invalid share code"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/join [post]
func JoinShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request body
	var requestBody struct {
		ShareCode string `json:"share_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	// Find the active share code
	var shareCode model.ShoplistShareCode
	err := db.GetDB().Where("code = ? AND expiry > ?", requestBody.ShareCode, time.Now()).First(&shareCode).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid share code"})
		return
	}

	// Check if user is already a member
	var existingMember model.ShoplistMember
	err = db.GetDB().Where("shop_list_id = ? AND member_id = ?", shareCode.ShopListID, userID).First(&existingMember).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Add user as member
	newMember := model.ShoplistMember{
		ShopListID: shareCode.ShopListID,
		MemberID:   userID,
	}
	if err := db.GetDB().Create(&newMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join shoplist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
