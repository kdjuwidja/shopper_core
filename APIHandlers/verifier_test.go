package apiHandlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/kdjuwidja/aishoppercommon/logger"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestVerifyTokenWithValidTokenAndScopes(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	// Create a valid token with scopes
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "test-user-id",
		"scope": "read write",
	})
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read", "write"}, func(c *gin.Context) {
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

func TestVerifyTokenWithMissingScopes(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	// Create a token without scopes
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user-id",
	})
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"code":"GEN_00001","error":"Invalid or missing bearer token."}`, w.Body.String())
}

func TestVerifyTokenWithInsufficientScopes(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	// Create a token with only read scope
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "test-user-id",
		"scope": "read",
	})
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read", "write"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"code":"GEN_00005","error":"Missing scope: write"}`, w.Body.String())
}

func TestVerifyTokenMissingToken(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"code":"GEN_00001","error":"Invalid or missing bearer token."}`, w.Body.String())
}

func TestVerifyTokenInvalidBearerFormat(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"code":"GEN_00001","error":"Invalid or missing bearer token."}`, w.Body.String())
}

func TestVerifyTokenInvalidJWTToken(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"code":"GEN_00001","error":"Invalid or missing bearer token."}`, w.Body.String())
}

func TestVerifyTokenInvalidClaims(t *testing.T) {
	logger.SetServiceName("test")
	logger.SetLevel("trace")

	rf := ResponseFactory{}
	tokenVerifier := InitializeTokenVerifier(rf)

	// Save original JWT_SECRET and restore after test
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set test secret
	os.Setenv("JWT_SECRET", "test-secret")

	router := setupRouter()
	router.Use(tokenVerifier.VerifyToken([]string{"read"}, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"code":"GEN_00001","error":"Invalid or missing bearer token."}`, w.Body.String())
}
