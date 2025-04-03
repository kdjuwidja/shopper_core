package apiHandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/model"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func TestCreateShoplistValid(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create test user
	testUser := model.User{
		ID:         "test-user-123",
		PostalCode: "238801", // Valid Singapore postal code
	}
	err := testDB.Create(&testUser).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUser.ID)
		c.Next()
	})

	router.PUT("/shoplist", CreateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"name": "Test Shoplist",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify database
	var shoplist model.Shoplist
	err = testDB.First(&shoplist, 1).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Shoplist", shoplist.Name)
	assert.Equal(t, testUser.ID, shoplist.OwnerID)

	// Check if the owner is also a member of the shoplist
	var member model.ShoplistMember
	err = testDB.Where("shop_list_id = ? AND member_id = ?", shoplist.ID, testUser.ID).First(&member).Error
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, member.MemberID)
}

func TestCreateShoplistInvalid(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create test user
	testUser := model.User{
		ID:         "test-user-123",
		PostalCode: "238801", // Valid Singapore postal code
	}
	err := testDB.Create(&testUser).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUser.ID)
		c.Next()
	})

	router.PUT("/shoplist", CreateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"invalid_field": "some value",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/shoplist", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Name is required",
	}, response)
}

func TestGetAllShoplistsOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		Nickname:   "Member",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplists
	shoplist1 := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Shoplist 1",
	}
	shoplist2 := model.Shoplist{
		ID:      2,
		OwnerID: owner.ID,
		Name:    "Shoplist 2",
	}
	err = testDB.Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplist2).Error
	assert.NoError(t, err)

	// Add owner as member to their own shoplists
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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/list", GetAllShoplists)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)

	GetAllShoplists(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedBody := map[string]interface{}{
		"shoplists": []interface{}{
			map[string]interface{}{
				"id":   float64(1),
				"name": "Shoplist 1",
				"owner": map[string]interface{}{
					"id":       "owner-123",
					"nickname": "Owner",
				},
			},
			map[string]interface{}{
				"id":   float64(2),
				"name": "Shoplist 2",
				"owner": map[string]interface{}{
					"id":       "owner-123",
					"nickname": "Owner",
				},
			},
		},
	}
	assert.Equal(t, expectedBody, response)

	// Verify the number of shoplists in the response
	shoplists, ok := response["shoplists"].([]interface{})
	assert.True(t, ok, "Response should contain shoplists array")
	assert.Equal(t, len(expectedBody["shoplists"].([]interface{})), len(shoplists),
		"Number of shoplists should match expected count")
}

func TestGetAllShoplistsMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		Nickname:   "Member",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplists
	shoplist1 := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Shoplist 1",
	}
	shoplist3 := model.Shoplist{
		ID:      3,
		OwnerID: member.ID,
		Name:    "Shoplist 3",
	}
	err = testDB.Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplist3).Error
	assert.NoError(t, err)

	// Add member to shoplist1
	shoplistMember := model.ShoplistMember{
		ID:         3,
		ShopListID: shoplist1.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add member as member to their own shoplist
	memberOwnShoplist := model.ShoplistMember{
		ID:         4,
		ShopListID: shoplist3.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&memberOwnShoplist).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/list", GetAllShoplists)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)

	GetAllShoplists(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedBody := map[string]interface{}{
		"shoplists": []interface{}{
			map[string]interface{}{
				"id":   float64(1),
				"name": "Shoplist 1",
				"owner": map[string]interface{}{
					"id":       "owner-123",
					"nickname": "Owner",
				},
			},
			map[string]interface{}{
				"id":   float64(3),
				"name": "Shoplist 3",
				"owner": map[string]interface{}{
					"id":       "member-123",
					"nickname": "Member",
				},
			},
		},
	}
	assert.Equal(t, expectedBody, response)

	// Verify the number of shoplists in the response
	shoplists, ok := response["shoplists"].([]interface{})
	assert.True(t, ok, "Response should contain shoplists array")
	assert.Equal(t, len(expectedBody["shoplists"].([]interface{})), len(shoplists),
		"Number of shoplists should match expected count")
}

func TestGetAllShoplistsNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create test user
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
		PostalCode: "238803",
	}
	err := testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/list", GetAllShoplists)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)

	GetAllShoplists(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedBody := map[string]interface{}{
		"shoplists": []interface{}{},
	}
	assert.Equal(t, expectedBody, response)

	// Verify the number of shoplists in the response
	shoplists, ok := response["shoplists"].([]interface{})
	assert.True(t, ok, "Response should contain shoplists array")
	assert.Equal(t, len(expectedBody["shoplists"].([]interface{})), len(shoplists),
		"Number of shoplists should match expected count")
}

