package apiHandlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	bizshoplist "netherrealmstudio.com/aishoppercore/m/biz/shoplist"
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
func (h *ShoplistHandler) LeaveShopList(c *gin.Context) {
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

	shoplistErr := h.shoplistBiz.LeaveShopList(userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		case bizshoplist.ShoplistNotMember:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		case bizshoplist.ShoplistFailedToProcess:
			c.JSON(http.StatusInternalServerError, gin.H{"error": shoplistErr.Message})
		}
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
func (h *ShoplistHandler) RequestShopListShareCode(c *gin.Context) {
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

	shareCode, shoplistErr := h.shoplistBiz.RequestShopListShareCode(userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		case bizshoplist.ShoplistNotMember:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		case bizshoplist.ShoplistNotOwner:
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can generate share codes"})
			return
		case bizshoplist.ShoplistFailedToProcess:
			c.JSON(http.StatusInternalServerError, gin.H{"error": shoplistErr.Message})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate share code"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"share_code": shareCode.Code,
		"expires_at": shareCode.Expiry.Format(time.RFC3339),
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
func (h *ShoplistHandler) RevokeShopListShareCode(c *gin.Context) {
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

	shoplistErr := h.shoplistBiz.RevokeShopListShareCode(userID, shoplistID)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		case bizshoplist.ShoplistNotMember:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		case bizshoplist.ShoplistNotOwner:
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can revoke share codes"})
		case bizshoplist.ShoplistFailedToProcess:
			c.JSON(http.StatusInternalServerError, gin.H{"error": shoplistErr.Message})
		}
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
func (h *ShoplistHandler) JoinShopList(c *gin.Context) {
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

	shoplistErr := h.shoplistBiz.JoinShopList(userID, requestBody.ShareCode)
	if shoplistErr != nil {
		switch shoplistErr.ErrCode {
		case bizshoplist.ShoplistFailedToProcess:
			c.JSON(http.StatusInternalServerError, gin.H{"error": shoplistErr.Message})
		case bizshoplist.ShoplistNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		case bizshoplist.ShoplistNotMember:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}
