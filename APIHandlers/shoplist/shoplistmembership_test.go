package apiHandlersshoplist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/db"
)

func TestLeaveShopListMember(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := db.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member1 := db.User{
		ID:         "member1-123",
		PostalCode: "238802",
	}
	member2 := db.User{
		ID:         "member2-123",
		PostalCode: "238803",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member2).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add members to shoplist
	shoplistMember1 := db.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member1.ID,
	}
	shoplistMember2 := db.ShoplistMember{
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member2.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&shoplistMember2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", shoplistHandler.LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member1.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	shoplistHandler.LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify member is removed from shoplist
	var memberCount int64
	err = testConn.GetDB().Model(&db.ShoplistMember{}).
		Where("shop_list_id = ? AND member_id = ?", testShoplist.ID, member1.ID).
		Count(&memberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

	// Verify shoplist still exists
	var shoplist db.Shoplist
	err = testConn.GetDB().First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err, "Shoplist should still exist")
	assert.Equal(t, testShoplist.ID, shoplist.ID, "Shoplist ID should match")

	// Verify remaining member count
	var remainingMemberCount int64
	err = testConn.GetDB().Model(&db.ShoplistMember{}).
		Where("shop_list_id = ?", testShoplist.ID).
		Count(&remainingMemberCount).Error
	assert.NoError(t, err)
	assert.Greater(t, remainingMemberCount, int64(0), "Some members should remain")
}

func TestLeaveShopListOwner(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := db.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	member1 := db.User{
		ID:         "member1-123",
		PostalCode: "238802",
	}
	member2 := db.User{
		ID:         "member2-123",
		PostalCode: "238803",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member2).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add members to shoplist
	shoplistMember1 := db.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member1.ID,
	}
	shoplistMember2 := db.ShoplistMember{
		ID:         3,
		ShopListID: testShoplist.ID,
		MemberID:   member2.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember1).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&shoplistMember2).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", shoplistHandler.LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	shoplistHandler.LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify member is removed from shoplist
	var memberCount int64
	err = testConn.GetDB().Model(&db.ShoplistMember{}).
		Where("shop_list_id = ? AND member_id = ?", testShoplist.ID, owner.ID).
		Count(&memberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

	// Verify ownership was transferred to another member
	var shoplist db.Shoplist
	err = testConn.GetDB().First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err)
	assert.True(t, shoplist.OwnerID == member1.ID || shoplist.OwnerID == member2.ID,
		"Ownership should be transferred to one of the remaining members")

	// Verify shoplist still exists
	err = testConn.GetDB().First(&shoplist, testShoplist.ID).Error
	assert.NoError(t, err, "Shoplist should still exist")
	assert.Equal(t, testShoplist.ID, shoplist.ID, "Shoplist ID should match")

	// Verify remaining member count
	var remainingMemberCount int64
	err = testConn.GetDB().Model(&db.ShoplistMember{}).
		Where("shop_list_id = ?", testShoplist.ID).
		Count(&remainingMemberCount).Error
	assert.NoError(t, err)
	assert.Greater(t, remainingMemberCount, int64(0), "Some members should remain")
}

func TestLeaveShopListLastMember(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := db.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist with only one member
	singleMemberShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Single Member Shoplist",
	}
	err = testConn.GetDB().Create(&singleMemberShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to single member shoplist
	singleMemberOwner := db.ShoplistMember{
		ID:         1,
		ShopListID: singleMemberShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&singleMemberOwner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", shoplistHandler.LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(singleMemberShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(singleMemberShoplist.ID)}}

	shoplistHandler.LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify member is removed from shoplist
	var memberCount int64
	err = testConn.GetDB().Model(&db.ShoplistMember{}).
		Where("shop_list_id = ? AND member_id = ?", singleMemberShoplist.ID, owner.ID).
		Count(&memberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), memberCount, "Member should be removed from shoplist")

	// Verify shoplist is deleted
	var shoplist db.Shoplist
	err = testConn.GetDB().First(&shoplist, singleMemberShoplist.ID).Error
	assert.Error(t, err, "Shoplist should be deleted")
	assert.True(t, err.Error() == "record not found", "Error should be record not found")

	// Verify no members remain
	var remainingMemberCount int64
	err = testConn.GetDB().Model(&db.ShoplistMember{}).
		Where("shop_list_id = ?", singleMemberShoplist.ID).
		Count(&remainingMemberCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), remainingMemberCount, "No members should remain")
}

