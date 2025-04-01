package util

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"netherrealmstudio.com/aishoppercore/m/model"
	testutil "netherrealmstudio.com/aishoppercore/m/testUtil"
)

func TestShareCodeGen(t *testing.T) {
	// Call the sharecodegen function
	code := GenerateShareCode(6)

	// Check if the generated code is not empty
	assert.NotEmpty(t, code)

	// Check if the generated code is 6 characters long
	assert.Equal(t, 6, len(code))

	// Check if the generated code contains only uppercase alphanumeric characters
	matched, err := regexp.MatchString("^[A-Z0-9]+$", code)
	if err != nil {
		t.Fatalf("Error while matching regex: %v", err)
	}
	assert.True(t, matched, "Expected share code to contain only uppercase alphanumeric characters")
}

func TestVerifyShareCode(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	// Create test user
	owner := model.User{
		ID:         "owner-123",
		Nickname:   "Test Owner",
		PostalCode: "238801",
	}
	err := testDB.Create(&owner).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		setupShareCode func() string
		expectedResult bool
	}{
		{
			name: "Share code exists and not expired",
			setupShareCode: func() string {
				// Create a new shoplist for this test
				shoplist := model.Shoplist{
					OwnerID: owner.ID,
					Name:    "Test Shoplist 1",
				}
				err := testDB.Create(&shoplist).Error
				assert.NoError(t, err)

				code := GenerateShareCode(6)
				shareCode := model.ShoplistShareCode{
					ShopListID: shoplist.ID,
					Code:       code,
					Expiry:     time.Now().Add(24 * time.Hour),
				}
				err = testDB.Create(&shareCode).Error
				assert.NoError(t, err)
				return code
			},
			expectedResult: false,
		},
		{
			name: "Share code exists and expired",
			setupShareCode: func() string {
				// Create a new shoplist for this test
				shoplist := model.Shoplist{
					OwnerID: owner.ID,
					Name:    "Test Shoplist 2",
				}
				err := testDB.Create(&shoplist).Error
				assert.NoError(t, err)

				code := GenerateShareCode(6)
				shareCode := model.ShoplistShareCode{
					ShopListID: shoplist.ID,
					Code:       code,
					Expiry:     time.Now().Add(-1 * time.Hour),
				}
				err = testDB.Create(&shareCode).Error
				assert.NoError(t, err)
				return code
			},
			expectedResult: true,
		},
		{
			name: "Share code with valid format but not in database",
			setupShareCode: func() string {
				return GenerateShareCode(6)
			},
			expectedResult: true,
		},
		{
			name: "Share code does not exist at all",
			setupShareCode: func() string {
				return "NONEXISTENT"
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := tt.setupShareCode()
			result := VerifyShareCodeFromDB(testDB, code)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
