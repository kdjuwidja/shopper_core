package APIHandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
)

func UserProfile(c *gin.Context) {
	// Read or create user profile, since user ID is created in the auth service, we can be sure that the user exists.
	userIDInterface, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	userID := userIDInterface.(string)

	var user model.User

	db := db.GetDB()
	result := db.First(&user, "id = ?", userID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			user = model.User{
				ID:         userID,
				PostalCode: "",
			}
			if err := db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, user)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
