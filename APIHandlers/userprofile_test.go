package apiHandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := db.InitializeTestDB(t)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Migrate the schema
	err = testDB.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return testDB
}

func teardownTestDB() {
	db.CloseTestDB()
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestVerifyPostalCode(t *testing.T) {
	tests := []struct {
		name        string
		postalCode  string
		expectValid bool
	}{
		{"Valid postal code", "A1B2C3", true},
		{"Invalid length", "A1B2C", false},
		{"Invalid format - numbers in letter positions", "123456", false},
		{"Invalid format - letters in number positions", "ABCDEF", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyPostalCode(tt.postalCode)
			assert.Equal(t, tt.expectValid, result)
		})
	}
}

func TestGetUserProfile(t *testing.T) {
	testDB := setupTestDB(t)
	defer teardownTestDB()

	// Create a test user
	testUser := model.User{
		ID:         "test-user-id",
		PostalCode: "A1B2C3",
	}
	testDB.Create(&testUser)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedUser   *model.User
	}{
		{
			name:           "Existing user",
			userID:         "test-user-id",
			expectedStatus: http.StatusOK,
			expectedUser:   &testUser,
		},
		{
			name:           "Non-existent user",
			userID:         "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectedUser:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", tt.userID)

			GetUserProfile(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response model.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, response.ID)
				assert.Equal(t, tt.expectedUser.PostalCode, response.PostalCode)
			}
		})
	}
}

func TestCreateOrUpdateUserProfile(t *testing.T) {
	testDB := setupTestDB(t)
	defer teardownTestDB()

	tests := []struct {
		name           string
		userID         string
		postalCode     string
		expectedStatus int
	}{
		{
			name:           "Valid postal code",
			userID:         "test-user-id",
			postalCode:     "A1B2C3",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid postal code format",
			userID:         "test-user-id",
			postalCode:     "123456",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty postal code",
			userID:         "test-user-id",
			postalCode:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", tt.userID)

			// Set request body
			reqBody := `{"postal_code": "` + tt.postalCode + `"}`
			c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateOrUpdateUserProfile(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response model.User
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, response.ID)
				assert.Equal(t, strings.ToUpper(tt.postalCode), response.PostalCode)

				// Verify user was saved in database
				var savedUser model.User
				err = testDB.First(&savedUser, "id = ?", tt.userID).Error
				assert.NoError(t, err)
				assert.Equal(t, strings.ToUpper(tt.postalCode), savedUser.PostalCode)
			}
		})
	}
}