func TestGetShoplistOwnerWithItems(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		Nickname:   "Member",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist with items
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
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test items to shoplist
	item1 := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 1",
		BrandName:  "Test Brand 1",
		ExtraInfo:  "Test Info 1",
		IsBought:   false,
	}
	item2 := model.ShoplistItem{
		ID:         2,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 2",
		BrandName:  "Test Brand 2",
		ExtraInfo:  "Test Info 2",
		IsBought:   true,
	}
	err = testDB.Create(&item1).Error
	assert.NoError(t, err)
	err = testDB.Create(&item2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	GetShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedBody := map[string]interface{}{
		"id":   float64(1),
		"name": "Test Shoplist",
		"owner": map[string]interface{}{
			"id":       "owner-123",
			"nickname": "Owner",
		},
		"members": []interface{}{
			map[string]interface{}{
				"id":       "owner-123",
				"nickname": "Owner",
			},
			map[string]interface{}{
				"id":       "member-123",
				"nickname": "Member",
			},
		},
		"items": []interface{}{
			map[string]interface{}{
				"id":         float64(1),
				"item_name":  "Test Item 1",
				"brand_name": "Test Brand 1",
				"extra_info": "Test Info 1",
				"is_bought":  false,
			},
			map[string]interface{}{
				"id":         float64(2),
				"item_name":  "Test Item 2",
				"brand_name": "Test Brand 2",
				"extra_info": "Test Info 2",
				"is_bought":  true,
			},
		},
	}
	assert.Equal(t, expectedBody["id"], response["id"])
	assert.Equal(t, expectedBody["name"], response["name"])
	assert.Equal(t, expectedBody["owner"], response["owner"])
	assert.ElementsMatch(t, expectedBody["members"], response["members"])
	assert.ElementsMatch(t, expectedBody["items"], response["items"])

	// Verify items are present
	items, ok := response["items"].([]interface{})
	assert.True(t, ok, "Response should contain items array")
	assert.Equal(t, len(expectedBody["items"].([]interface{})), len(items), "Number of items should match expected count")
}

