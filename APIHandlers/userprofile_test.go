package apiHandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/db"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func setUpTestEnv(t *testing.T) (*UserProfileHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	userProfileHandler := InitializeUserProfileHandler(*testDBConn)
	return userProfileHandler, testDBConn
}

func TestGetUserProfileWithExistingUser(t *testing.T) {
	userProfileHandler, testDBConn := setUpTestEnv(t)

	// Create a test user
	testUser := db.User{
		ID:         "test-user-id",
		Nickname:   "Test User",
		PostalCode: "A1B2C3",
	}
	testDBConn.GetDB().Create(&testUser)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "test-user-id")

	userProfileHandler.GetUserProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response db.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, response.ID)
	assert.Equal(t, testUser.Nickname, response.Nickname)
	assert.Equal(t, testUser.PostalCode, response.PostalCode)
}

func TestGetUserProfileWithNonExistentUser(t *testing.T) {
	userProfileHandler, _ := setUpTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "non-existent-id")

	userProfileHandler.GetUserProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateUserProfile(t *testing.T) {
	userProfileHandler, testDBConn := setUpTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "new-user-id")

	// Set request body
	reqBody := `{"nickname": "Test User", "postal_code": "A1B2C3"}`
	c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	userProfileHandler.CreateOrUpdateUserProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response is empty JSON
	assert.Equal(t, "{}", w.Body.String())

	// Verify user was saved in database
	var savedUser db.User
	err := testDBConn.GetDB().First(&savedUser, "id = ?", "new-user-id").Error
	assert.NoError(t, err)
	assert.Equal(t, "A1B2C3", savedUser.PostalCode)
	assert.Equal(t, "Test User", savedUser.Nickname)
}

func TestUpdateUserProfile(t *testing.T) {
	userProfileHandler, testDBConn := setUpTestEnv(t)

	// Create a test user for update tests
	testUser := db.User{
		ID:         "test-user-id",
		Nickname:   "Original Name",
		PostalCode: "A1B2C3",
	}
	testDBConn.GetDB().Create(&testUser)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "test-user-id")

	// Set request body
	reqBody := `{"nickname": "Updated Name", "postal_code": "B2C3D4"}`
	c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	userProfileHandler.CreateOrUpdateUserProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response is empty JSON
	assert.Equal(t, "{}", w.Body.String())

	// Verify user was updated in database
	var savedUser db.User
	err := testDBConn.GetDB().First(&savedUser, "id = ?", "test-user-id").Error
	assert.NoError(t, err)
	assert.Equal(t, strings.ToUpper("B2C3D4"), savedUser.PostalCode)
	assert.Equal(t, "Updated Name", savedUser.Nickname)
}
