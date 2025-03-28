package testutil

import (
	"testing"

	"gorm.io/gorm"
	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/model"
)

// SetupTestDB initializes a test database and migrates the schema
func SetupTestDB(t *testing.T) *gorm.DB {
	testDB, err := db.InitializeTestDB(t)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Migrate the schema
	err = testDB.AutoMigrate(&model.User{}, &model.Shoplist{}, &model.ShoplistItem{}, &model.ShoplistMember{}, &model.ShoplistShareCode{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return testDB
}

// TeardownTestDB cleans up the test database
func TeardownTestDB() {
	db.CloseTestDB()
}
