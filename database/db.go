
package database

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/driver/postgres" 
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/parmesh-04/golinkcheck-monitor/config"
)

// InitDB initializes the database connection and runs migrations.
func InitDB(cfg config.Config) (*gorm.DB, error) {
	slog.Info("Initializing database connection...")

	var db *gorm.DB
	var err error
	dbURL := cfg.DatabaseURL

	if strings.HasPrefix(dbURL, "sqlite:") {
		
		dbPath := strings.TrimPrefix(dbURL, "sqlite:")
		slog.Info("Connecting to SQLite database", "path", dbPath)
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	} else if strings.HasPrefix(dbURL, "postgres:") || strings.HasPrefix(dbURL, "postgresql:") {
		
		slog.Info("Connecting to PostgreSQL database...")
		// The postgres driver can use the full URL (DSN) directly.
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		

	} else {
		return nil, fmt.Errorf("unsupported database type. only 'sqlite' and 'postgres' are supported")
	}

	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	slog.Info("Database connection established.")

	slog.Info("Running database migrations...")
	err = db.AutoMigrate(&Monitor{}, &CheckResult{})
	if err != nil {
		slog.Error("Failed to run database migrations", "error", err)
		return nil, err
	}

	slog.Info("Database migrations completed successfully.")
	return db, nil
}