func TestGetShoplistOwnerWithEmptyShoplist(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create empty shoplist
	emptyShoplist := model.Shoplist{
		ID:      2,
		OwnerID: owner.ID,
		Name:    "Empty Shoplist",
	}
	err = testDB.Create(&emptyShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	emptyOwnerMember := model.ShoplistMember{
		ID:         2,
		ShopListID: emptyShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&emptyOwnerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/2", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "2"}}

	GetShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedBody := map[string]interface{}{
		"id":   float64(2),
		"name": "Empty Shoplist",
		"owner": map[string]interface{}{
			"id":       "owner-123",
			"nickname": "Owner",
		},
		"members": []interface{}{
			map[string]interface{}{
				"id":       "owner-123",
				"nickname": "Owner",
			},
		},
		"items": []interface{}{},
	}
	assert.Equal(t, expectedBody["id"], response["id"])
	assert.Equal(t, expectedBody["name"], response["name"])
	assert.Equal(t, expectedBody["owner"], response["owner"])
	assert.ElementsMatch(t, expectedBody["members"], response["members"])
	assert.ElementsMatch(t, expectedBody["items"], response["items"])
}

func TestGetShoplistMemberWithItems(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		Nickname:   "Member",
		PostalCode: "238802",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist with items
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
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test items to shoplist
	item1 := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 1",
		BrandName:  "Test Brand 1",
		ExtraInfo:  "Test Info 1",
		IsBought:   false,
	}
	item2 := model.ShoplistItem{
		ID:         2,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 2",
		BrandName:  "Test Brand 2",
		ExtraInfo:  "Test Info 2",
		IsBought:   true,
	}
	err = testDB.Create(&item1).Error
	assert.NoError(t, err)
	err = testDB.Create(&item2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	GetShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expectedBody := map[string]interface{}{
		"id":   float64(1),
		"name": "Test Shoplist",
		"owner": map[string]interface{}{
			"id":       "owner-123",
			"nickname": "Owner",
		},
		"members": []interface{}{
			map[string]interface{}{
				"id":       "owner-123",
				"nickname": "Owner",
			},
			map[string]interface{}{
				"id":       "member-123",
				"nickname": "Member",
			},
		},
		"items": []interface{}{
			map[string]interface{}{
				"id":         float64(1),
				"item_name":  "Test Item 1",
				"brand_name": "Test Brand 1",
				"extra_info": "Test Info 1",
				"is_bought":  false,
			},
			map[string]interface{}{
				"id":         float64(2),
				"item_name":  "Test Item 2",
				"brand_name": "Test Brand 2",
				"extra_info": "Test Info 2",
				"is_bought":  true,
			},
		},
	}
	assert.Equal(t, expectedBody["id"], response["id"])
	assert.Equal(t, expectedBody["name"], response["name"])
	assert.Equal(t, expectedBody["owner"], response["owner"])
	assert.ElementsMatch(t, expectedBody["members"], response["members"])
	assert.ElementsMatch(t, expectedBody["items"], response["items"])

	// Verify items are present
	items, ok := response["items"].([]interface{})
	assert.True(t, ok, "Response should contain items array")
	assert.Equal(t, len(expectedBody["items"].([]interface{})), len(items), "Number of items should match expected count")
}

func TestGetShoplistNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	GetShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestGetShoplistNonExistent(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	router.GET("/shoplist/:id", GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/99999", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	GetShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestUpdateShoplistOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Original Name",
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
	router.PUT("/shoplist/:id", UpdateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify database
	var shoplist model.Shoplist
	err = testDB.First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, requestBody["name"], shoplist.Name)
}

func TestUpdateShoplistMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		Nickname:   "Member",
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
		Name:    "Original Name",
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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", UpdateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Only the owner can update this shoplist",
	}, response)
}

func TestUpdateShoplistNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
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
		Name:    "Original Name",
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
	router.PUT("/shoplist/:id", UpdateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestUpdateShoplistNonExistent(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", UpdateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"name": "Updated Name",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/99999/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestUpdateShoplistEmptyName(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Original Name",
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
	router.PUT("/shoplist/:id", UpdateShoplist)

	// Create request
	requestBody := map[string]interface{}{
		"name": "",
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Name is required",
	}, response)
}

func TestLeaveShopListMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member1 := model.User{
		ID:         "member1-123",
		PostalCode: "238802",
	}
	member2 := model.User{
		ID:         "member2-123",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member1).Error
	assert.NoError(t, err)
	err = testDB.Create(&member2).Error
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

	// Add members to shoplist
	shoplistMember1 := model.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member1.ID,
	}
	shoplistMember2 := model.ShoplistMember{
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member2.ID,
	}
	err = testDB.Create(&shoplistMember1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplistMember2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member1.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"message": "Successfully left the shoplist",
	}, response)

	// Verify member is removed from shoplist
	var memberCount int64
	err = testDB.Model(&model.ShoplistMember{}).
		Where("shop_list_id = ? AND member_id = ?", testShoplist.ID, member1.ID).
		Count(&memberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

	// Verify shoplist still exists
	var shoplist model.Shoplist
	err = testDB.First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err, "Shoplist should still exist")
	assert.Equal(t, testShoplist.ID, shoplist.ID, "Shoplist ID should match")

	// Verify remaining member count
	var remainingMemberCount int64
	err = testDB.Model(&model.ShoplistMember{}).
		Where("shop_list_id = ?", testShoplist.ID).
		Count(&remainingMemberCount).Error
	assert.NoError(t, err)
	assert.Greater(t, remainingMemberCount, int64(0), "Some members should remain")
}

func TestLeaveShopListOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member1 := model.User{
		ID:         "member1-123",
		PostalCode: "238802",
	}
	member2 := model.User{
		ID:         "member2-123",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member1).Error
	assert.NoError(t, err)
	err = testDB.Create(&member2).Error
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

	// Add members to shoplist
	shoplistMember1 := model.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member1.ID,
	}
	shoplistMember2 := model.ShoplistMember{
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member2.ID,
	}
	err = testDB.Create(&shoplistMember1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplistMember2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"message": "Successfully left the shoplist",
	}, response)

	// Verify member is removed from shoplist
	var memberCount int64
	err = testDB.Model(&model.ShoplistMember{}).
		Where("shop_list_id = ? AND member_id = ?", testShoplist.ID, owner.ID).
		Count(&memberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

	// Verify ownership was transferred to another member
	var shoplist model.Shoplist
	err = testDB.First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err)
	assert.True(t, shoplist.OwnerID == member1.ID || shoplist.OwnerID == member2.ID,
		"Ownership should be transferred to one of the remaining members")

	// Verify shoplist still exists
	err = testDB.First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err, "Shoplist should still exist")
	assert.Equal(t, testShoplist.ID, shoplist.ID, "Shoplist ID should match")

	// Verify remaining member count
	var remainingMemberCount int64
	err = testDB.Model(&model.ShoplistMember{}).
		Where("shop_list_id = ?", testShoplist.ID).
		Count(&remainingMemberCount).Error
	assert.NoError(t, err)
	assert.Greater(t, remainingMemberCount, int64(0), "Some members should remain")
}

func TestLeaveShopListLastMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist with only one member
	singleMemberShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Single Member Shoplist",
	}
	err = testDB.Create(&singleMemberShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to single member shoplist
	singleMemberOwner := model.ShoplistMember{
		ID:         1,
		ShopListID: singleMemberShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&singleMemberOwner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(singleMemberShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(singleMemberShoplist.ID)}}

	LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"message": "Successfully left the shoplist",
	}, response)

	// Verify member is removed from shoplist
	var memberCount int64
	err = testDB.Model(&model.ShoplistMember{}).
		Where("shop_list_id = ? AND member_id = ?", singleMemberShoplist.ID, owner.ID).
		Count(&memberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

	// Verify shoplist is deleted
	var shoplist model.Shoplist
	err = testDB.First(&shoplist, singleMemberShoplist.ID).Error
	assert.Error(t, err, "Shoplist should be deleted")
	assert.True(t, err.Error() == "record not found", "Error should be record not found")

	// Verify no members remain
	var remainingMemberCount int64
	err = testDB.Model(&model.ShoplistMember{}).
		Where("shop_list_id = ?", singleMemberShoplist.ID).
		Count(&remainingMemberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), remainingMemberCount, "No members should remain")
}

func TestLeaveShopListNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238804",
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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestLeaveShopListNonExistent(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	member1 := model.User{
		ID:         "member1-123",
		PostalCode: "238802",
	}
	err := testDB.Create(&member1).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/99999/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member1.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestRequestShopListShareCodeOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
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
	router.POST("/shoplist/:id/share-code", RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "share_code", "Response should contain share_code")
	assert.Contains(t, response, "expires_at", "Response should contain expires_at")

	// Verify share code format
	shareCode, ok := response["share_code"].(string)
	assert.True(t, ok, "Share code should be a string")
	assert.Len(t, shareCode, 6, "Share code should be 6 characters long")
	assert.Regexp(t, "^[A-Z0-9]+$", shareCode, "Share code should only contain uppercase letters and numbers")

	// Verify expiration time format
	expiresAt, ok := response["expires_at"].(string)
	assert.True(t, ok, "Expires at should be a string")
	expiryTime, err := time.Parse(time.RFC3339, expiresAt)
	assert.NoError(t, err, "Expires at should be in RFC3339 format")

	// Verify expiration time is exactly 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	assert.True(t, expiryTime.Sub(expectedExpiry) < time.Second && expectedExpiry.Sub(expiryTime) < time.Second,
		"Expiration time should be exactly 24 hours from now")
}

func TestRequestShopListShareCodeOwnerReplaceExisting(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
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

	// Create initial share code
	initialShareCode := model.ShoplistShareCode{
		ID:         1,
		ShopListID: testShoplist.ID,
		Code:       "OLD123",
		Expiry:     time.Now().Add(24 * time.Hour),
	}
	err = testDB.Create(&initialShareCode).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "share_code", "Response should contain share_code")
	assert.Contains(t, response, "expires_at", "Response should contain expires_at")

	// Verify share code format
	shareCode, ok := response["share_code"].(string)
	assert.True(t, ok, "Share code should be a string")
	assert.Len(t, shareCode, 6, "Share code should be 6 characters long")
	assert.Regexp(t, "^[A-Z0-9]+$", shareCode, "Share code should only contain uppercase letters and numbers")
	assert.NotEqual(t, "OLD123", shareCode, "New share code should be different from old one")

	// Verify only one share code exists
	var shareCodeCount int64
	err = testDB.Model(&model.ShoplistShareCode{}).
		Where("shop_list_id = ?", testShoplist.ID).
		Count(&shareCodeCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), shareCodeCount, "Should have exactly one share code")

	// Verify new share code is saved
	var newShareCode model.ShoplistShareCode
	err = testDB.Where("shop_list_id = ?", testShoplist.ID).First(&newShareCode).Error
	assert.NoError(t, err)
	assert.Equal(t, response["share_code"], newShareCode.Code, "Share code in database should match response")
}

func TestRequestShopListShareCodeMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		Nickname:   "Member",
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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Only the owner can generate share codes",
	}, response)
}

func TestRequestShopListShareCodeNonMember(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
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

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestRequestShopListShareCodeNonExistent(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/99999/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Not found",
	}, response)
}

func TestAddItemToShopListOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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