func TestLeaveShopListNonMember(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := db.User{
		ID:         "owner-123",
		PostalCode: "238801",
	}
	nonMember := db.User{
		ID:         "non-member-123",
		PostalCode: "238804",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", shoplistHandler.LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/"+strconv.Itoa(testShoplist.ID)+"/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: strconv.Itoa(testShoplist.ID)}}

	shoplistHandler.LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"code": "SHP_00001", "error": "Shoplist not found.",
	}, response)
}

func TestLeaveShopListNonExistent(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	member1 := db.User{
		ID:         "member1-123",
		PostalCode: "238802",
	}
	err := testConn.GetDB().Create(&member1).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/leave", shoplistHandler.LeaveShopList)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/99999/leave", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member1.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	shoplistHandler.LeaveShopList(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"code": "SHP_00001", "error": "Shoplist not found.",
	}, response)
}

func TestRequestShopListShareCodeOwner(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", shoplistHandler.RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.RequestShopListShareCode(c)

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
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Create initial share code
	initialShareCode := db.ShoplistShareCode{
		ID:         1,
		ShopListID: testShoplist.ID,
		Code:       "OLD123",
		Expiry:     time.Now().Add(24 * time.Hour),
	}
	err = testConn.GetDB().Create(&initialShareCode).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", shoplistHandler.RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.RequestShopListShareCode(c)

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
	err = testConn.GetDB().Model(&db.ShoplistShareCode{}).
		Where("shop_list_id = ?", testShoplist.ID).
		Count(&shareCodeCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), shareCodeCount, "Should have exactly one share code")

	// Verify new share code is saved
	var newShareCode db.ShoplistShareCode
	err = testConn.GetDB().Where("shop_list_id = ?", testShoplist.ID).First(&newShareCode).Error
	assert.NoError(t, err)
	assert.Equal(t, response["share_code"], newShareCode.Code, "Share code in database should match response")
}

func TestRequestShopListShareCodeMember(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	member := db.User{
		ID:         "member-123",
		Nickname:   "Member",
		PostalCode: "238802",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&member).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Add member to shoplist
	shoplistMember := db.ShoplistMember{
		ID:         2,
		ShopListID: testShoplist.ID,
		MemberID:   member.ID,
	}
	err = testConn.GetDB().Create(&shoplistMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", shoplistHandler.RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", member.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"code": "SHP_00002", "error": "Only the owner can perform this action.",
	}, response)
}

func TestRequestShopListShareCodeNonMember(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test users
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	nonMember := db.User{
		ID:         "non-member-123",
		Nickname:   "Non-Member",
		PostalCode: "238803",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)
	err = testConn.GetDB().Create(&nonMember).Error
	assert.NoError(t, err)

	// Create test shoplist
	testShoplist := db.Shoplist{
		ID:      1,
		OwnerID: owner.ID,
		Name:    "Test Shoplist",
	}
	err = testConn.GetDB().Create(&testShoplist).Error
	assert.NoError(t, err)

	// Add owner as member to shoplist
	ownerMember := db.ShoplistMember{
		ID:         1,
		ShopListID: testShoplist.ID,
		MemberID:   owner.ID,
	}
	err = testConn.GetDB().Create(&ownerMember).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", shoplistHandler.RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/1/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", nonMember.ID)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	shoplistHandler.RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"code": "SHP_00001", "error": "Shoplist not found.",
	}, response)
}

func TestRequestShopListShareCodeNonExistent(t *testing.T) {
	// Setup test database
	shoplistHandler, testConn := setUpShoplistTestEnv(t)

	// Create test user
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Owner",
		PostalCode: "238801",
	}
	err := testConn.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/shoplist/:id/share-code", shoplistHandler.RequestShopListShareCode)

	// Create request
	req, _ := http.NewRequest("POST", "/shoplist/99999/share-code", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", owner.ID)
	c.Params = []gin.Param{{Key: "id", Value: "99999"}}

	shoplistHandler.RequestShopListShareCode(c)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"code": "SHP_00001", "error": "Shoplist not found.",
	}, response)
}
