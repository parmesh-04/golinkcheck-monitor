// File: database/db.go
package database

import (
	"fmt"
	"log/slog"

	"github.com/parmesh-04/golinkcheck-monitor/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection using the detailed config struct.
func InitDB(cfg config.Config) (*gorm.DB, error) {
	slog.Info("Initializing PostgreSQL database connection...")

	// Build the Data Source Name (DSN) for PostgreSQL from the config fields.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		slog.Error("Failed to connect to PostgreSQL database", "error", err)
		return nil, err
	}

	slog.Info("Database connection established.")

	slog.Info("Running database migrations...")
	// Ensure the User model is included in the migration.
	err = db.AutoMigrate(&User{}, &Monitor{}, &CheckResult{})
	if err != nil {
		slog.Error("Failed to run database migrations", "error", err)
		return nil, err
	}

	slog.Info("Database migrations completed successfully.")
	return db, nil
}