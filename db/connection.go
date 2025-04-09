package db

import "gorm.io/gorm"

type DBConnectionPool interface {
	Initialize() error
	AutoMigrate() error
	DropTables() error
	GetDB() *gorm.DB
	Close() error
}
