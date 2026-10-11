package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens GORMs PostgreSQL connection
// The caller should ping the DB before serving any traffic
func Connect(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	connection, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil,fmt.Errorf("open PostgreSQL connection: %w", err)
	}
	sqlDB, err := connection.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying SQL connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return connection, nil
}
