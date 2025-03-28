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

func TestCreateShoplist(t *testing.T) {
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

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid shoplist creation",
			requestBody: map[string]interface{}{
				"name": "Test Shoplist",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   map[string]interface{}{},
		},
		{
			name: "Missing name",
			requestBody: map[string]interface{}{
				"invalid_field": "some value",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Name is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Add middleware to set user_id
			router.Use(func(c *gin.Context) {
				c.Set("user_id", testUser.ID)
				c.Next()
			})

			router.PUT("/shoplist", CreateShoplist)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/shoplist", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If creation was successful, verify database
			if tt.expectedStatus == http.StatusCreated {
				var shoplist model.Shoplist
				err := testDB.First(&shoplist, 1).Error
				assert.NoError(t, err)
				assert.Equal(t, "Test Shoplist", shoplist.Name)
				assert.Equal(t, testUser.ID, shoplist.OwnerID)
			}
		})
	}
}

func TestGetAllShoplists(t *testing.T) {
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
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)
	err = testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplists
	shoplist1 := model.Shoplist{
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Shoplist 1",
	}
	shoplist2 := model.Shoplist{
		ID:      10001,
		OwnerID: owner.ID,
		Name:    "Shoplist 2",
	}
	shoplist3 := model.Shoplist{
		ID:      10002,
		OwnerID: member.ID,
		Name:    "Shoplist 3",
	}
	err = testDB.Create(&shoplist1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplist2).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplist3).Error
	assert.NoError(t, err)

	// Add owner as member to their own shoplists
	ownerMember1 := model.ShoplistMember{
		ID:         10000,
		ShopListID: shoplist1.ID,
		MemberID:   owner.ID,
	}
	ownerMember2 := model.ShoplistMember{
		ID:         10001,
		ShopListID: shoplist2.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember1).Error
	assert.NoError(t, err)
	err = testDB.Create(&ownerMember2).Error
	assert.NoError(t, err)

	// Add member to shoplist1
	shoplistMember := model.ShoplistMember{
		ID:         10002,
		ShopListID: shoplist1.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add member as member to their own shoplist
	memberOwnShoplist := model.ShoplistMember{
		ID:         10003,
		ShopListID: shoplist3.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&memberOwnShoplist).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Owner getting all shoplists",
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"shoplists": []interface{}{
					map[string]interface{}{
						"id":   float64(10000),
						"name": "Shoplist 1",
						"owner": map[string]interface{}{
							"id":       "owner-123",
							"nickname": "Owner",
						},
					},
					map[string]interface{}{
						"id":   float64(10001),
						"name": "Shoplist 2",
						"owner": map[string]interface{}{
							"id":       "owner-123",
							"nickname": "Owner",
						},
					},
				},
			},
		},
		{
			name:           "Member getting all shoplists",
			userID:         member.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"shoplists": []interface{}{
					map[string]interface{}{
						"id":   float64(10000),
						"name": "Shoplist 1",
						"owner": map[string]interface{}{
							"id":       "owner-123",
							"nickname": "Owner",
						},
					},
					map[string]interface{}{
						"id":   float64(10002),
						"name": "Shoplist 3",
						"owner": map[string]interface{}{
							"id":       "member-123",
							"nickname": "Member",
						},
					},
				},
			},
		},
		{
			name:           "Non-member getting all shoplists",
			userID:         nonMember.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"shoplists": []interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			c.Set("user_id", tt.userID)

			GetAllShoplists(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// Verify the number of shoplists in the response
			if tt.expectedStatus == http.StatusOK {
				shoplists, ok := response["shoplists"].([]interface{})
				assert.True(t, ok, "Response should contain shoplists array")
				assert.Equal(t, len(tt.expectedBody["shoplists"].([]interface{})), len(shoplists),
					"Number of shoplists should match expected count")
			}
		})
	}
}

