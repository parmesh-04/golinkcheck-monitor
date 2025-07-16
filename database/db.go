// database/db.go

package database

import (
	"fmt"
	"log/slog" 
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/parmesh-04/golinkcheck-monitor/config"
)

// InitDB initializes the database connection and runs migrations.
func InitDB(cfg config.Config) (*gorm.DB, error) {
	slog.Info("Initializing database connection...") // <-- CHANGED

	var db *gorm.DB
	var err error

	if strings.HasPrefix(cfg.DatabaseURL, "sqlite:") {
		dbPath := strings.TrimPrefix(cfg.DatabaseURL, "sqlite:")
		// Add more context to the log!
		slog.Info("Connecting to SQLite database", "path", dbPath) // <-- CHANGED
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	} else {
		return nil, fmt.Errorf("unsupported database type. only 'sqlite:' is supported for now")
	}

	if err != nil {
		// Add structured error logging
		slog.Error("Failed to connect to database", "error", err) // <-- CHANGED
		return nil, err
	}

	slog.Info("Database connection established.") // <-- CHANGED

	slog.Info("Running database migrations...") // <-- CHANGED
	// We are migrating the Monitor and CheckResult models
	err = db.AutoMigrate(&Monitor{}, &CheckResult{})
	if err != nil {
		slog.Error("Failed to run database migrations", "error", err) // <-- CHANGED
		return nil, err
	}

	slog.Info("Database migrations completed successfully.") // <-- CHANGED
	return db, nil
}