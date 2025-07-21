// File: database/models.go

package database

import (
	"time"

	"gorm.io/gorm"
)

// NEW: User struct represents a user in the database.
// It will have a one-to-many relationship with Monitors (one user can have many monitors).
type User struct {
	// gorm.Model provides ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model

	Username     string `gorm:"unique;not null"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	// Role can be used for authorization, e.g., 'user' or 'admin'.
	Role string `gorm:"default:'user';not null"`

	// This defines the one-to-many relationship in GORM.
	Monitors []Monitor
}

// Monitor represents a website to be checked.
type Monitor struct {
	// gorm.Model provides four default fields: ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model

	// IMPORTANT CHANGE: A URL is no longer globally unique. It should be unique PER USER.
	// We create a composite index to enforce this rule at the database level.
	URL string `gorm:"uniqueIndex:idx_user_url;not null"`

	// IntervalSec is how often this URL should be checked, in seconds.
	IntervalSec int `gorm:"not null"`

	// Active indicates whether this monitor is currently running.
	Active bool `gorm:"default:true"`

	// LastCheckedAt records the timestamp of the last health check.
	LastCheckedAt *time.Time

	// NextCheckAt records the timestamp when the next check is scheduled.
	NextCheckAt *time.Time

	// NEW: This is the foreign key that links a Monitor back to a User.
	// We also add the composite index here.
	UserID uint `gorm:"uniqueIndex:idx_user_url;not null"`
	// We don't need a User struct here because we don't usually load the user
	// when we load a monitor. The UserID is sufficient.
}

// CheckResult represents the outcome of a single health check for a Monitor.
// It corresponds to the 'check_results' table in the database.
type CheckResult struct {
	gorm.Model

	// MonitorID is the foreign key that links this result back to its Monitor.
	MonitorID uint `gorm:"not null;index"`

	// Cascading to ensure data integrity.
	Monitor Monitor `gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// StatusCode is the HTTP status code received (e.g., 200, 404).
	StatusCode int

	// ErrorMessage stores any network or other errors that occurred during the check.
	ErrorMessage string `gorm:"type:text"`

	// DurationMs is how long the check took, in milliseconds.
	DurationMs int64

	// CheckedAt is the timestamp when this check was performed.
	CheckedAt time.Time `gorm:"not null"`
}