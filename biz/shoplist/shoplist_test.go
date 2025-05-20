package bizshoplist

import (
	"context"
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/stretchr/testify/assert"
	dbmodel "netherealmstudio.com/m/v2/db"
	testutil "netherealmstudio.com/m/v2/testUtil"
)

func setupSearchTestData(t *testing.T, dbPool *db.MySQLConnectionPool) {
	// Create test users
	users := []dbmodel.User{
		{
			ID:         "test_user",
			Nickname:   "Test User",
			PostalCode: "238801",
		},
		{
			ID:         "test_user2",
			Nickname:   "Test User 2",
			PostalCode: "238802",
		},
	}
	for _, user := range users {
		err := dbPool.GetDB().Create(&user).Error
		assert.NoError(t, err)
	}

	// Create test shoplists
	shoplists := []dbmodel.Shoplist{
		{
			ID:      1,
			OwnerID: users[0].ID,
			Name:    "Test Shoplist 1",
		},
		{
			ID:      2,
			OwnerID: users[1].ID,
			Name:    "Test Shoplist 2",
		},
		{
			ID:      3,
			OwnerID: users[1].ID,
			Name:    "Test Shoplist 3",
		},
	}
	for _, shoplist := range shoplists {
		err := dbPool.GetDB().Create(&shoplist).Error
		assert.NoError(t, err)
	}

	// Add members to shoplists
	members := []dbmodel.ShoplistMember{
		{
			ID:         1,
			ShopListID: shoplists[0].ID,
			MemberID:   users[0].ID,
		},
		{
			ID:         2,
			ShopListID: shoplists[1].ID,
			MemberID:   users[1].ID,
		},
		{
			ID:         3,
			ShopListID: shoplists[2].ID,
			MemberID:   users[1].ID,
		},
		{
			ID:         4,
			ShopListID: shoplists[2].ID,
			MemberID:   users[0].ID, // First user is member of second user's third shoplist
		},
	}
	for _, member := range members {
		err := dbPool.GetDB().Create(&member).Error
		assert.NoError(t, err)
	}

	// Add test items to shoplists
	items := []dbmodel.ShoplistItem{
		// Items for first user's shoplist
		{
			ID:         1,
			ShopListID: shoplists[0].ID,
			ItemName:   "Item 1",
			BrandName:  "Brand 1",
			ExtraInfo:  "Info 1",
			IsBought:   false,
		},
		{
			ID:         2,
			ShopListID: shoplists[0].ID,
			ItemName:   "Item 2",
			BrandName:  "Brand 2",
			ExtraInfo:  "Info 2",
			IsBought:   true,
		},
		{
			ID:         3,
			ShopListID: shoplists[0].ID,
			ItemName:   "Item 3",
			BrandName:  "Brand 3",
			ExtraInfo:  "Info 3",
			IsBought:   false,
		},
		// Items for second user's second shoplist
		{
			ID:         4,
			ShopListID: shoplists[1].ID,
			ItemName:   "Item 4",
			BrandName:  "Brand 4",
			ExtraInfo:  "Info 4",
			IsBought:   false,
		},
		{
			ID:         5,
			ShopListID: shoplists[1].ID,
			ItemName:   "Item 5",
			BrandName:  "Brand 5",
			ExtraInfo:  "Info 5",
			IsBought:   true,
		},
		// Items for shared shoplist
		{
			ID:         6,
			ShopListID: shoplists[2].ID,
			ItemName:   "Shared Item 1",
			BrandName:  "Shared Brand 1",
			ExtraInfo:  "Shared Info 1",
			IsBought:   false,
		},
		{
			ID:         7,
			ShopListID: shoplists[2].ID,
			ItemName:   "Shared Item 2",
			BrandName:  "Shared Brand 2",
			ExtraInfo:  "Shared Info 2",
			IsBought:   true,
		},
	}

	for _, item := range items {
		err := dbPool.GetDB().Create(&item).Error
		assert.NoError(t, err)
	}
}

