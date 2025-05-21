package bizshoplist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testutil "netherealmstudio.com/m/v2/testUtil"
)

func TestGetShoplistMembers(t *testing.T) {
	dbPool := testutil.SetupTestEnv(t)
	setupSearchTestData(t, dbPool)
	biz := InitializeShoplistBiz(*dbPool)

	tests := []struct {
		name            string
		userID          string
		shoplistID      int
		expectedMembers []ShoplistMember
		expectedError   *ShoplistError
	}{
		{
			name:       "successful get members for owner",
			userID:     "test_user",
			shoplistID: 1,
			expectedMembers: []ShoplistMember{
				{
					ID:       "test_user",
					Nickname: "Test User",
				},
			},
			expectedError: nil,
		},
		{
			name:       "successful get members for shared shoplist",
			userID:     "test_user",
			shoplistID: 3,
			expectedMembers: []ShoplistMember{
				{
					ID:       "test_user",
					Nickname: "Test User",
				},
				{
					ID:       "test_user2",
					Nickname: "Test User 2",
				},
			},
			expectedError: nil,
		},
		{
			name:            "shoplist not found",
			userID:          "test_user",
			shoplistID:      999,
			expectedMembers: nil,
			expectedError:   NewShoplistError(ShoplistNotFound, "record not found"),
		},
		{
			name:            "user not a member",
			userID:          "test_user",
			shoplistID:      2,
			expectedMembers: nil,
			expectedError:   NewShoplistError(ShoplistNotMember, "User is not a member of the shoplist."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members, err := biz.GetShoplistMembers(tt.userID, tt.shoplistID)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.ErrCode, err.ErrCode)
				assert.Nil(t, members)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, len(tt.expectedMembers), len(members))

				// Create a map of expected members for easier comparison
				expectedMemberMap := make(map[string]ShoplistMember)
				for _, member := range tt.expectedMembers {
					expectedMemberMap[member.ID] = member
				}

				// Verify each member
				for _, member := range members {
					expectedMember, exists := expectedMemberMap[member.ID]
					assert.True(t, exists, "Unexpected member ID: %s", member.ID)
					assert.Equal(t, expectedMember.Nickname, member.Nickname)
				}
			}
		})
	}
}
