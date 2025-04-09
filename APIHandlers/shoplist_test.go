package apiHandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func setUpShoplistTestEnv(t *testing.T) (*ShoplistHandler, *db.MySQLConnectionPool) {
	testDBConn := testutil.SetupTestEnv(t)
	shoplistHandler := InitializeShoplistHandler(*testDBConn)
	return shoplistHandler, testDBConn
}

func TestCreateShoplistValid(t *testing.T) {
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	testUser := model.User{
		ID:         "test-user-123",
		PostalCode: "238801", // Valid Singapore postal code
	}
	err := testConn.GetDB().Create(&testUser).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUser.ID)
		c.Next()
	})

	router.PUT("/shoplist", shoplistHandler.CreateShoplist)

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
	err = testConn.GetDB().First(&shoplist, 1).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Shoplist", shoplist.Name)
	assert.Equal(t, testUser.ID, shoplist.OwnerID)

	// Check if the owner is also a member of the shoplist
	var member model.ShoplistMember
	err = testConn.GetDB().Where("shop_list_id = ? AND member_id = ?", shoplist.ID, testUser.ID).First(&member).Error
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, member.MemberID)
}

func TestCreateShoplistInvalid(t *testing.T) {
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	testUser := model.User{
		ID:         "test-user-123",
		PostalCode: "238801", // Valid Singapore postal code
	}
	err := testConn.GetDB().Create(&testUser).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUser.ID)
		c.Next()
	})

	router.PUT("/shoplist", shoplistHandler.CreateShoplist)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

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
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member).Error
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
	err = testConn.GetDB().Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&shoplist2).Error
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
	err = testConn.GetDB().Create(&ownerMember1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&ownerMember2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/list", shoplistHandler.GetAllShoplists)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)

	shoplistHandler.GetAllShoplists(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

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
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member).Error
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
	err = testConn.GetDB().Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&shoplist3).Error
	assert.NoError(t, err)

	// Add member to shoplist1
	shoplistMember := model.ShoplistMember{
		ID:         3,
		ShopListID: shoplist1.ID,
		MemberID:   member.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add member as member to their own shoplist
	memberOwnShoplist := model.ShoplistMember{
		ID:         4,
		ShopListID: shoplist3.ID,
		MemberID:   member.ID,
	}
	err = testConn.GetDB().Create(&memberOwnShoplist).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/list", shoplistHandler.GetAllShoplists)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)

	shoplistHandler.GetAllShoplists(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
		PostalCode: "238803",
	}
	err := testConn.GetDB().Create(&nonMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/list", shoplistHandler.GetAllShoplists)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)

	shoplistHandler.GetAllShoplists(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

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
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist with items
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember).Error
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
	err = testConn.GetDB().Create(&item1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&item2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", shoplistHandler.GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.GetShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create empty shoplist
	emptyShoplist := model.Shoplist{
		ID:      2,
		OwnerID: owner.ID,
		Name:    "Empty Shoplist",
	}
	err = testConn.GetDB().Create(&emptyShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	emptyOwnerMember := model.ShoplistMember{
		ID:         2,
		ShopListID: emptyShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&emptyOwnerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", shoplistHandler.GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/2", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "2"}}

	shoplistHandler.GetShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

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
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist with items
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember).Error
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
	err = testConn.GetDB().Create(&item1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&item2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", shoplistHandler.GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.GetShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238803",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", shoplistHandler.GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.GetShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/shoplist/:id", shoplistHandler.GetShoplist)

	// Create request
	req, _ := http.NewRequest("GET", "/shoplist/99999", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	shoplistHandler.GetShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Original Name",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", shoplistHandler.UpdateShoplist)

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

	shoplistHandler.UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify database
	var shoplist model.Shoplist
	err = testConn.GetDB().First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, requestBody["name"], shoplist.Name)
}

func TestUpdateShoplistMember(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

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
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Original Name",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", shoplistHandler.UpdateShoplist)

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

	shoplistHandler.UpdateShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

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
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Original Name",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", shoplistHandler.UpdateShoplist)

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

	shoplistHandler.UpdateShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", shoplistHandler.UpdateShoplist)

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

	shoplistHandler.UpdateShoplist(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Original Name",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/shoplist/:id", shoplistHandler.UpdateShoplist)

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

	shoplistHandler.UpdateShoplist(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"error": "Name is required",
	}, response)
}
