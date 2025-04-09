package testutil

import (
	"testing"

	"netherrealmstudio.com/aishoppercore/m/db"
	"netherrealmstudio.com/aishoppercore/m/logger"
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

func SetupTestLogger() {
	err := logger.Init("test")
	if err != nil {
		panic(err)
	}
}

func TeardownTestLogger() {
	logger.Close()
}

func SetupTestEnv(t *testing.T) *db.MySQLConnectionPool {
	SetupTestLogger()
	testDBConn := SetupTestDB(t)

	t.Cleanup(TeardownTestLogger)
	return testDBConn
}
