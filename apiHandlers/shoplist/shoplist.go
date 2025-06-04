package apiHandlersshoplist

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/logger"

	"netherealmstudio.com/m/v2/apiHandlers"
	bizmodels "netherealmstudio.com/m/v2/biz"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"
)

// CreateShoplist creates a new shoplist
// @Summary Create a new shoplist
// @Description Creates a new shoplist with the specified name. The authenticated user becomes the owner of the shoplist.
// @Tags shoplist
// @Accept json
// @Produce json
// @Param request body struct{Name string} true "Shoplist details"
// @Success 201 {object} map[string]interface{} "Successfully created shoplist"
// @Failure 400 {object} map[string]string "Name is required"
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to create shoplist"
// @Router /shoplist [put]
func (h *ShoplistHandler) CreateShoplist(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("CreateShoplist: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInvalidRequestBody)
		return
	}

	if req.Name == "" {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "name")
		return
	}

	err := h.shoplistBiz.CreateShoplist(c, userID, req.Name)
	if err != nil {
		logger.Errorf("CreateShoplist: Failed to create shoplist. Error: %s", err.Error())
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	h.responseFactory.CreateCreatedResponse(c, nil)
}

// UpdateShoplist updates a shoplist's name
// @Summary Update a shoplist's name
// @Description Updates the name of a specific shoplist by its ID
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Param name body string true "Name of the shoplist"
// @Success 200 {object} gin.H
// @Failure 400 {object} map[string]string "Name is required"
// @Failure 403 {object} map[string]string "Only the owner can update this shoplist"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to process shoplist data"
// @Router /shoplist/{id} [post]
func (h *ShoplistHandler) UpdateShoplist(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("UpdateShoplist: User ID is empty.")
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
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "name")
		return
	}

	shoplistErr := h.shoplistBiz.UpdateShoplist(c, userID, shoplistID, requestBody.Name)
	if shoplistErr != nil {
		if shoplistErr.ErrCode == bizshoplist.ShoplistNotFound {
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		} else if shoplistErr.ErrCode == bizshoplist.ShoplistNotOwned {
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotOwned)
		} else {
			logger.Errorf("UpdateShoplist: Failed to update shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	h.responseFactory.CreateOKResponse(c, nil)
}

// GetAllShoplistItems retrieves all shoplist items for a user
// @Summary Get all shoplist items for a user
// @Description Retrieves all shoplist items for a user
// @Tags shoplist
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Successfully retrieved shoplist items"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to fetch shoplist items"
// @Router /shoplist/items [get]
func (h *ShoplistHandler) GetAllShoplistAndItemsForUser(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("GetAllShoplistItems: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	shoplists, err := h.shoplistBiz.GetAllShoplistAndItemsForUser(c.Request.Context(), userID)
	if err != nil {
		logger.Errorf("GetAllShoplistItems: Failed to get shoplist items. Error: %s", err.Error())
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	shoplistItems := make([]bizmodels.ShoplistItem, 0)
	for _, shoplist := range shoplists {
		shoplistItems = append(shoplistItems, shoplist.Items...)
	}

	flyers, matchErr := h.matchBiz.MatchShoplistItemsWithFlyer(c.Request.Context(), shoplistItems)
	if matchErr != nil {
		logger.Errorf("Failed to match shoplist items with flyers: %v", matchErr)
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Transform the response to match the desired format
	response := make([]ShoplistResponse, 0)
	for _, shoplist := range shoplists {
		shoplistResp := ShoplistResponse{
			ID:   shoplist.ID,
			Name: shoplist.Name,
			Owner: OwnerResponse{
				ID:       shoplist.OwnerID,
				Nickname: shoplist.OwnerNickname,
			},
			Items: make([]ItemResponse, 0),
		}

		// Add items for this shoplist
		for _, item := range shoplist.Items {
			flyerResp := make([]FlyerResponse, 0)

			flyersForItem, exists := flyers[item.ID]
			if exists {
				for _, flyer := range flyersForItem {
					flyerResp = append(flyerResp, FlyerResponse{
						Store:         flyer.Store,
						Brand:         flyer.Brand,
						StartDate:     flyer.StartDateTime,
						EndDate:       flyer.EndDateTime,
						ProductName:   flyer.ProductName,
						Description:   flyer.Description,
						OriginalPrice: flyer.OriginalPrice,
						PrePriceText:  flyer.PrePriceText,
						PriceText:     flyer.PriceText,
						PostPriceText: flyer.PostPriceText,
					})
				}
			}

			shoplistResp.Items = append(shoplistResp.Items, ItemResponse{
				ID:        item.ID,
				Name:      item.ItemName,
				BrandName: item.BrandName,
				ExtraInfo: item.ExtraInfo,
				IsBought:  item.IsBought,
				Flyer:     flyerResp,
			})
		}

		response = append(response, shoplistResp)
	}

	h.responseFactory.CreateOKResponse(c, map[string]interface{}{"shoplists": response})
}

func (h *ShoplistHandler) GetShoplistAndItemsForUserByShoplistID(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("GetShoplistAndItemsForUserByShoplistID: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	shoplistID, perr := strconv.Atoi(c.Param("id"))
	if perr != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	shoplist, err := h.shoplistBiz.GetShoplistAndItems(c.Request.Context(), userID, shoplistID)
	if err != nil {
		logger.Errorf("GetShoplistAndItemsForUserByShoplistID: Failed to get shoplist. Error: %s", err.Error())
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		return
	}

	// Transform the shoplist into the response format
	shoplistResp := ShoplistResponse{
		ID:   shoplist.ID,
		Name: shoplist.Name,
		Owner: OwnerResponse{
			ID:       shoplist.OwnerID,
			Nickname: shoplist.OwnerNickname,
		},
		Items: make([]ItemResponse, 0),
	}

	if len(shoplist.Items) > 0 {
		flyers, matchErr := h.matchBiz.MatchShoplistItemsWithFlyer(c.Request.Context(), shoplist.Items)
		if matchErr != nil {
			logger.Errorf("GetShoplistAndItemsForUserByShoplistID: Failed to match shoplist items with flyers. Error: %s", matchErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
			return
		}

		// Add items for this shoplist
		for _, item := range shoplist.Items {
			flyerResp := make([]FlyerResponse, 0)

			flyersForItem, exists := flyers[item.ID]
			if exists {
				for _, flyer := range flyersForItem {
					flyerResp = append(flyerResp, FlyerResponse{
						Store:         flyer.Store,
						Brand:         flyer.Brand,
						StartDate:     flyer.StartDateTime,
						EndDate:       flyer.EndDateTime,
						ProductName:   flyer.ProductName,
						Description:   flyer.Description,
						OriginalPrice: flyer.OriginalPrice,
						PrePriceText:  flyer.PrePriceText,
						PriceText:     flyer.PriceText,
						PostPriceText: flyer.PostPriceText,
					})
				}
			}

			shoplistResp.Items = append(shoplistResp.Items, ItemResponse{
				ID:        item.ID,
				Name:      item.ItemName,
				BrandName: item.BrandName,
				ExtraInfo: item.ExtraInfo,
				IsBought:  item.IsBought,
				Flyer:     flyerResp,
			})
		}
	}

	h.responseFactory.CreateOKResponse(c, shoplistResp)
}