func TestGetShoplist(t *testing.T) {
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
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
		PostalCode: "238803",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member).Error
	assert.NoError(t, err)
	err = testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist with items
	testShoplist := model.Shoplist{
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Create empty shoplist
	emptyShoplist := model.Shoplist{
		ID:      10001,
		OwnerID: owner.ID,
		Name:    "Empty Shoplist",
	}
	err = testDB.Create(&emptyShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplists
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	emptyOwnerMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: emptyShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)
	err = testDB.Create(&emptyOwnerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10002,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test items to shoplist
	item1 := model.ShoplistItem{
		ID:         10000,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 1",
		BrandName:  "Test Brand 1",
		ExtraInfo:  "Test Info 1",
		IsBought:   false,
	}
	item2 := model.ShoplistItem{
		ID:         10001,
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

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Owner accessing shoplist with items",
			shoplistID:     10000,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":   float64(10000),
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
						"id":         float64(10000),
						"item_name":  "Test Item 1",
						"brand_name": "Test Brand 1",
						"extra_info": "Test Info 1",
						"is_bought":  false,
					},
					map[string]interface{}{
						"id":         float64(10001),
						"item_name":  "Test Item 2",
						"brand_name": "Test Brand 2",
						"extra_info": "Test Info 2",
						"is_bought":  true,
					},
				},
			},
		},
		{
			name:           "Owner accessing empty shoplist",
			shoplistID:     10001,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":   float64(10001),
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
			},
		},
		{
			name:           "Member accessing shoplist with items",
			shoplistID:     10000,
			userID:         member.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":   float64(10000),
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
						"id":         float64(10000),
						"item_name":  "Test Item 1",
						"brand_name": "Test Brand 1",
						"extra_info": "Test Info 1",
						"is_bought":  false,
					},
					map[string]interface{}{
						"id":         float64(10001),
						"item_name":  "Test Item 2",
						"brand_name": "Test Brand 2",
						"extra_info": "Test Info 2",
						"is_bought":  true,
					},
				},
			},
		},
		{
			name:           "Non-member accessing shoplist",
			shoplistID:     10000,
			userID:         nonMember.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Not authorized to access this shoplist",
			},
		},
		{
			name:           "Non-existent shoplist",
			shoplistID:     99999,
			userID:         owner.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/shoplist/:id", GetShoplist)

			// Create request
			req, _ := http.NewRequest("GET", "/shoplist/"+strconv.Itoa(tt.shoplistID), nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			GetShoplist(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody["id"], response["id"])
			assert.Equal(t, tt.expectedBody["name"], response["name"])
			assert.Equal(t, tt.expectedBody["owner"], response["owner"])
			assert.ElementsMatch(t, tt.expectedBody["members"], response["members"])
			assert.ElementsMatch(t, tt.expectedBody["items"], response["items"])

			// For successful responses, verify items are present
			if tt.expectedStatus == http.StatusOK {
				items, ok := response["items"].([]interface{})
				assert.True(t, ok, "Response should contain items array")
				expectedItems := tt.expectedBody["items"].([]interface{})
				assert.Equal(t, len(expectedItems), len(items), "Number of items should match expected count")
			}
		})
	}
}

func TestUpdateShoplist(t *testing.T) {
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
	nonMember := model.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
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
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Original Name",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Owner updating shoplist name",
			shoplistID: 10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{},
		},
		{
			name:       "Member trying to update shoplist",
			shoplistID: 10000,
			userID:     member.ID,
			requestBody: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "Only the owner can update this shoplist",
			},
		},
		{
			name:       "Non-member trying to update shoplist",
			shoplistID: 10000,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:       "Updating non-existent shoplist",
			shoplistID: 99999,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:       "Empty name update",
			shoplistID: 10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"name": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Name is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.PUT("/shoplist/:id", UpdateShoplist)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/update", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			UpdateShoplist(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If update was successful, verify database
			if tt.expectedStatus == http.StatusOK {
				var shoplist model.Shoplist
				err := testDB.First(&shoplist, tt.shoplistID).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody["name"], shoplist.Name)
			}
		})
	}
}

