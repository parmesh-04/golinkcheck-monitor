
package database

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/parmesh-04/golinkcheck-monitor/config" 
)

// InitDB initializes the database connection and runs migrations.
// It returns a pointer to the GORM DB object or an error.
func InitDB(cfg config.Config) (*gorm.DB, error) {
	log.Println("Initializing database connection...")

	var db *gorm.DB
	var err error

	// For this project, we'll primarily handle SQLite.
	// We check the prefix of the DatabaseURL from our config.
	if strings.HasPrefix(cfg.DatabaseURL, "sqlite:") {
		// GORM's SQLite driver expects just the file path, so we trim the prefix.
		dbPath := strings.TrimPrefix(cfg.DatabaseURL, "sqlite:")
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	} else {
		// In a real application, you might have another 'else if' block here for "postgres:".
		return nil, fmt.Errorf("unsupported database type. only 'sqlite:' is supported for now")
	}

	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	log.Println("Database connection established.")

	// Run database migrations.
	// AutoMigrate will create tables, missing columns, and missing indexes.
	// It will NOT delete unused columns, to protect your data.
	log.Println("Running database migrations...")
	err = db.AutoMigrate(&Monitor{}, &CheckResult{})
	if err != nil {
		log.Printf("Failed to run database migrations: %v", err)
		return nil, err
	}

	log.Println("Database migrations completed successfully.")
	return db, nil
}