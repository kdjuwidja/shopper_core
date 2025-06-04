package apiHandlersshoplist

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/logger"

	"netherealmstudio.com/m/v2/apiHandlers"
	bizshoplist "netherealmstudio.com/m/v2/biz/shoplist"
)

// GetShoplistMembers returns the members of a shoplist
// @Summary Get shoplist members
// @Description Returns the members of a shoplist
// @Tags shoplist
// @Accept json
// @Produce json
// @Param id path int true "Shoplist ID"
// @Success 200 {object} map[string]interface{} "Successfully got shoplist members"
// @Failure 400 {object} map[string]string "Invalid shoplist ID"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /shoplist/{id}/members [GET]
func (h *ShoplistHandler) GetShoplistMembers(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("GetShoplistMembers: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	members, shoplistErr := h.shoplistBiz.GetShoplistMembers(c, userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		default:
			logger.Errorf("GetShoplistMembers: Failed to get shoplist members. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	// Transform members into the response format
	responseMembers := make([]map[string]string, 0)
	for _, member := range members {
		responseMembers = append(responseMembers, map[string]string{
			"id":       member.ID,
			"nickname": member.Nickname,
		})
	}

	h.responseFactory.CreateOKResponse(c, map[string]interface{}{
		"members": responseMembers,
	})
}

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
func (h *ShoplistHandler) LeaveShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("LeaveShopList: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	shoplistErr := h.shoplistBiz.LeaveShopList(c, userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistFailedToProcess:
			logger.Errorf("LeaveShopList: Failed to process shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		default:
			logger.Errorf("LeaveShopList: Failed to process shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}
		return
	}

	h.responseFactory.CreateOKResponse(c, nil)
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
func (h *ShoplistHandler) RequestShopListShareCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("RequestShopListShareCode: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	shareCode, shoplistErr := h.shoplistBiz.RequestShopListShareCode(c, userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotOwner:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotOwned)
		case bizshoplist.ShoplistFailedToProcess:
			logger.Errorf("RequestShopListShareCode: Failed to process shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		default:
			logger.Errorf("RequestShopListShareCode: Failed to generate share code. Error: %s", shoplistErr.Error())
		}

		return
	}

	respData := map[string]interface{}{
		"share_code": shareCode.Code,
		"expires_at": shareCode.Expiry.Format(time.RFC3339),
	}

	h.responseFactory.CreateOKResponse(c, respData)
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
func (h *ShoplistHandler) RevokeShopListShareCode(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("RevokeShopListShareCode: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Get shoplist ID from URL
	shoplistID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredParam, "id")
		return
	}

	shoplistErr := h.shoplistBiz.RevokeShopListShareCode(c, userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotOwner:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotOwned)
		case bizshoplist.ShoplistFailedToProcess:
			logger.Errorf("RevokeShopListShareCode: Failed to process shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		default:
			logger.Errorf("RevokeShopListShareCode: Failed to revoke share code. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}

		return
	}

	h.responseFactory.CreateOKResponse(c, nil)
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
func (h *ShoplistHandler) JoinShopList(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		logger.Errorf("JoinShopList: User ID is empty.")
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	// Parse request body
	var requestBody struct {
		ShareCode string `json:"share_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "share_code")
		return
	}

	shoplistErr := h.shoplistBiz.JoinShopList(c, userID, requestBody.ShareCode)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistFailedToProcess:
			logger.Errorf("JoinShopList: Failed to process shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		case bizshoplist.ShoplistNotFound:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		case bizshoplist.ShoplistNotMember:
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrShoplistNotFound)
		default:
			logger.Errorf("JoinShopList: Failed to join shoplist. Error: %s", shoplistErr.Error())
			h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		}

		return
	}

	h.responseFactory.CreateOKResponse(c, nil)
}