func TestLeaveShopList(t *testing.T) {
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
	nonMember := model.User{
		ID:         "non-member-123",
		PostalCode: "238804",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)
	err = testDB.Create(&member1).Error
	assert.NoError(t, err)
	err = testDB.Create(&member2).Error
	assert.NoError(t, err)
	err = testDB.Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := model.Shoplist{
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add members to shoplist
	shoplistMember1 := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member1.ID,
	}
	shoplistMember2 := model.ShoplistMember{
		ID:         10002,
		ShopListID: testShoplist.ID,
		MemberID:   member2.ID,
	}
	err = testDB.Create(&shoplistMember1).Error
	assert.NoError(t, err)
	err = testDB.Create(&shoplistMember2).Error
	assert.NoError(t, err)

	// Create another shoplist with only one member
	singleMemberShoplist := model.Shoplist{
		ID:      10001,
		OwnerID: owner.ID,
		Name:    "Single Member Shoplist",
	}
	err = testDB.Create(&singleMemberShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to single member shoplist
	singleMemberOwner := model.ShoplistMember{
		ID:         10003,
		ShopListID: singleMemberShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&singleMemberOwner).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Member leaving shoplist",
			shoplistID:     10000,
			userID:         member1.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Successfully left the shoplist",
			},
		},
		{
			name:           "Owner leaving shoplist",
			shoplistID:     10000,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Successfully left the shoplist",
			},
		},
		{
			name:           "Last member leaving shoplist",
			shoplistID:     10001,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Successfully left the shoplist",
			},
		},
		{
			name:           "Non-member trying to leave shoplist",
			shoplistID:     10000,
			userID:         nonMember.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:           "Leaving non-existent shoplist",
			shoplistID:     99999,
			userID:         member1.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/shoplist/:id/leave", LeaveShopList)

			// Create request
			req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/leave", nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			LeaveShopList(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If leaving was successful, verify database
			if tt.expectedStatus == http.StatusOK {
				// Verify member is removed from shoplist
				var memberCount int64
				err := testDB.Model(&model.ShoplistMember{}).
					Where("shop_list_id = ? AND member_id = ?", tt.shoplistID, tt.userID).
					Count(&memberCount).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

				// If owner left, verify ownership was transferred to another member
				if tt.userID == owner.ID && tt.shoplistID != singleMemberShoplist.ID {
					var shoplist model.Shoplist
					err := testDB.First(&shoplist, tt.shoplistID).Error
					assert.NoError(t, err)
					assert.True(t, shoplist.OwnerID == member1.ID || shoplist.OwnerID == member2.ID,
						"Ownership should be transferred to one of the remaining members")
				}

				// If last member left, verify shoplist is deleted
				if tt.shoplistID == singleMemberShoplist.ID {
					var shoplist model.Shoplist
					err := testDB.First(&shoplist, tt.shoplistID).Error
					assert.Error(t, err, "Shoplist should be deleted")
					assert.True(t, err.Error() == "record not found", "Error should be record not found")
				} else {
					// Verify shoplist still exists when members remain
					var shoplist model.Shoplist
					err := testDB.First(&shoplist, tt.shoplistID).Error
					assert.NoError(t, err, "Shoplist should still exist")
					assert.Equal(t, tt.shoplistID, shoplist.ID, "Shoplist ID should match")
				}

				// Verify remaining member count
				var remainingMemberCount int64
				err = testDB.Model(&model.ShoplistMember{}).
					Where("shop_list_id = ?", tt.shoplistID).
					Count(&remainingMemberCount).Error
				assert.NoError(t, err)

				if tt.shoplistID == singleMemberShoplist.ID {
					assert.Equal(t, int64(0), remainingMemberCount, "No members should remain")
				} else {
					assert.Greater(t, remainingMemberCount, int64(0), "Some members should remain")
				}
			}
		})
	}
}