func TestUpdateShoplistItem(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member := model.User{
		ID:         "member-123",
		PostalCode: "238802",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
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

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Create unique items for each test case
	item1 := model.ShoplistItem{
		ID:         1,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 1",
		BrandName:  "Test Brand 1",
		ExtraInfo:  "Test Info 1",
		IsBought:   false,
	}
	item2 := model.ShoplistItem{
		ID:         2,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 2",
		BrandName:  "Test Brand 2",
		ExtraInfo:  "Test Info 2",
		IsBought:   true,
	}
	item3 := model.ShoplistItem{
		ID:         3,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 3",
		BrandName:  "Test Brand 3",
		ExtraInfo:  "Test Info 3",
		IsBought:   false,
	}
	item4 := model.ShoplistItem{
		ID:         4,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 4",
		BrandName:  "Test Brand 4",
		ExtraInfo:  "Test Info 4",
		IsBought:   false,
	}

	// Create items in the database
	err = testDB.Create(&item1).Error
	assert.NoError(t, err)
	err = testDB.Create(&item2).Error
	assert.NoError(t, err)
	err = testDB.Create(&item3).Error
	assert.NoError(t, err)
	err = testDB.Create(&item4).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		itemID         int
		userID         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Owner updating item bought status",
			shoplistID: 1,
			itemID:     1,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(1),
				"item_name":  "Test Item 1",
				"brand_name": "Test Brand 1",
				"extra_info": "Test Info 1",
				"is_bought":  true,
			},
		},
		{
			name:       "Member updating item bought status",
			shoplistID: 1,
			itemID:     2,
			userID:     member.ID,
			requestBody: map[string]interface{}{
				"is_bought": false,
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(2),
				"item_name":  "Test Item 2",
				"brand_name": "Test Brand 2",
				"extra_info": "Test Info 2",
				"is_bought":  false,
			},
		},
		{
			name:       "Non-member trying to update item",
			shoplistID: 1,
			itemID:     1,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:       "Updating item from non-existent shoplist",
			shoplistID: 99999,
			itemID:     1,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:       "Updating non-existent item",
			shoplistID: 1,
			itemID:     99999,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Item not found",
			},
		},
		{
			name:       "Updating item from different shoplist",
			shoplistID: 1,
			itemID:     9, // Item from a different shoplist
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Item not found",
			},
		},
		{
			name:       "Updating item name only",
			shoplistID: 1,
			itemID:     1, // Use item1
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name": "Updated Item Name",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(1),
				"item_name":  "Updated Item Name",
				"brand_name": "Test Brand 1",
				"extra_info": "Test Info 1",
				"is_bought":  true,
			},
		},
		{
			name:       "Updating brand name only",
			shoplistID: 1,
			itemID:     2, // Use item2
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"brand_name": "Updated Brand Name",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(2),
				"item_name":  "Test Item 2",
				"brand_name": "Updated Brand Name",
				"extra_info": "Test Info 2",
				"is_bought":  false,
			},
		},
		{
			name:       "Updating extra info only",
			shoplistID: 1,
			itemID:     3, // Use item3
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"extra_info": "Updated Extra Info",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(3),
				"item_name":  "Test Item 3",
				"brand_name": "Test Brand 3",
				"extra_info": "Updated Extra Info",
				"is_bought":  false,
			},
		},
		{
			name:       "Updating multiple fields",
			shoplistID: 1,
			itemID:     4, // Use item4
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name":  "Updated Item Name",
				"brand_name": "Updated Brand Name",
				"extra_info": "Updated Extra Info",
				"is_bought":  true,
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(4),
				"item_name":  "Updated Item Name",
				"brand_name": "Updated Brand Name",
				"extra_info": "Updated Extra Info",
				"is_bought":  true,
			},
		},
		{
			name:           "Empty request body",
			shoplistID:     1,
			itemID:         1,
			userID:         owner.ID,
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Request body must include at least one of item_name, brand_name, extra_info or Is_bought.",
			},
		},
		{
			name:       "Empty request body with null values",
			shoplistID: 1,
			itemID:     1,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name":  nil,
				"brand_name": nil,
				"extra_info": nil,
				"is_bought":  nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Request body must include at least one of item_name, brand_name, extra_info or Is_bought.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/shoplist/:id/items/:itemId", UpdateShoplistItem)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/items/"+strconv.Itoa(tt.itemID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("userID", tt.userID)
			c.Params = []gin.Param{
				{Key: "id", Value: strconv.Itoa(tt.shoplistID)},
				{Key: "itemId", Value: strconv.Itoa(tt.itemID)},
			}

			UpdateShoplistItem(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If update was successful, verify database
			if tt.expectedStatus == http.StatusOK {
				var item model.ShoplistItem
				err := testDB.First(&item, tt.itemID).Error
				assert.NoError(t, err)

				// Verify all fields match expected values
				assert.Equal(t, tt.expectedBody["item_name"], item.ItemName, "Item name should match expected value")
				assert.Equal(t, tt.expectedBody["brand_name"], item.BrandName, "Brand name should match expected value")
				assert.Equal(t, tt.expectedBody["extra_info"], item.ExtraInfo, "Extra info should match expected value")
				assert.Equal(t, tt.expectedBody["is_bought"], item.IsBought, "Is bought status should match expected value")
			}
		})
	}
}

func TestUpdateShoplistItemOwner(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
	testDB := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB()

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
