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

func TestVerifyShareCodeExistsAndNotExpired(t *testing.T) {
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

	// Create a new shoplist for this test
	shoplist := model.Shoplist{
		OwnerID: owner.ID,
		Name:    "Test Shoplist 1",
	}
	err = testDB.Create(&shoplist).Error
	assert.NoError(t, err)

	code := GenerateShareCode(6)
	shareCode := model.ShoplistShareCode{
		ShopListID: shoplist.ID,
		Code:       code,
		Expiry:     time.Now().Add(24 * time.Hour),
	}
	err = testDB.Create(&shareCode).Error
	assert.NoError(t, err)

	result := VerifyShareCodeFromDB(testDB, code)
	assert.False(t, result)
}

func TestVerifyShareCodeExistsAndExpired(t *testing.T) {
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

	// Create a new shoplist for this test
	shoplist := model.Shoplist{
		OwnerID: owner.ID,
		Name:    "Test Shoplist 2",
	}
	err = testDB.Create(&shoplist).Error
	assert.NoError(t, err)

	code := GenerateShareCode(6)
	shareCode := model.ShoplistShareCode{
		ShopListID: shoplist.ID,
		Code:       code,
		Expiry:     time.Now().Add(-1 * time.Hour),
	}
	err = testDB.Create(&shareCode).Error
	assert.NoError(t, err)

	result := VerifyShareCodeFromDB(testDB, code)
	assert.True(t, result)
}

func TestVerifyShareCodeValidFormatNotInDB(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	code := GenerateShareCode(6)
	result := VerifyShareCodeFromDB(testDB, code)
	assert.True(t, result)
}

func TestVerifyShareCodeNonExistent(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	t.Cleanup(testutil.TeardownTestDB)

	result := VerifyShareCodeFromDB(testDB, "NONEXISTENT")
	assert.True(t, result)
}