func TestRequestShopListShareCode(t *testing.T) {
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
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Owner requesting share code",
			shoplistID:     10000,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"share_code": "", // Will be verified for format
				"expires_at": "", // Will be verified for format
			},
		},
		{
			name:           "Member requesting share code",
			shoplistID:     10000,
			userID:         member.ID,
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "Only the owner can generate share codes",
			},
		},
		{
			name:           "Non-member requesting share code",
			shoplistID:     10000,
			userID:         nonMember.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:           "Requesting share code for non-existent shoplist",
			shoplistID:     99999,
			userID:         owner.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/shoplist/:id/share-code", RequestShopListShareCode)

			// Create request
			req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/share-code", nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			RequestShopListShareCode(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// For successful responses, verify the structure but not the exact values
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, response, "share_code", "Response should contain share_code")
				assert.Contains(t, response, "expires_at", "Response should contain expires_at")

				// Verify share code format (8 characters, alphanumeric)
				shareCode, ok := response["share_code"].(string)
				assert.True(t, ok, "Share code should be a string")
				assert.Len(t, shareCode, 6, "Share code should be 8 characters long")
				assert.Regexp(t, "^[A-Z0-9]+$", shareCode, "Share code should only contain uppercase letters and numbers")

				// Verify expiration time format (RFC3339)
				expiresAt, ok := response["expires_at"].(string)
				assert.True(t, ok, "Expires at should be a string")
				expiryTime, err := time.Parse(time.RFC3339, expiresAt)
				assert.NoError(t, err, "Expires at should be in RFC3339 format")

				// Verify expiration time is exactly 24 hours from now
				expectedExpiry := time.Now().Add(24 * time.Hour)
				// Allow for a small time difference (within 1 second) due to test execution time
				assert.True(t, expiryTime.Sub(expectedExpiry) < time.Second && expectedExpiry.Sub(expiryTime) < time.Second,
					"Expiration time should be exactly 24 hours from now")
			} else {
				assert.Equal(t, tt.expectedBody, response)
			}

			// If share code was generated successfully, verify database
			if tt.expectedStatus == http.StatusOK {
				var shoplist model.Shoplist
				err := testDB.First(&shoplist, tt.shoplistID).Error
				assert.NoError(t, err)
			}
		})
	}
}

func TestRevokeShopListShareCode(t *testing.T) {
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

	// Create test shoplist with share code
	testShoplist := model.Shoplist{
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Owner revoking share code",
			shoplistID:     10000,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Share code revoked successfully",
			},
		},
		{
			name:           "Member trying to revoke share code",
			shoplistID:     10000,
			userID:         member.ID,
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "Only the owner can revoke share codes",
			},
		},
		{
			name:           "Non-member trying to revoke share code",
			shoplistID:     10000,
			userID:         nonMember.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:           "Revoking share code for non-existent shoplist",
			shoplistID:     99999,
			userID:         owner.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:           "Revoking non-existent share code",
			shoplistID:     10000,
			userID:         owner.ID,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "No active share code to revoke",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.DELETE("/shoplist/:id/share-code", RevokeShopListShareCode)

			// Create request
			req, _ := http.NewRequest("DELETE", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/share-code", nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("userID", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			RevokeShopListShareCode(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If share code was revoked successfully, verify database
			if tt.expectedStatus == http.StatusOK {
				var shoplist model.Shoplist
				err := testDB.First(&shoplist, tt.shoplistID).Error
				assert.NoError(t, err)
			}
		})
	}
}

