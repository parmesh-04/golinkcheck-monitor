// scheduler/scheduler.go

package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/parmesh-04/golinkcheck-monitor/checker"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"github.com/parmesh-04/golinkcheck-monitor/metrics" // Import our new metrics package
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler manages all the scheduled monitoring jobs.
type Scheduler struct {
	cronRunner *cron.Cron
	db         *gorm.DB
	config     config.Config
	activeJobs map[uint]cron.EntryID
}

// NewScheduler creates and configures a new Scheduler.
func NewScheduler(db *gorm.DB, cfg config.Config) *Scheduler {
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cronRunner: c,
		db:         db,
		config:     cfg,
		activeJobs: make(map[uint]cron.EntryID),
	}
}

// Start loads all active monitors from the database, schedules them, and updates metrics.
func (s *Scheduler) Start() {
	slog.Info("Scheduler starting...")

	var monitors []database.Monitor
	s.db.Where("active = ?", true).Find(&monitors)

	for _, monitor := range monitors {
		s.AddMonitorJob(monitor)
	}

	s.cronRunner.Start()

	// Set the initial value for our active jobs gauge.
	metrics.ActiveJobs.Set(float64(len(s.activeJobs)))

	slog.Info("Scheduler started", "active_jobs", len(s.activeJobs))
}

// Stop gracefully shuts down the cron runner.
func (s *Scheduler) Stop() {
	slog.Info("Scheduler stopping...")
	ctx := s.cronRunner.Stop()
	<-ctx.Done()
	slog.Info("Scheduler stopped")
}

// AddMonitorJob adds a new monitoring job and instruments it with metrics.
func (s *Scheduler) AddMonitorJob(monitor database.Monitor) {
	m := monitor
	schedule := fmt.Sprintf("@every %ds", m.IntervalSec)

	entryID, err := s.cronRunner.AddFunc(schedule, func() {
		slog.Info("-> Running check", "monitor_id", m.ID, "url", m.URL)
		checkStartTime := time.Now() // Start timer for metric

		timeout := time.Duration(s.config.MonitorCheckTimeoutSec) * time.Second
		checkResult := checker.Check(m.URL, timeout)

		// --- METRICS INSTRUMENTATION ---
		// Observe the duration in our histogram.
		durationInSeconds := time.Since(checkStartTime).Seconds()
		metrics.CheckDuration.Observe(durationInSeconds)

		// Increment the total checks counter with the appropriate status label.
		if checkResult.ErrorMessage != "" {
			metrics.ChecksTotal.WithLabelValues("failure").Inc()
		} else {
			metrics.ChecksTotal.WithLabelValues("success").Inc()
		}
		// --- END METRICS ---

		checkResult.MonitorID = m.ID
		if dbErr := s.db.Create(&checkResult).Error; dbErr != nil {
			slog.Error("Error saving check result", "monitor_id", m.ID, "error", dbErr)
		} else {
			slog.Info(
				"<- Check successful",
				"monitor_id", m.ID,
				"status_code", checkResult.StatusCode,
				"duration_ms", checkResult.DurationMs,
			)
		}
	})

	if err != nil {
		slog.Error("Error adding monitor to scheduler", "monitor_id", m.ID, "error", err)
		return
	}

	s.activeJobs[m.ID] = entryID
	// Increment the active jobs gauge since we've added one.
	metrics.ActiveJobs.Inc()

	slog.Info(
		"Scheduled new monitor",
		"monitor_id", m.ID,
		"url", m.URL,
		"interval_sec", m.IntervalSec,
		"job_id", entryID,
	)
}

// RemoveMonitorJob removes a job from the scheduler and updates metrics.
func (s *Scheduler) RemoveMonitorJob(monitorID uint) {
	entryID, found := s.activeJobs[monitorID]
	if !found {
		slog.Warn("Could not find job to remove", "monitor_id", monitorID)
		return
	}

	s.cronRunner.Remove(entryID)
	delete(s.activeJobs, monitorID)

	// Decrement the active jobs gauge since we've removed one.
	metrics.ActiveJobs.Dec()

	slog.Info("Removed job from scheduler", "monitor_id", monitorID, "job_id", entryID)
}