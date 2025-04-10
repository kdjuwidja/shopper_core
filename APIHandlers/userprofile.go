package apiHandlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/logger"
	bizuser "netherrealmstudio.com/aishoppercore/m/biz/user"
	"netherrealmstudio.com/aishoppercore/m/db"
)

// UserProfileHandler dependencies
type UserProfileHandler struct {
	userBiz bizuser.UserBiz
}

// Dependency Injection for UserProfileHandler
func InitializeUserProfileHandler(dbPool db.MySQLConnectionPool) *UserProfileHandler {
	return &UserProfileHandler{
		userBiz: *bizuser.InitializeUserBiz(dbPool),
	}
}

func (h *UserProfileHandler) GetUserProfile(c *gin.Context) {
	userIDInterface, _ := c.Get("userID")
	userID := userIDInterface.(string)

	user, err := h.userBiz.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserProfileHandler) CreateOrUpdateUserProfile(c *gin.Context) {
	userIDInterface, _ := c.Get("userID")
	userID := userIDInterface.(string)

	logger.Tracef("Creating or updating user profile for user %s", userID)
	logger.Debugf("Request body: %v", c.Request.Body)

	var req struct {
		Nickname   string `json:"nickname"`
		PostalCode string `json:"postal_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Nickname == "" {
		logger.Tracef("%s: Nickname is empty", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "nickname is required"})
		return
	}

	if req.PostalCode == "" {
		logger.Tracef("%s: Postal code is empty", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "postal_code is required"})
		return
	}

	postalCode := strings.ToUpper(req.PostalCode)
	if !bizuser.VerifyPostalCode(postalCode) {
		logger.Tracef("%s: Invalid postal code", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid postal code"})
		return
	}

	user := db.User{
		ID:         userID,
		Nickname:   req.Nickname,
		PostalCode: postalCode,
	}

	err := h.userBiz.CreateOrUpdateUserProfile(userID, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Tracef("%s: User profile created or updated", userID)
	c.JSON(http.StatusOK, gin.H{})
}
