package APIHandlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
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

func verifyPostalCode(postalCode string) bool {
	if len(postalCode) != 6 {
		return false
	}

	// Check odd positions (1,3,5) are letters
	for i := 0; i < 6; i += 2 {
		if postalCode[i] < 'A' || postalCode[i] > 'Z' {
			return false
		}
	}

	// Check even positions (2,4,6) are numbers
	for i := 1; i < 6; i += 2 {
		if postalCode[i] < '0' || postalCode[i] > '9' {
			return false
		}
	}

	return true
}

func CreateOrUpdateUserProfile(c *gin.Context) {
	userIDInterface, _ := c.Get("userID")
	userID := userIDInterface.(string)

	var req struct {
		PostalCode string `json:"postal_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.PostalCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postal_code is required"})
		return
	}

	postalCode := strings.ToUpper(req.PostalCode)
	if !verifyPostalCode(postalCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid postal code"})
		return
	}

	db := db.GetDB()

	user := model.User{
		ID:         userID,
		PostalCode: postalCode,
	}

	// upsert user profile
	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
