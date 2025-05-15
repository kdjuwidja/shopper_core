package apihandlersuser

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"netherealmstudio.com/m/v2/apiHandlers"
	bizuser "netherealmstudio.com/m/v2/biz/user"
	dbmodel "netherealmstudio.com/m/v2/db"
)

// UserProfileHandler dependencies
type UserProfileHandler struct {
	userBiz         bizuser.UserBiz
	responseFactory apiHandlers.ResponseFactory
}

// Dependency Injection for UserProfileHandler
func InitializeUserProfileHandler(dbPool db.MySQLConnectionPool, responseFactory apiHandlers.ResponseFactory) *UserProfileHandler {
	return &UserProfileHandler{
		userBiz:         *bizuser.InitializeUserBiz(dbPool),
		responseFactory: responseFactory,
	}
}

func (h *UserProfileHandler) GetUserProfile(c *gin.Context) {
	userIDInterface, _ := c.Get("userID")
	userID := userIDInterface.(string)

	user, err := h.userBiz.GetUserProfile(userID)
	if err != nil {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrUserProfileNotFound)
		return
	}

	h.responseFactory.CreateOKResponse(c, user)
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
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInvalidRequestBody)
		return
	}

	if req.Nickname == "" {
		logger.Tracef("%s: Nickname is empty", userID)
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "nickname")
		return
	}

	if req.PostalCode == "" {
		logger.Tracef("%s: Postal code is empty", userID)
		h.responseFactory.CreateErrorResponsef(c, apiHandlers.ErrMissingRequiredField, "postal_code")
		return
	}

	postalCode := strings.ToUpper(req.PostalCode)
	if !bizuser.VerifyPostalCode(postalCode) {
		logger.Tracef("%s: Invalid postal code", userID)
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInvalidPostalCode)
		return
	}

	user := dbmodel.User{
		ID:         userID,
		Nickname:   req.Nickname,
		PostalCode: postalCode,
	}

	err := h.userBiz.CreateOrUpdateUserProfile(userID, &user)
	if err != nil {
		h.responseFactory.CreateErrorResponse(c, apiHandlers.ErrInternalServerError)
		return
	}

	logger.Tracef("%s: User profile created or updated", userID)
	h.responseFactory.CreateOKResponse(c, nil)
}
