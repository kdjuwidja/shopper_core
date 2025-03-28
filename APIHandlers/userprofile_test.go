package apiHandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/model"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

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
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create a test user
	testUser := model.User{
		ID:         "test-user-id",
		Nickname:   "Test User",
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
				assert.Equal(t, tt.expectedUser.Nickname, response.Nickname)
				assert.Equal(t, tt.expectedUser.PostalCode, response.PostalCode)
			}
		})
	}
}

func TestCreateOrUpdateUserProfile(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Run create test
	t.Run("Create new user", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "new-user-id")

		// Set request body
		reqBody := `{"nickname": "Test User", "postal_code": "A1B2C3"}`
		c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateOrUpdateUserProfile(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify response is empty JSON
		assert.Equal(t, "{}", w.Body.String())

		// Verify user was saved in database
		var savedUser model.User
		err := testDB.First(&savedUser, "id = ?", "new-user-id").Error
		assert.NoError(t, err)
		assert.Equal(t, "A1B2C3", savedUser.PostalCode)
		assert.Equal(t, "Test User", savedUser.Nickname)
	})

	// Then run update tests
	// Create a test user for update tests
	testUser := model.User{
		ID:         "test-user-id",
		Nickname:   "Original Name",
		PostalCode: "A1B2C3",
	}
	testDB.Create(&testUser)

	updateTests := []struct {
		name           string
		userID         string
		nickname       string
		postalCode     string
		expectedStatus int
	}{
		{
			name:           "Update existing user",
			userID:         "test-user-id",
			nickname:       "Updated Name",
			postalCode:     "B2C3D4",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range updateTests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", tt.userID)

			// Set request body
			reqBody := `{"nickname": "` + tt.nickname + `", "postal_code": "` + tt.postalCode + `"}`
			c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateOrUpdateUserProfile(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Verify response is empty JSON
				assert.Equal(t, "{}", w.Body.String())

				// Verify user was saved in database
				var savedUser model.User
				err := testDB.First(&savedUser, "id = ?", tt.userID).Error
				assert.NoError(t, err)
				assert.Equal(t, strings.ToUpper(tt.postalCode), savedUser.PostalCode)
				assert.Equal(t, tt.nickname, savedUser.Nickname)
			}
		})
	}
}
