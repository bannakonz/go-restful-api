package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitDB initializes the database connection.
func InitDB() error {
	logLevel := logger.Warn

	// Show SQL queries when Gin is running in debug mode.
	if gin.Mode() == gin.DebugMode {
		logLevel = logger.Info
	}

	var err error

	db, err = gorm.Open(sqlite.Open("./gorm.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	return nil
}

// GetDB returns the initialized GORM database instance.
func GetDB() *gorm.DB {
	return db
}

// CloseDB closes the underlying database connection pool.
func CloseDB() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL database: %w", err)
	}

	return sqlDB.Close()
}
