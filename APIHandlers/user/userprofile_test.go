package apihandlersuser

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/apiHandlers"
	dbmodel "netherrealmstudio.com/aishoppercore/m/db"
)

func TestGetUserProfileWithExistingUser(t *testing.T) {
	userProfileHandler, testDBConn := setUpTestEnv(t)

	// Create a test user
	testUser := dbmodel.User{
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
	var response dbmodel.User
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

	var response apiHandlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "USR_00002", response.Code)
	assert.Equal(t, "User profile not found.", response.Error)
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
	var savedUser dbmodel.User
	err := testDBConn.GetDB().First(&savedUser, "id = ?", "new-user-id").Error
	assert.NoError(t, err)
	assert.Equal(t, "A1B2C3", savedUser.PostalCode)
	assert.Equal(t, "Test User", savedUser.Nickname)
}

func TestUpdateUserProfile(t *testing.T) {
	userProfileHandler, testDBConn := setUpTestEnv(t)

	// Create a test user for update tests
	testUser := dbmodel.User{
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
	var savedUser dbmodel.User
	err := testDBConn.GetDB().First(&savedUser, "id = ?", "test-user-id").Error
	assert.NoError(t, err)
	assert.Equal(t, strings.ToUpper("B2C3D4"), savedUser.PostalCode)
	assert.Equal(t, "Updated Name", savedUser.Nickname)
}

func TestCreateUserProfileWithInvalidPostalCode(t *testing.T) {
	userProfileHandler, _ := setUpTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "new-user-id")

	// Set request body with invalid postal code
	reqBody := `{"nickname": "Test User", "postal_code": "123456"}`
	c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	userProfileHandler.CreateOrUpdateUserProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apiHandlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "USR_00001", response.Code)
	assert.Equal(t, "Invalid postal code.", response.Error)
}

func TestCreateUserProfileWithMissingNickname(t *testing.T) {
	userProfileHandler, _ := setUpTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "new-user-id")

	// Set request body with missing nickname
	reqBody := `{"postal_code": "A1B2C3"}`
	c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	userProfileHandler.CreateOrUpdateUserProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apiHandlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "GEN_00003", response.Code)
	assert.Equal(t, "Missing field in body: nickname", response.Error)
}

func TestCreateUserProfileWithMissingPostalCode(t *testing.T) {
	userProfileHandler, _ := setUpTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "new-user-id")

	// Set request body with missing postal code
	reqBody := `{"nickname": "Test User"}`
	c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	userProfileHandler.CreateOrUpdateUserProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apiHandlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "GEN_00003", response.Code)
	assert.Equal(t, "Missing field in body: postal_code", response.Error)
}

func TestCreateUserProfileWithInvalidJSON(t *testing.T) {
	userProfileHandler, _ := setUpTestEnv(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "new-user-id")

	// Set invalid JSON request body
	reqBody := `{"nickname": "Test User", "postal_code": "A1B2C3", invalid_json}`
	c.Request = httptest.NewRequest("POST", "/user", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	userProfileHandler.CreateOrUpdateUserProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apiHandlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "GEN_00002", response.Code)
	assert.Equal(t, "Invalid request body", response.Error)
}
