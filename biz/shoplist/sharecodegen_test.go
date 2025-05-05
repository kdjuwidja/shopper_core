package bizshoplist

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"netherealmstudio.com/m/v2/db"
	testutil "netherealmstudio.com/m/v2/testUtil"
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
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Test Owner",
		PostalCode: "238801",
	}
	err := testDB.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create a new shoplist for this test
	shoplist := db.Shoplist{
		OwnerID: owner.ID,
		Name:    "Test Shoplist 1",
	}
	err = testDB.GetDB().Create(&shoplist).Error
	assert.NoError(t, err)

	code := GenerateShareCode(6)
	shareCode := db.ShoplistShareCode{
		ShopListID: shoplist.ID,
		Code:       code,
		Expiry:     time.Now().Add(24 * time.Hour),
	}
	err = testDB.GetDB().Create(&shareCode).Error
	assert.NoError(t, err)

	result := VerifyShareCodeFromDB(testDB.GetDB(), code)
	assert.False(t, result)
}

func TestVerifyShareCodeExistsAndExpired(t *testing.T) {
	testDB := testutil.SetupTestEnv(t)

	// Create test user
	owner := db.User{
		ID:         "owner-123",
		Nickname:   "Test Owner",
		PostalCode: "238801",
	}
	err := testDB.GetDB().Create(&owner).Error
	assert.NoError(t, err)

	// Create a new shoplist for this test
	shoplist := db.Shoplist{
		OwnerID: owner.ID,
		Name:    "Test Shoplist 2",
	}
	err = testDB.GetDB().Create(&shoplist).Error
	assert.NoError(t, err)

	code := GenerateShareCode(6)
	shareCode := db.ShoplistShareCode{
		ShopListID: shoplist.ID,
		Code:       code,
		Expiry:     time.Now().Add(-1 * time.Hour),
	}
	err = testDB.GetDB().Create(&shareCode).Error
	assert.NoError(t, err)

	result := VerifyShareCodeFromDB(testDB.GetDB(), code)
	assert.True(t, result)
}

func TestVerifyShareCodeValidFormatNotInDB(t *testing.T) {
	testDB := testutil.SetupTestEnv(t)

	code := GenerateShareCode(6)
	result := VerifyShareCodeFromDB(testDB.GetDB(), code)
	assert.True(t, result)
}

func TestVerifyShareCodeNonExistent(t *testing.T) {
	testDB := testutil.SetupTestEnv(t)

	result := VerifyShareCodeFromDB(testDB.GetDB(), "NONEXISTENT")
	assert.True(t, result)
}
