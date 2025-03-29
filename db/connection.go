package db

import (
	"fmt"
	"sync"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"netherrealmstudio.com/aishoppercore/m/model"
)

type Config struct {
	User         string
	Password     string
	Host         string
	Port         string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

var (
	db   *gorm.DB
	once sync.Once
)

// Initialize sets up the database connection
func Initialize(config *Config) error {
	var err error
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.User, config.Password, config.Host, config.Port, config.DBName)

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			return
		}

		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	})
	return err
}

func AutoMigrate() error {
	err := db.AutoMigrate(
		&model.ShoplistItem{},
		&model.ShoplistMember{},
		&model.ShoplistShareCode{},
		&model.Shoplist{},
		&model.User{},
	)

	if err != nil {
		return err
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("error getting underlying *sql.DB: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// InitializeTestDB sets up a connection to a local MySQL database for testing
func InitializeTestDB(t *testing.T) (*gorm.DB, error) {
	dsn := "ai_shopper_dev:password@tcp(localhost:4306)/test_db?parseTime=true"
	testDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	// Override the global db variable for testing
	db = testDB

	return testDB, nil
}

// CloseTestDB resets the database connection
func CloseTestDB() {
	if db != nil {
		// Drop all tables
		db.Migrator().DropTable(
			&model.ShoplistItem{},
			&model.ShoplistMember{},
			&model.ShoplistShareCode{},
			&model.Shoplist{},
			&model.User{},
		)
		db = nil
	}
}