func TestGetShoplistItemsByUserId(t *testing.T) {
	dbPool := testutil.SetupTestEnv(t)
	setupSearchTestData(t, dbPool)
	biz := InitializeShoplistBiz(*dbPool)

	tests := []struct {
		name          string
		userID        string
		expectedItems []ShoplistItem
		expectedError *ShoplistError
	}{
		{
			name:   "successful get all items for user",
			userID: "test_user",
			expectedItems: []ShoplistItem{
				{ID: 1, ShopListID: 1, ItemName: "Item 1", BrandName: "Brand 1"},
				{ID: 2, ShopListID: 1, ItemName: "Item 2", BrandName: "Brand 2"},
				{ID: 3, ShopListID: 1, ItemName: "Item 3", BrandName: "Brand 3"},
				{ID: 6, ShopListID: 3, ItemName: "Shared Item 1", BrandName: "Shared Brand 1"},
				{ID: 7, ShopListID: 3, ItemName: "Shared Item 2", BrandName: "Shared Brand 2"},
			},
			expectedError: nil,
		},
		{
			name:   "successful get all items for second user",
			userID: "test_user2",
			expectedItems: []ShoplistItem{
				{ID: 4, ShopListID: 2, ItemName: "Item 4", BrandName: "Brand 4"},
				{ID: 5, ShopListID: 2, ItemName: "Item 5", BrandName: "Brand 5"},
				{ID: 6, ShopListID: 3, ItemName: "Shared Item 1", BrandName: "Shared Brand 1"},
				{ID: 7, ShopListID: 3, ItemName: "Shared Item 2", BrandName: "Shared Brand 2"},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := biz.GetAllShoplistAndItemsForUser(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, items)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedItems, items)
			}
		})
	}
}

func TestGetShoplistWithMembers(t *testing.T) {
	dbPool := testutil.SetupTestEnv(t)
	setupSearchTestData(t, dbPool)
	biz := InitializeShoplistBiz(*dbPool)

	tests := []struct {
		name          string
		shoplistID    int
		expectedData  *ShoplistData
		expectedError *ShoplistError
	}{
		{
			name:       "successful get shoplist with members",
			shoplistID: 1,
			expectedData: &ShoplistData{
				ShopListID: 1,
				OwnerID:    "test_user",
				Members: map[string]struct {
					MemberID string
					Nickname string
				}{
					"test_user": {
						MemberID: "test_user",
						Nickname: "Test User",
					},
				},
			},
			expectedError: nil,
		},
		{
			name:       "successful get shared shoplist with members",
			shoplistID: 3,
			expectedData: &ShoplistData{
				ShopListID: 3,
				OwnerID:    "test_user2",
				Members: map[string]struct {
					MemberID string
					Nickname string
				}{
					"test_user2": {
						MemberID: "test_user2",
						Nickname: "Test User 2",
					},
					"test_user": {
						MemberID: "test_user",
						Nickname: "Test User",
					},
				},
			},
			expectedError: nil,
		},
		{
			name:          "shoplist not found",
			shoplistID:    999,
			expectedData:  nil,
			expectedError: NewShoplistError(ShoplistNotFound, "record not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := biz.GetShoplistWithMembers(tt.shoplistID)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.ErrCode, err.ErrCode)
				assert.Nil(t, data)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedData.ShopListID, data.ShopListID)
				assert.Equal(t, tt.expectedData.OwnerID, data.OwnerID)
				assert.Equal(t, len(tt.expectedData.Members), len(data.Members))
				for memberID, member := range tt.expectedData.Members {
					assert.Contains(t, data.Members, memberID)
					assert.Equal(t, member.MemberID, data.Members[memberID].MemberID)
					assert.Equal(t, member.Nickname, data.Members[memberID].Nickname)
				}
			}
		})
	}
}
