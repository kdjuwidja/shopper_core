package bizmatch

import (
	"context"
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/stretchr/testify/assert"
	bizmodels "netherealmstudio.com/m/v2/biz"
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

func TestGetShoplistItems(t *testing.T) {
	dbPool := testutil.SetupTestEnv(t)
	setupSearchTestData(t, dbPool)
	biz := NewMatchShoplistItemsWithFlyerBiz(nil, dbPool)

	tests := []struct {
		name            string
		userID          string
		shoplistItemIds []int
		expectedItems   []bizmodels.ShoplistItem
		expectedError   *MatchError
	}{
		{
			name:            "successful get items from own shoplist",
			userID:          "test_user",
			shoplistItemIds: []int{1, 2, 3},
			expectedItems: []bizmodels.ShoplistItem{
				{ID: 1, ItemName: "Item 1", BrandName: "Brand 1"},
				{ID: 2, ItemName: "Item 2", BrandName: "Brand 2"},
				{ID: 3, ItemName: "Item 3", BrandName: "Brand 3"},
			},
			expectedError: nil,
		},
		{
			name:            "successful get items from shared shoplist",
			userID:          "test_user",
			shoplistItemIds: []int{6, 7},
			expectedItems: []bizmodels.ShoplistItem{
				{ID: 6, ItemName: "Shared Item 1", BrandName: "Shared Brand 1"},
				{ID: 7, ItemName: "Shared Item 2", BrandName: "Shared Brand 2"},
			},
			expectedError: nil,
		},
		{
			name:            "empty item IDs",
			userID:          "test_user",
			shoplistItemIds: []int{},
			expectedItems:   nil,
			expectedError:   NewMatchError(MatchItemIdNotProvided, "No item IDs provided."),
		},
		{
			name:            "no items found",
			userID:          "test_user",
			shoplistItemIds: []int{999},
			expectedItems:   nil,
			expectedError:   NewMatchError(MatchShoplistItemNotFound, "No shoplist items found."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := biz.GetShoplistItems(context.Background(), tt.userID, tt.shoplistItemIds)
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
