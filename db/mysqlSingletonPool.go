package db

import (
	"sync"

	"github.com/kdjuwidja/aishoppercommon/logger"
)

var (
	mysqlOnce sync.Once
	mysqlConn *MySQLConnectionPool
)

func InitializeMySQLConnPoolSingleton(conn DBConnectionPool) (*MySQLConnectionPool, error) {
	mysqlOnce.Do(func() {
		var ok bool
		mysqlConn, ok = conn.(*MySQLConnectionPool)
		if !ok {
			logger.Panic("Failed to initialize database connection: invalid connection type.")
		}

		err := mysqlConn.Initialize()
		if err != nil {
			logger.Panicf("Failed to initialize database connection: %v", err)
		}
	})

	return mysqlConn, nil
}

func GetInstance() *MySQLConnectionPool {
	return mysqlConn
}
