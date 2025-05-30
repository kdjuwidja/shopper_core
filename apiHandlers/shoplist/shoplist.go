package apiHandlersshoplist

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/logger"

	"netherealmstudio.com/m/v2/apiHandlers"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"
)

type ShoplistHandler struct {
	shoplistBiz     *bizshoplist.ShoplistBiz
	responseFactory apiHandlers.ResponseFactory
}

// Dependency Injection for ShoplistHandler
func InitializeShoplistHandler(dbPool db.MySQLConnectionPool, responseFactory apiHandlers.ResponseFactory) *ShoplistHandler {
	return &ShoplistHandler{
		shoplistBiz:     bizshoplist.InitializeShoplistBiz(dbPool),
		responseFactory: responseFactory,
	}
}

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

	err := h.shoplistBiz.CreateShoplist(userID, req.Name)
	if err != nil {
		logger.Errorf("CreateShoplist: Failed to create shoplist. Error: %s", err.Error())
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	h.responseFactory.CreateCreatedResponse(c, nil)
}

// GetAllShoplists retrieves all shoplists for the authenticated user
// @Summary Get all shoplists
// @Description Retrieves all shoplists where the authenticated user is a member, including the owner's information.
// @Tags shoplist
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Successfully retrieved shoplists"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Failed to fetch shoplists"
// @Failure 500 {object} map[string]string "Failed to process shoplist data"
// @Router /shoplist [get]
func (h *ShoplistHandler) GetAllShoplists(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("GetAllShoplists: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	shoplistData, err := h.shoplistBiz.GetAllShoplists(userID)
	if err != nil {
		logger.Errorf("GetAllShoplists: Failed to get all shoplists. Error: %s", err.Error())
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Transform the response to only include owner nickname
	type ShoplistResponse struct {
		ID    int    `json:"id" gorm:"column:id"`
		Name  string `json:"name" gorm:"column:name"`
		Owner struct {
			Owner_id string `json:"id" gorm:"column:owner_id"`
			Nickname string `json:"nickname" gorm:"column:owner_nickname"`
		} `json:"owner"`
	}

	response := make([]ShoplistResponse, 0)
	for _, shoplist := range shoplistData {
		response = append(response, ShoplistResponse{
			ID:   shoplist.ID,
			Name: shoplist.Name,
			Owner: struct {
				Owner_id string `json:"id" gorm:"column:owner_id"`
				Nickname string `json:"nickname" gorm:"column:owner_nickname"`
			}{
				Owner_id: shoplist.OwnerID,
				Nickname: shoplist.OwnerNickname,
			},
		})
	}

	h.responseFactory.CreateOKResponse(c, map[string]interface{}{"shoplists": response})
}

// GetShoplist retrieves a specific shoplist and its items
// @Summary Get a specific shoplist
// @Description Retrieves a shoplist by its ID along with its items and members
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} ResponseModel
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Failed to fetch shoplist"
// @Failure 500 {object} map[string]string "Failed to process shoplist data"
// @Router /shoplist/{id} [get]
func (h *ShoplistHandler) GetShoplist(c *gin.Context) {
	// Get shoplist ID from URL parameters
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("GetShoplist: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	shoplistData, shoplistItems, shoplistMembers, shoplistErr := h.shoplistBiz.GetShoplist(userID, shoplistID)
	if shoplistErr != nil {
		if shoplistErr.ErrCode == bizshoplist.ShoplistNotFound {
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		} else {
			logger.Errorf("GetShoplist: Failed to get shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	// massage response data into expected format
	if shoplistData == nil {
		logger.Errorf("GetShoplist: shoplistData == nil, shoplistId: %d", shoplistID)
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	type ResponseModel struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Owner struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"owner"`
		Members []struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"members"`
		Items []struct {
			ID        int    `json:"id"`
			ItemName  string `json:"item_name"`
			BrandName string `json:"brand_name"`
			ExtraInfo string `json:"extra_info"`
			IsBought  bool   `json:"is_bought"`
			Thumbnail string `json:"thumbnail"`
		} `json:"items"`
	}

	responseMembers := make([]struct {
		ID       string "json:\"id\""
		Nickname string "json:\"nickname\""
	}, 0)
	for _, member := range shoplistMembers {
		responseMembers = append(responseMembers, struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		}{ID: member.ID, Nickname: member.Nickname})
	}

	responseItems := make([]struct {
		ID        int    `json:"id"`
		ItemName  string `json:"item_name"`
		BrandName string `json:"brand_name"`
		ExtraInfo string `json:"extra_info"`
		IsBought  bool   `json:"is_bought"`
		Thumbnail string `json:"thumbnail"`
	}, 0)
	for _, item := range shoplistItems {
		responseItems = append(responseItems, struct {
			ID        int    `json:"id"`
			ItemName  string `json:"item_name"`
			BrandName string `json:"brand_name"`
			ExtraInfo string `json:"extra_info"`
			IsBought  bool   `json:"is_bought"`
			Thumbnail string `json:"thumbnail"`
		}{
			ID:        item.ID,
			ItemName:  item.ItemName,
			BrandName: item.BrandName,
			ExtraInfo: item.ExtraInfo,
			IsBought:  item.IsBought,
			Thumbnail: item.Thumbnail,
		})
	}

	response := ResponseModel{
		ID:   shoplistData.ID,
		Name: shoplistData.Name,
		Owner: struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		}{
			ID:       shoplistData.OwnerID,
			Nickname: shoplistData.OwnerNickname,
		},
		Members: responseMembers,
		Items:   responseItems,
	}
	h.responseFactory.CreateOKResponse(c, response)
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

	shoplistErr := h.shoplistBiz.UpdateShoplist(userID, shoplistID, requestBody.Name)
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

	// Transform the response to match the desired format
	response := make([]ShoplistResponse, 0)
	for _, shoplist := range shoplists {
		shoplistResp := ShoplistResponse{
			ID:   shoplist.ID,
			Name: shoplist.Name,
			Owner: struct {
				ID       string `json:"id"`
				Nickname string `json:"nickname"`
			}{
				ID:       shoplist.OwnerID,
				Nickname: shoplist.OwnerNickname,
			},
			Items: make([]struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				BrandName string `json:"brand_name"`
				ExtraInfo string `json:"extra_info"`
				IsBought  bool   `json:"is_bought"`
			}, 0),
		}

		// Add items for this shoplist
		for _, item := range shoplist.Items {
			shoplistResp.Items = append(shoplistResp.Items, struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				BrandName string `json:"brand_name"`
				ExtraInfo string `json:"extra_info"`
				IsBought  bool   `json:"is_bought"`
			}{
				ID:        item.ID,
				Name:      item.ItemName,
				BrandName: item.BrandName,
				ExtraInfo: item.ExtraInfo,
				IsBought:  item.IsBought,
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
		Owner: struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		}{
			ID:       shoplist.OwnerID,
			Nickname: shoplist.OwnerNickname,
		},
		Items: make([]struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			BrandName string `json:"brand_name"`
			ExtraInfo string `json:"extra_info"`
			IsBought  bool   `json:"is_bought"`
		}, 0),
	}

	// Add items for this shoplist
	for _, item := range shoplist.Items {
		shoplistResp.Items = append(shoplistResp.Items, struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			BrandName string `json:"brand_name"`
			ExtraInfo string `json:"extra_info"`
			IsBought  bool   `json:"is_bought"`
		}{
			ID:        item.ID,
			Name:      item.ItemName,
			BrandName: item.BrandName,
			ExtraInfo: item.ExtraInfo,
			IsBought:  item.IsBought,
		})
	}

	h.responseFactory.CreateOKResponse(c, shoplistResp)
}
