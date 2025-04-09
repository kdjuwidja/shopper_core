package apiHandlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	bizshoplist "netherrealmstudio.com/aishoppercore/m/biz/shoplist"
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
func (h *ShoplistHandler) AddItemToShopList(c *gin.Context) {
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

	newItem, shoplistErr := h.shoplistBiz.AddItemToShopList(userID, shoplistID, requestBody.ItemName, requestBody.BrandName, requestBody.ExtraInfo)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		case bizshoplist.ShoplistItemNameEmpty:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Item name is required."})
		case bizshoplist.ShoplistFailedToCreate:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item."})
		}
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
func (h *ShoplistHandler) RemoveItemFromShopList(c *gin.Context) {
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

	shoplistErr := h.shoplistBiz.RemoveItemFromShopList(userID, shoplistID, itemID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found."})
		case bizshoplist.ShoplistNotMember:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found."})
		case bizshoplist.ShoplistItemNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found."})
		case bizshoplist.ShoplistFailedToProcess:
			c.JSON(http.StatusInternalServerError, gin.H{"error": shoplistErr.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item."})
		}
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
func (h *ShoplistHandler) UpdateShoplistItem(c *gin.Context) {
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

	updatedItem, shoplistErr := h.shoplistBiz.UpdateShoplistItem(userID, shoplistID, itemID, requestBody.ItemName, requestBody.BrandName, requestBody.ExtraInfo, requestBody.IsBought)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found."})
		case bizshoplist.ShoplistNotMember:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found."})
		case bizshoplist.ShoplistItemNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found."})
		case bizshoplist.ShoplistFailedToProcess:
			c.JSON(http.StatusInternalServerError, gin.H{"error": shoplistErr.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item."})
		}
		return
	}

	// Return the updated item
	c.JSON(http.StatusOK, gin.H{
		"id":         itemID,
		"item_name":  updatedItem.ItemName,
		"brand_name": updatedItem.BrandName,
		"extra_info": updatedItem.ExtraInfo,
		"is_bought":  updatedItem.IsBought,
	})
}
