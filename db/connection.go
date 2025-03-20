package db

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
