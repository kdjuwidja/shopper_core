package apiHandlersshoplist

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"
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
		logger.Errorf("AddItemToShopList: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	// Parse request body
	var requestBody struct {
		ItemName  string `json:"item_name" binding:"required"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
		Thumbnail string `json:"thumbnail"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "item_name")
		return
	}

	newItem, shoplistErr := h.shoplistBiz.AddItemToShopList(c, userID, shoplistID, requestBody.ItemName, requestBody.BrandName, requestBody.ExtraInfo, requestBody.Thumbnail)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistItemNameEmpty:
			h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "item_name")
		case bizshoplist.ShoplistFailedToCreate:
			logger.Errorf("AddItemToShopList: Failed to add item. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		default:
			logger.Errorf("AddItemToShopList: Failed to add item. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	// Return the created item
	respData := map[string]interface{}{
		"id":         newItem.ID,
		"item_name":  newItem.ItemName,
		"brand_name": newItem.BrandName,
		"extra_info": newItem.ExtraInfo,
		"is_bought":  newItem.IsBought,
		"thumbnail":  newItem.Thumbnail,
	}

	h.responseFactory.CreateCreatedResponse(c, respData)
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
		logger.Errorf("RemoveItemFromShopList: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Parse shoplist ID and item ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "itemId")
		return
	}

	shoplistErr := h.shoplistBiz.RemoveItemFromShopList(c, userID, shoplistID, itemID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistItemNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistItemNotFound)
		case bizshoplist.ShoplistFailedToProcess:
			logger.Errorf("RemoveItemFromShopList: Failed to remove item. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		default:
			logger.Errorf("RemoveItemFromShopList: Failed to remove item. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	h.responseFactory.CreateOKResponse(c, nil)
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
		logger.Errorf("UpdateShoplistItem: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Parse shoplist ID and item ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "itemId")
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
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "item_name")
		return
	}

	// Check if request body is empty
	if requestBody.ItemName == nil && requestBody.BrandName == nil &&
		requestBody.ExtraInfo == nil && requestBody.IsBought == nil {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrMissingRequiredFieldUpdateShoplistItem)
		return
	}

	updatedItem, shoplistErr := h.shoplistBiz.UpdateShoplistItem(c, userID, shoplistID, itemID, requestBody.ItemName, requestBody.BrandName, requestBody.ExtraInfo, requestBody.IsBought)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistItemNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistItemNotFound)
		case bizshoplist.ShoplistFailedToProcess:
			logger.Errorf("UpdateShoplistItem: Failed to update item. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		default:
			logger.Errorf("UpdateShoplistItem: Failed to update item. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	// Return the updated item
	respData := map[string]interface{}{
		"id":         itemID,
		"item_name":  updatedItem.ItemName,
		"brand_name": updatedItem.BrandName,
		"extra_info": updatedItem.ExtraInfo,
		"is_bought":  updatedItem.IsBought,
	}

	h.responseFactory.CreateOKResponse(c, respData)
}
