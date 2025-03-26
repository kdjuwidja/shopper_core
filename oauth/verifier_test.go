package oauth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestVerifyToken(t *testing.T) {
	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid or missing bearer token"}`,
		},
		{
			name:           "Invalid bearer format",
			token:          "InvalidFormat token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid or missing bearer token"}`,
		},
		{
			name:           "Invalid JWT token",
			token:          "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
		{
			name:           "Token with invalid claims",
			token:          "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter()
			router.Use(VerifyToken([]string{}, func(c *gin.Context) {
				c.Status(http.StatusOK)
			}))

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestVerifyTokenWithValidToken(t *testing.T) {
	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	// Create a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user-id",
	})
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	router := setupRouter()
	router.Use(VerifyToken([]string{}, func(c *gin.Context) {
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "test-user-id", userID)
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVerifyTokenWithMissingJWTSecret(t *testing.T) {
	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Clear JWT_SECRET
	os.Setenv("JWT_SECRET", "")

	router := setupRouter()
	router.Use(VerifyToken([]string{}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer some.token.here")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"JWT secret not configured"}`, w.Body.String())
}
