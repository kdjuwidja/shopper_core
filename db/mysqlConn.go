package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"netherrealmstudio.com/aishoppercore/m/logger"
	"netherrealmstudio.com/aishoppercore/m/model"
)

type MySQLConnectionPool struct {
	User         string
	Password     string
	Host         string
	Port         string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int

	db *gorm.DB
}

func (c *MySQLConnectionPool) Configure(user, password, host, port, dbName string, maxOpenConns, maxIdleConns int) {
	c.User = user
	c.Password = password
	c.Host = host
	c.Port = port
	c.DBName = dbName
	c.MaxOpenConns = maxOpenConns
	c.MaxIdleConns = maxIdleConns
}

func (c *MySQLConnectionPool) Initialize() error {
	if c.User == "" || c.Password == "" || c.Host == "" || c.Port == "" || c.DBName == "" {
		return fmt.Errorf("missing required configuration parameters")
	}

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.DBName)
	logger.Debugf("dsn: %s", dsn)

	c.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	return err
}

func (c *MySQLConnectionPool) AutoMigrate() error {
	err := c.db.AutoMigrate(
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

func (c *MySQLConnectionPool) DropTables() error {
	return c.db.Migrator().DropTable(
		&model.ShoplistItem{},
		&model.ShoplistMember{},
		&model.ShoplistShareCode{},
		&model.Shoplist{},
		&model.User{},
	)
}

func (c *MySQLConnectionPool) GetDB() *gorm.DB {
	return c.db
}

func (c *MySQLConnectionPool) Close() error {
	if c.db != nil {
		sqlDB, err := c.db.DB()
		if err != nil {
			return fmt.Errorf("error getting underlying *sql.DB: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}
