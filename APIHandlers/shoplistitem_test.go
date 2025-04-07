package apiHandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/model"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func TestAddItemToShopListOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items", AddItemToShopList)

	// Create request
	requestBody := map[string]interface{}{
		"item_name":  "Test Item",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist/1/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	AddItemToShopList(c)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"id":         float64(1),
		"item_name":  "Test Item",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
		"is_bought":  false,
	}, response)

	// Verify database
	var item model.ShoplistItem
	err = testDB.First(&item, response["id"]).Error
	assert.NoError(t, err)
	assert.Equal(t, requestBody["item_name"], item.ItemName)
	assert.Equal(t, requestBody["brand_name"], item.BrandName)
	assert.Equal(t, requestBody["extra_info"], item.ExtraInfo)
	assert.False(t, item.IsBought)
	assert.Equal(t, 1, item.ShopListID)
}

func TestAddItemToShopListMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items", AddItemToShopList)

	// Create request
	requestBody := map[string]interface{}{
		"item_name":  "Another Item",
		"brand_name": "Another Brand",
		"extra_info": "Another Info",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist/1/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	AddItemToShopList(c)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"id":         float64(1),
		"item_name":  "Another Item",
		"brand_name": "Another Brand",
		"extra_info": "Another Info",
		"is_bought":  false,
	}, response)

	// Verify database
	var item model.ShoplistItem
	err = testDB.First(&item, response["id"]).Error
	assert.NoError(t, err)
	assert.Equal(t, requestBody["item_name"], item.ItemName)
	assert.Equal(t, requestBody["brand_name"], item.BrandName)
	assert.Equal(t, requestBody["extra_info"], item.ExtraInfo)
	assert.False(t, item.IsBought)
	assert.Equal(t, 1, item.ShopListID)
}

func TestAddItemToShopListNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items", AddItemToShopList)

	// Create request
	requestBody := map[string]interface{}{
		"item_name":  "Test Item",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist/1/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	AddItemToShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestAddItemToShopListNonExistentShoplist(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items", AddItemToShopList)

	// Create request
	requestBody := map[string]interface{}{
		"item_name":  "Test Item",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist/99999/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	AddItemToShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestAddItemToShopListMissingRequiredFields(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items", AddItemToShopList)

	// Create request
	requestBody := map[string]interface{}{
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist/1/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	AddItemToShopList(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Item name is required",
	}, response)
}

func TestAddItemToShopListEmptyName(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items", AddItemToShopList)

	// Create request
	requestBody := map[string]interface{}{
		"item_name":  "",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist/1/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	AddItemToShopList(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Item name is required",
	}, response)
}

func TestRemoveItemFromShopListOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

	// Create request
	req, _ := http.NewRequest("DELETE", "/shoplist/1/items/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	RemoveItemFromShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify item is deleted from database
	var deletedItem model.ShoplistItem
	err = testDB.First(&deletedItem, 1).Error
	assert.Error(t, err, "Item should be deleted from database")
	assert.True(t, err.Error() == "record not found", "Error should be record not found")
}

func TestRemoveItemFromShopListMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

	// Create request
	req, _ := http.NewRequest("DELETE", "/shoplist/1/items/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	RemoveItemFromShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify item is deleted from database
	var deletedItem model.ShoplistItem
	err = testDB.First(&deletedItem, 1).Error
	assert.Error(t, err, "Item should be deleted from database")
	assert.True(t, err.Error() == "record not found", "Error should be record not found")
}

func TestRemoveItemFromShopListNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

	// Create request
	req, _ := http.NewRequest("DELETE", "/shoplist/1/items/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	RemoveItemFromShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)

	// Verify item still exists in database
	var existingItem model.ShoplistItem
	err = testDB.First(&existingItem, 1).Error
	assert.NoError(t, err, "Item should still exist in database")
}

func TestRemoveItemFromShopListNonExistentShoplist(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

	// Create request
	req, _ := http.NewRequest("DELETE", "/shoplist/99999/items/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "99999"},
		{Key: "itemId", Value: "1"},
	}

	RemoveItemFromShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestRemoveItemFromShopListNonExistentItem(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

	// Create request
	req, _ := http.NewRequest("DELETE", "/shoplist/1/items/99999", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "99999"},
	}

	RemoveItemFromShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Item not found",
	}, response)
}

func TestRemoveItemFromShopListDifferentShoplist(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplists
	shoplist1 := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist 1",
	}
	shoplist2 := model.Shoplist{
		ID:      2,
		OwnerID: owner.ID,
		Name:    "Test Shoplist 2",
	}
	err = testDB.Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplist2).Error
	assert.NoError(t, err)

	// Add owner as member to both shoplists
	ownerMember1 := model.ShoplistMember{
		ID:         1,
		ShopListID: shoplist1.ID,
		MemberID:   owner.ID,
	}
	ownerMember2 := model.ShoplistMember{
		ID:         2,
		ShopListID: shoplist2.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember1).Error
	assert.NoError(t, err)
	err = testDB.Create(&ownerMember2).Error
	assert.NoError(t, err)

	// Add test item to second shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: shoplist2.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

	// Create request
	req, _ := http.NewRequest("DELETE", "/shoplist/1/items/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	RemoveItemFromShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Item not found",
	}, response)

	// Verify item still exists in database
	var existingItem model.ShoplistItem
	err = testDB.First(&existingItem, 1).Error
	assert.NoError(t, err, "Item should still exist in database")
}

func TestUpdateShoplistItemOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request
	requestBody := map[string]interface{}{
		"is_bought": true,
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/1/items/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"id":         float64(1),
		"item_name":  "Test Item",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
		"is_bought":  true,
	}, response)

	// Verify database
	var updatedItem model.ShoplistItem
	err = testDB.First(&updatedItem, 1).Error
	assert.NoError(t, err)
	assert.True(t, updatedItem.IsBought)
}

func TestUpdateShoplistItemMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   true,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request
	requestBody := map[string]interface{}{
		"is_bought": false,
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/1/items/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"id":         float64(1),
		"item_name":  "Test Item",
		"brand_name": "Test Brand",
		"extra_info": "Test Info",
		"is_bought":  false,
	}, response)

	// Verify database
	var updatedItem model.ShoplistItem
	err = testDB.First(&updatedItem, 1).Error
	assert.NoError(t, err)
	assert.False(t, updatedItem.IsBought)
}

func TestUpdateShoplistItemNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request
	requestBody := map[string]interface{}{
		"is_bought": true,
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/1/items/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)

	// Verify item is unchanged in database
	var existingItem model.ShoplistItem
	err = testDB.First(&existingItem, 1).Error
	assert.NoError(t, err)
	assert.False(t, existingItem.IsBought)
}

func TestUpdateShoplistItemNonExistentShoplist(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request
	requestBody := map[string]interface{}{
		"is_bought": true,
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/99999/items/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "99999"},
		{Key: "itemId", Value: "1"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestUpdateShoplistItemNonExistentItem(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request
	requestBody := map[string]interface{}{
		"is_bought": true,
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/1/items/99999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "99999"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Item not found",
	}, response)
}

func TestUpdateShoplistItemDifferentShoplist(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplists
	shoplist1 := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist 1",
	}
	shoplist2 := model.Shoplist{
		ID:      2,
		OwnerID: owner.ID,
		Name:    "Test Shoplist 2",
	}
	err = testDB.Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplist2).Error
	assert.NoError(t, err)

	// Add owner as member to both shoplists
	ownerMember1 := model.ShoplistMember{
		ID:         1,
		ShopListID: shoplist1.ID,
		MemberID:   owner.ID,
	}
	ownerMember2 := model.ShoplistMember{
		ID:         2,
		ShopListID: shoplist2.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember1).Error
	assert.NoError(t, err)
	err = testDB.Create(&ownerMember2).Error
	assert.NoError(t, err)

	// Add test item to second shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: shoplist2.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request
	requestBody := map[string]interface{}{
		"is_bought": true,
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/1/items/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Item not found",
	}, response)

	// Verify item still exists in database
	var existingItem model.ShoplistItem
	err = testDB.First(&existingItem, 1).Error
	assert.NoError(t, err)
	assert.False(t, existingItem.IsBought)
}

func TestUpdateShoplistItemEmptyRequest(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add test item to shoplist
	item := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item",
		BrandName:  "Test Brand",
		ExtraInfo:  "Test Info",
		IsBought:   false,
	}
	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

	// Create request with empty JSON object
	requestBody := map[string]interface{}{}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/1/items/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{
		{Key: "id", Value: "1"},
		{Key: "itemId", Value: "1"},
	}

	UpdateShoplistItem(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Request body must include at least one of item_name, brand_name, extra_info or Is_bought.",
	}, response)

	// Verify item is unchanged in database
	var existingItem model.ShoplistItem
	err = testDB.First(&existingItem, 1).Error
	assert.NoError(t, err)
	assert.False(t, existingItem.IsBought)
}
