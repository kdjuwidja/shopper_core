package apiHandlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/logger"
	"netherrealmstudio.com/aishoppercore/m/model"
	"netherrealmstudio.com/aishoppercore/m/util"
)

func GetUserProfile(c *gin.Context) {
	userIDInterface, _ := c.Get("userID")
	userID := userIDInterface.(string)

	var user model.User

	db := db.GetDB()
	result := db.First(&user, "id = ?", userID)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func CreateOrUpdateUserProfile(c *gin.Context) {
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
	if !util.VerifyPostalCode(postalCode) {
		logger.Tracef("%s: Invalid postal code", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid postal code"})
		return
	}

	db := db.GetDB()

	user := model.User{
		ID:         userID,
		Nickname:   req.Nickname,
		PostalCode: postalCode,
	}

	// upsert user profile
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"postal_code", "nickname"}),
	}).Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Tracef("%s: User profile created or updated", userID)
	c.JSON(http.StatusOK, gin.H{})
}
