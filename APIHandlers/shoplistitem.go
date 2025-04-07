package apiHandlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
)

// AddItemToShopList adds a new item to a shoplist
// @Summary Add a new item to a shoplist
// @Description Adds a new item to a specific shoplist. The user must be a member of the shoplist to add items.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
//
//	@Param request body struct {
//	    ItemName  string `json:"item_name" binding:"required"`
//	    BrandName string `json:"brand_name"`
//	    ExtraInfo string `json:"extra_info"`
//	} true "Item details"
//
// @Success 201 {object} map[string]interface{} "Successfully added item"
// @Failure 400 {object} map[string]string "Item name is required"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/{id}/items [put]
func AddItemToShopList(c *gin.Context) {
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

	// Parse request body
	var requestBody struct {
		ItemName  string `json:"item_name" binding:"required"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item name is required"})
		return
	}

	// Check if user is a member of the shoplist
	var member model.ShoplistMember
	err = db.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplistID, userID).First(&member).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Create new item
	newItem := model.ShoplistItem{
		ShopListID: shoplistID,
		ItemName:   requestBody.ItemName,
		BrandName:  requestBody.BrandName,
		ExtraInfo:  requestBody.ExtraInfo,
		IsBought:   false,
	}

	// Save to database
	if err := db.GetDB().Create(&newItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item"})
		return
	}

	// Return the created item
	c.JSON(http.StatusCreated, gin.H{
		"id":         newItem.ID,
		"item_name":  newItem.ItemName,
		"brand_name": newItem.BrandName,
		"extra_info": newItem.ExtraInfo,
		"is_bought":  newItem.IsBought,
	})
}

// RemoveItemFromShopList removes an item from a shoplist
// @Summary Remove an item from a shoplist
// @Description Removes a specific item from a shoplist. The user must be a member of the shoplist to remove items.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Param itemId path int true "Item ID"
// @Success 200 {object} gin.H "Successfully removed item"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/{id}/items/{itemId} [delete]
func RemoveItemFromShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse shoplist ID and item ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
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

	// Check if item exists and belongs to the shoplist
	var item model.ShoplistItem
	err = db.GetDB().Where("id = ? AND shop_list_id = ?", itemID, shoplistID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check item"})
		return
	}

	// Delete the item
	err = db.GetDB().Unscoped().Delete(&item).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// UpdateShoplistItem updates the bought status of an item
// @Summary Update an item in a shoplist
// @Description Updates the details of a specific item in a shoplist. At least one of the fields (item_name, brand_name, extra_info, is_bought) must be present in the request body. The user must be a member of the shoplist to update items.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Param itemId path int true "Item ID"
//
//	@Param request body struct {
//	    ItemName  *string `json:"item_name"`
//	    BrandName *string `json:"brand_name"`
//	    ExtraInfo *string `json:"extra_info"`
//	    IsBought  *bool   `json:"is_bought"`
//	} true "Item details"
//
// @Success 200 {object} map[string]interface{} "Successfully updated item"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 400 {object} map[string]string "Invalid item ID"
// @Failure 400 {object} map[string]string "At least one field must be present in the request body"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Router /shoplist/{id}/items/{itemId} [post]
func UpdateShoplistItem(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse shoplist ID and item ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shoplist ID"})
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Parse request body
	var requestBody struct {
		ItemName  *string `json:"item_name"`
		BrandName *string `json:"brand_name"`
		ExtraInfo *string `json:"extra_info"`
		IsBought  *bool   `json:"is_bought"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if request body is empty
	if requestBody.ItemName == nil && requestBody.BrandName == nil &&
		requestBody.ExtraInfo == nil && requestBody.IsBought == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body must include at least one of item_name, brand_name, extra_info or Is_bought."})
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

	// Check if item exists and belongs to the shoplist
	var item model.ShoplistItem
	err = db.GetDB().Where("id = ? AND shop_list_id = ?", itemID, shoplistID).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check item"})
		return
	}

	// Update only the fields that are provided in the request
	updates := make(map[string]interface{})
	if requestBody.ItemName != nil {
		updates["item_name"] = *requestBody.ItemName
	}
	if requestBody.BrandName != nil {
		updates["brand_name"] = *requestBody.BrandName
	}
	if requestBody.ExtraInfo != nil {
		updates["extra_info"] = *requestBody.ExtraInfo
	}
	if requestBody.IsBought != nil {
		updates["is_bought"] = *requestBody.IsBought
	}

	// Update the item
	err = db.GetDB().Model(&item).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	// Return the updated item
	c.JSON(http.StatusOK, gin.H{
		"id":         item.ID,
		"item_name":  item.ItemName,
		"brand_name": item.BrandName,
		"extra_info": item.ExtraInfo,
		"is_bought":  item.IsBought,
	})
}
