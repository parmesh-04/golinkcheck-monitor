
package database

import (
	"time"

	"gorm.io/gorm"
)


type Monitor struct {
	// gorm.Model provides four default fields: ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model

	URL string `gorm:"uniqueIndex;not null"` // Each URL must be unique and not empty

	// IntervalSec is how often this URL should be checked, in seconds.
	IntervalSec int `gorm:"not null"`

	// Active indicates whether this monitor is currently running.
	Active bool `gorm:"default:true"`

	// LastCheckedAt records the timestamp of the last health check.
	// It's a pointer so it can be nil if never checked.
	LastCheckedAt *time.Time

	// NextCheckAt records the timestamp when the next check is scheduled.
	NextCheckAt *time.Time
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
	DurationMs int

	// CheckedAt is the timestamp when this check was performed.
	CheckedAt time.Time `gorm:"not null"`
}