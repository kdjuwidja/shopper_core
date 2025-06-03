package bizshoplist

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

func TestGetShoplistItemsByUserId(t *testing.T) {
	dbPool := testutil.SetupTestEnv(t)
	setupSearchTestData(t, dbPool)
	biz := InitializeShoplistBiz(*dbPool)

	tests := []struct {
		name              string
		userID            string
		expectedShoplists []*bizmodels.Shoplist
		expectedError     *ShoplistError
	}{
		{
			name:   "successful get all items for user",
			userID: "test_user",
			expectedShoplists: []*bizmodels.Shoplist{
				{
					ID:            1,
					Name:          "Test Shoplist 1",
					OwnerID:       "test_user",
					OwnerNickname: "Test User",
					Items: []bizmodels.ShoplistItem{
						{ID: 1, ShopListID: 1, ItemName: "Item 1", BrandName: "Brand 1", ExtraInfo: "Info 1", IsBought: false},
						{ID: 2, ShopListID: 1, ItemName: "Item 2", BrandName: "Brand 2", ExtraInfo: "Info 2", IsBought: true},
						{ID: 3, ShopListID: 1, ItemName: "Item 3", BrandName: "Brand 3", ExtraInfo: "Info 3", IsBought: false},
					},
				},
				{
					ID:            3,
					Name:          "Test Shoplist 3",
					OwnerID:       "test_user2",
					OwnerNickname: "Test User 2",
					Items: []bizmodels.ShoplistItem{
						{ID: 6, ShopListID: 3, ItemName: "Shared Item 1", BrandName: "Shared Brand 1", ExtraInfo: "Shared Info 1", IsBought: false},
						{ID: 7, ShopListID: 3, ItemName: "Shared Item 2", BrandName: "Shared Brand 2", ExtraInfo: "Shared Info 2", IsBought: true},
					},
				},
			},
			expectedError: nil,
		},
		{
			name:   "successful get all items for second user",
			userID: "test_user2",
			expectedShoplists: []*bizmodels.Shoplist{
				{
					ID:            2,
					Name:          "Test Shoplist 2",
					OwnerID:       "test_user2",
					OwnerNickname: "Test User 2",
					Items: []bizmodels.ShoplistItem{
						{ID: 4, ShopListID: 2, ItemName: "Item 4", BrandName: "Brand 4", ExtraInfo: "Info 4", IsBought: false},
						{ID: 5, ShopListID: 2, ItemName: "Item 5", BrandName: "Brand 5", ExtraInfo: "Info 5", IsBought: true},
					},
				},
				{
					ID:            3,
					Name:          "Test Shoplist 3",
					OwnerID:       "test_user2",
					OwnerNickname: "Test User 2",
					Items: []bizmodels.ShoplistItem{
						{ID: 6, ShopListID: 3, ItemName: "Shared Item 1", BrandName: "Shared Brand 1", ExtraInfo: "Shared Info 1", IsBought: false},
						{ID: 7, ShopListID: 3, ItemName: "Shared Item 2", BrandName: "Shared Brand 2", ExtraInfo: "Shared Info 2", IsBought: true},
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shoplists, err := biz.GetAllShoplistAndItemsForUser(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, shoplists)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, len(tt.expectedShoplists), len(shoplists))
				for i, expectedShoplist := range tt.expectedShoplists {
					assert.Equal(t, expectedShoplist.ID, shoplists[i].ID)
					assert.Equal(t, expectedShoplist.Name, shoplists[i].Name)
					assert.Equal(t, expectedShoplist.OwnerID, shoplists[i].OwnerID)
					assert.Equal(t, expectedShoplist.OwnerNickname, shoplists[i].OwnerNickname)

					// Compare items
					assert.Equal(t, len(expectedShoplist.Items), len(shoplists[i].Items))
					for j, expectedItem := range expectedShoplist.Items {
						assert.Equal(t, expectedItem.ID, shoplists[i].Items[j].ID)
						assert.Equal(t, expectedItem.ShopListID, shoplists[i].Items[j].ShopListID)
						assert.Equal(t, expectedItem.ItemName, shoplists[i].Items[j].ItemName)
						assert.Equal(t, expectedItem.BrandName, shoplists[i].Items[j].BrandName)
						assert.Equal(t, expectedItem.ExtraInfo, shoplists[i].Items[j].ExtraInfo)
						assert.Equal(t, expectedItem.IsBought, shoplists[i].Items[j].IsBought)
					}
				}
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

func TestGetShoplistAndItems(t *testing.T) {
	dbPool := testutil.SetupTestEnv(t)
	setupSearchTestData(t, dbPool)
	biz := InitializeShoplistBiz(*dbPool)

	tests := []struct {
		name             string
		userID           string
		shoplistID       int
		expectedShoplist *bizmodels.Shoplist
		expectedError    *ShoplistError
	}{
		{
			name:       "successful get shoplist and items for owner",
			userID:     "test_user",
			shoplistID: 1,
			expectedShoplist: &bizmodels.Shoplist{
				ID:            1,
				Name:          "Test Shoplist 1",
				OwnerID:       "test_user",
				OwnerNickname: "Test User",
				Items: []bizmodels.ShoplistItem{
					{ID: 1, ShopListID: 1, ItemName: "Item 1", BrandName: "Brand 1", ExtraInfo: "Info 1", IsBought: false},
					{ID: 2, ShopListID: 1, ItemName: "Item 2", BrandName: "Brand 2", ExtraInfo: "Info 2", IsBought: true},
					{ID: 3, ShopListID: 1, ItemName: "Item 3", BrandName: "Brand 3", ExtraInfo: "Info 3", IsBought: false},
				},
			},
			expectedError: nil,
		},
		{
			name:       "successful get shared shoplist and items for member",
			userID:     "test_user",
			shoplistID: 3,
			expectedShoplist: &bizmodels.Shoplist{
				ID:            3,
				Name:          "Test Shoplist 3",
				OwnerID:       "test_user2",
				OwnerNickname: "Test User 2",
				Items: []bizmodels.ShoplistItem{
					{ID: 6, ShopListID: 3, ItemName: "Shared Item 1", BrandName: "Shared Brand 1", ExtraInfo: "Shared Info 1", IsBought: false},
					{ID: 7, ShopListID: 3, ItemName: "Shared Item 2", BrandName: "Shared Brand 2", ExtraInfo: "Shared Info 2", IsBought: true},
				},
			},
			expectedError: nil,
		},
		{
			name:             "shoplist not found",
			userID:           "test_user",
			shoplistID:       999,
			expectedShoplist: nil,
			expectedError:    NewShoplistError(ShoplistNotFound, "Shoplist not found."),
		},
		{
			name:             "user not a member",
			userID:           "test_user",
			shoplistID:       2,
			expectedShoplist: nil,
			expectedError:    NewShoplistError(ShoplistNotFound, "Shoplist not found."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shoplist, err := biz.GetShoplistAndItems(context.Background(), tt.userID, tt.shoplistID)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.ErrCode, err.ErrCode)
				assert.Nil(t, shoplist)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedShoplist.ID, shoplist.ID)
				assert.Equal(t, tt.expectedShoplist.Name, shoplist.Name)
				assert.Equal(t, tt.expectedShoplist.OwnerID, shoplist.OwnerID)
				assert.Equal(t, tt.expectedShoplist.OwnerNickname, shoplist.OwnerNickname)

				// Compare items
				assert.Equal(t, len(tt.expectedShoplist.Items), len(shoplist.Items))
				for i, expectedItem := range tt.expectedShoplist.Items {
					assert.Equal(t, expectedItem.ID, shoplist.Items[i].ID)
					assert.Equal(t, expectedItem.ShopListID, shoplist.Items[i].ShopListID)
					assert.Equal(t, expectedItem.ItemName, shoplist.Items[i].ItemName)
					assert.Equal(t, expectedItem.BrandName, shoplist.Items[i].BrandName)
					assert.Equal(t, expectedItem.ExtraInfo, shoplist.Items[i].ExtraInfo)
					assert.Equal(t, expectedItem.IsBought, shoplist.Items[i].IsBought)
				}
			}
		})
	}
}
