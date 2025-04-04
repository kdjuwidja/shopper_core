package oauth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"netherrealmstudio.com/aishoppercore/m/logger"
)

type VerifyFunc gin.HandlerFunc

func VerifyToken(scopes []string, next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" || len(token) < 7 || token[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing bearer token"})
			c.Abort()
			return
		}

		token = token[7:]

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			logger.Warn("JWT secret not configured. Using default secret.")
			secret = "my-secret-key"
		}

		tokenObj, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != "HS256" {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !tokenObj.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		mapClaims, ok := tokenObj.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// TODO: Check if token has scope

		// Extract and set user ID
		if userID, exists := mapClaims["sub"].(string); exists {
			c.Set("userID", userID)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing user ID in token"})
			c.Abort()
			return
		}

		next(c)
	}
}