func TestJoinShopList(t *testing.T) {
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

	// Create test shoplist with share code
	testShoplist := model.Shoplist{
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Create another shoplist with expired share code
	expiredShoplist := model.Shoplist{
		ID:      10001,
		OwnerID: owner.ID,
		Name:    "Expired Shoplist",
	}
	err = testDB.Create(&expiredShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to expired shoplist
	expiredOwnerMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: expiredShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&expiredOwnerMember).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Valid share code join",
			shoplistID: 10000,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"share_code": "TEST123",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Successfully joined the shoplist",
			},
		},
		{
			name:       "Invalid share code",
			shoplistID: 10000,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"share_code": "INVALID",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid share code",
			},
		},
		{
			name:       "Expired share code",
			shoplistID: 10001,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"share_code": "EXPIRED",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Share code has expired",
			},
		},
		{
			name:       "Empty share code",
			shoplistID: 10000,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"share_code": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Share code is required",
			},
		},
		{
			name:       "Missing share code field",
			shoplistID: 10000,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"invalid_field": "some value",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Share code is required",
			},
		},
		{
			name:       "Already a member",
			shoplistID: 10000,
			userID:     member.ID,
			requestBody: map[string]interface{}{
				"share_code": "TEST123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Already a member of this shoplist",
			},
		},
		{
			name:       "Non-existent shoplist",
			shoplistID: 99999,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"share_code": "TEST123",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/shoplist/:id/join", JoinShopList)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/join", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("userID", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			JoinShopList(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If join was successful, verify database
			if tt.expectedStatus == http.StatusOK {
				var memberCount int64
				err := testDB.Model(&model.ShoplistMember{}).
					Where("shop_list_id = ? AND member_id = ?", tt.shoplistID, tt.userID).
					Count(&memberCount).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(1), memberCount, "User should be added as member")
			}
		})
	}
}

func TestAddItemToShopList(t *testing.T) {
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
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		shoplistID     int
		userID         string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Owner adding item",
			shoplistID: 10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name":  "Test Item",
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":         float64(10000),
				"item_name":  "Test Item",
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
				"is_bought":  false,
			},
		},
		{
			name:       "Member adding item",
			shoplistID: 10000,
			userID:     member.ID,
			requestBody: map[string]interface{}{
				"item_name":  "Another Item",
				"brand_name": "Another Brand",
				"extra_info": "Another Info",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":         float64(10001),
				"item_name":  "Another Item",
				"brand_name": "Another Brand",
				"extra_info": "Another Info",
				"is_bought":  false,
			},
		},
		{
			name:       "Non-member trying to add item",
			shoplistID: 10000,
			userID:     nonMember.ID,
			requestBody: map[string]interface{}{
				"item_name":  "Test Item",
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:       "Adding item to non-existent shoplist",
			shoplistID: 99999,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name":  "Test Item",
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:       "Missing required fields",
			shoplistID: 10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Item name is required",
			},
		},
		{
			name:       "Invalid quantity",
			shoplistID: 10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name":  "Test Item",
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Quantity must be greater than 0",
			},
		},
		{
			name:       "Empty name",
			shoplistID: 10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"item_name":  "",
				"brand_name": "Test Brand",
				"extra_info": "Test Info",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Item name is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/shoplist/:id/items", AddItemToShopList)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/items", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("userID", tt.userID)
			c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(tt.shoplistID)}}

			AddItemToShopList(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If item was added successfully, verify database
			if tt.expectedStatus == http.StatusCreated {
				var item model.ShoplistItem
				err := testDB.First(&item, response["id"]).Error
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody["item_name"], item.ItemName)
				assert.Equal(t, tt.requestBody["brand_name"], item.BrandName)
				assert.Equal(t, tt.requestBody["extra_info"], item.ExtraInfo)
				assert.False(t, item.IsBought)
				assert.Equal(t, tt.shoplistID, item.ShopListID)
			}
		})
	}
}

