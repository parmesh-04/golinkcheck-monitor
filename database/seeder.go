
package database

import (
	"log/slog"

	"golang.org/x/crypto/bcrypt" // NEW: Import bcrypt for password hashing
	"gorm.io/gorm"
)

// Seed runs the database seeders to populate it with an initial admin user and sample data.
func Seed(db *gorm.DB) {
	// We check if any users exist. If they do, we assume the DB is already seeded.
	var userCount int64
	db.Model(&User{}).Count(&userCount)

	if userCount > 0 {
		// You can uncomment this for debugging if you want.
		// slog.Info("Database already contains users. Skipping seed process.")
		return
	}

	slog.Info("Database is empty. Seeding with an initial admin user and test data...")

	// --- Step 1: Create an Admin User ---
	// We must create a user first, so we have a valid UserID to assign to the monitors.
	
	// Use bcrypt to create a secure password hash.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password for seeder", "error", err)
		return
	}

	adminUser := User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "admin", // Explicitly set the role to 'admin'
	}

	// Create the admin user in the database.
	if err := db.Create(&adminUser).Error; err != nil {
		slog.Error("Failed to seed admin user", "error", err)
		return
	}
	slog.Info("Admin user created successfully", "user_id", adminUser.ID, "email", adminUser.Email)


	// --- Step 2: Create Monitors for the Admin User ---
	// Now we use `adminUser.ID` as the UserID for all the monitors.
	monitors := []Monitor{
		{URL: "https://www.google.com", IntervalSec: 60, Active: true, UserID: adminUser.ID},
		{URL: "https://www.github.com", IntervalSec: 60, Active: true, UserID: adminUser.ID},
		{URL: "https://www.cloudflare.com", IntervalSec: 120, Active: true, UserID: adminUser.ID},
		{URL: "https://httpstat.us/503", IntervalSec: 300, Active: true, UserID: adminUser.ID}, // A site that is always down
		{URL: "https://www.inactive.com", IntervalSec: 999, Active: false, UserID: adminUser.ID}, // An inactive monitor
	}

	// Use GORM's Create method to perform a batch insert of all monitor records.
	if err := db.Create(&monitors).Error; err != nil {
		slog.Error("Failed to seed monitors", "error", err)
	} else {
		slog.Info("Database seeded successfully with sample monitors", "records_created", len(monitors))
	}
}