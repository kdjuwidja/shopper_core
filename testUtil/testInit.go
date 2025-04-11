package testutil

import (
	"testing"

	"github.com/kdjuwidja/aishoppercommon/logger"
	"netherrealmstudio.com/aishoppercore/m/db"
)

// SetupTestDB initializes a test database and migrates the schema
func SetupTestDB(t *testing.T) *db.MySQLConnectionPool {
	testConn := &db.MySQLConnectionPool{}
	testConn.Configure("ai_shopper_dev", "password", "localhost", "4306", "test_db", 25, 10)

	err := testConn.Initialize()
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
	testDBConn := SetupTestDB(t)

	return testDBConn
}