func TestRemoveItemFromShopList(t *testing.T) {
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
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test items to shoplist
	item1 := model.ShoplistItem{
		ID:         10000,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 1",
		BrandName:  "Test Brand 1",
		ExtraInfo:  "Test Info 1",
		IsBought:   false,
	}
	item2 := model.ShoplistItem{
		ID:         10001,
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

	tests := []struct {
		name           string
		shoplistID     int
		itemID         int
		userID         string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Owner removing item",
			shoplistID:     10000,
			itemID:         10000,
			userID:         owner.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Item removed successfully",
			},
		},
		{
			name:           "Member removing item",
			shoplistID:     10000,
			itemID:         10001,
			userID:         member.ID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Item removed successfully",
			},
		},
		{
			name:           "Non-member trying to remove item",
			shoplistID:     10000,
			itemID:         10000,
			userID:         nonMember.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:           "Removing item from non-existent shoplist",
			shoplistID:     99999,
			itemID:         10000,
			userID:         owner.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Shoplist not found",
			},
		},
		{
			name:           "Removing non-existent item",
			shoplistID:     10000,
			itemID:         99999,
			userID:         owner.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Item not found",
			},
		},
		{
			name:           "Removing item from different shoplist",
			shoplistID:     10000,
			itemID:         10002, // Item from a different shoplist
			userID:         owner.ID,
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Item not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Gin router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.DELETE("/shoplist/:id/items/:itemId", RemoveItemFromShopList)

			// Create request
			req, _ := http.NewRequest("DELETE", "/shoplist/"+strconv.Itoa(tt.shoplistID)+"/items/"+strconv.Itoa(tt.itemID), nil)
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

			RemoveItemFromShopList(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			// If item was removed successfully, verify database
			if tt.expectedStatus == http.StatusOK {
				var item model.ShoplistItem
				err := testDB.First(&item, tt.itemID).Error
				assert.Error(t, err, "Item should be deleted from database")
				assert.True(t, err.Error() == "record not found", "Error should be record not found")
			}
		})
	}
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
		ID:      10000,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testDB.Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := model.ShoplistMember{
		ID:         10000,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testDB.Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := model.ShoplistMember{
		ID:         10001,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testDB.Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Add test items to shoplist
	item1 := model.ShoplistItem{
		ID:         10000,
		ShopListID: testShoplist.ID,
		ItemName:   "Test Item 1",
		BrandName:  "Test Brand 1",
		ExtraInfo:  "Test Info 1",
		IsBought:   false,
	}
	item2 := model.ShoplistItem{
		ID:         10001,
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
			shoplistID: 10000,
			itemID:     10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(10000),
				"item_name":  "Test Item 1",
				"brand_name": "Test Brand 1",
				"extra_info": "Test Info 1",
				"is_bought":  true,
			},
		},
		{
			name:       "Member updating item bought status",
			shoplistID: 10000,
			itemID:     10001,
			userID:     member.ID,
			requestBody: map[string]interface{}{
				"is_bought": false,
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":         float64(10001),
				"item_name":  "Test Item 2",
				"brand_name": "Test Brand 2",
				"extra_info": "Test Info 2",
				"is_bought":  false,
			},
		},
		{
			name:       "Non-member trying to update item",
			shoplistID: 10000,
			itemID:     10000,
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
			itemID:     10000,
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
			shoplistID: 10000,
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
			shoplistID: 10000,
			itemID:     10002, // Item from a different shoplist
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
			name:       "Missing is_bought field",
			shoplistID: 10000,
			itemID:     10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"invalid_field": "some value",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Is_bought field is required",
			},
		},
		{
			name:       "Trying to update other fields",
			shoplistID: 10000,
			itemID:     10000,
			userID:     owner.ID,
			requestBody: map[string]interface{}{
				"is_bought": true,
				"item_name": "New Name",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Only is_bought field can be updated",
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
				assert.Equal(t, tt.requestBody["is_bought"], item.IsBought, "Item bought status should be updated")
				// Verify other fields remain unchanged
				assert.Equal(t, tt.expectedBody["item_name"], item.ItemName, "Item name should not be changed")
				assert.Equal(t, tt.expectedBody["brand_name"], item.BrandName, "Brand name should not be changed")
				assert.Equal(t, tt.expectedBody["extra_info"], item.ExtraInfo, "Extra info should not be changed")
			}
		})
	}
}
