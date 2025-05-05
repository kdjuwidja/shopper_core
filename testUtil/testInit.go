package testutil

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/db"
	"github.com/kdjuwidja/aishoppercommon/logger"
	dbmodel "netherealmstudio.com/m/v2/db"
)

// SetupTestDB initializes a test database and migrates the schema
func SetupTestDB(t *testing.T, models []interface{}) *db.MySQLConnectionPool {
	testConn, err := db.InitializeMySQLConnectionPool("ai_shopper_dev", "password", "localhost", "4306", "test_db", 25, 10, models)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	testConn.DropTables()
	err = testConn.AutoMigrate()
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return testConn
}

// TeardownTestDB cleans up the test database
func TeardownTestDB(testConn *db.MySQLConnectionPool) {
	testConn.Close()
}

func SetupTestEnv(t *testing.T) *db.MySQLConnectionPool {
	logger.SetServiceName("test")
	logger.SetLevel("trace")
	models := []interface{}{
		&dbmodel.Shoplist{},
		&dbmodel.ShoplistItem{},
		&dbmodel.ShoplistMember{},
		&dbmodel.ShoplistShareCode{},
		&dbmodel.User{},
	}
	testDBConn := SetupTestDB(t, models)

	return testDBConn
}
