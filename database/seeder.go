package database

import (
	"log/slog"

	"gorm.io/gorm"
)

// Seed runs the database seeders to populate it with initial data.
func Seed(db *gorm.DB) {
	// We only want to seed if the monitors table is completely empty.
	var count int64
	db.Model(&Monitor{}).Count(&count)

	if count > 0 {
		slog.Info("Database already contains data. Skipping seed process.")
		return
	}

	slog.Info("Database is empty. Seeding with initial test data...")

	// Create a slice of Monitor structs with our default data.
	monitors := []Monitor{
		{URL: "https://www.google.com", IntervalSec: 60, Active: true},
		{URL: "https://www.github.com", IntervalSec: 60, Active: true},
		{URL: "https://www.cloudflare.com", IntervalSec: 120, Active: true},
		{URL: "https://httpstat.us/503", IntervalSec: 300, Active: true}, // A site that is always down
		{URL: "https://www.inactive.com", IntervalSec: 999, Active: false}, // An inactive monitor
	}

	// Use GORM's Create method to perform a batch insert of all records in the slice.
	if err := db.Create(&monitors).Error; err != nil {
		slog.Error("Failed to seed database", "error", err)
	} else {
		slog.Info("Database seeded successfully", "records_created", len(monitors))
	}
}