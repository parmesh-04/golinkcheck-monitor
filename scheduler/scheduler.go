// scheduler/scheduler.go

package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/parmesh-04/golinkcheck-monitor/checker"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler manages all the scheduled monitoring jobs.
type Scheduler struct {
	cronRunner *cron.Cron
	db         *gorm.DB
	config     config.Config
	// A map to keep track of running jobs. Key is Monitor ID, value is Cron Entry ID.
	activeJobs map[uint]cron.EntryID
}

// NewScheduler creates and configures a new Scheduler.
func NewScheduler(db *gorm.DB, cfg config.Config) *Scheduler {
	//  seconds is the smallest unit of time.
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cronRunner: c,
		db:         db,
		config:     cfg,
		activeJobs: make(map[uint]cron.EntryID),
	}
}

// Start loads all active monitors from the database and schedules them.

func (s *Scheduler) Start() {
	slog.Info("Scheduler starting...")

	// 1. Load active monitors from the database.
	var monitors []database.Monitor
	s.db.Where("active = ?", true).Find(&monitors)
	

	// 2. Schedule a job for each monitor.
	// AddMonitorJob defined below
	for _, monitor := range monitors {
		s.AddMonitorJob(monitor)
	}

	// 3. Start the cron runner in the background.
	s.cronRunner.Start()
	slog.Info("Scheduler started", "active_jobs", len(s.activeJobs))
}

// Stop shuts down the cron runner.
func (s *Scheduler) Stop() {
	slog.Info("Scheduler stopping...")
	// The Stop method waits for any running jobs to complete.
	ctx := s.cronRunner.Stop()
	<-ctx.Done() // Wait until the stop is complete.
	slog.Info("Scheduler stopped")
}

// AddMonitorJob adds a new monitoring job to the scheduler.
// This is the core function that defines what happens on each check.
func (s *Scheduler) AddMonitorJob(monitor database.Monitor) {
	// This is important! We create a local copy of the monitor variable.
	// This ensures that each scheduled job gets its own, correct version of the monitor.
	m := monitor

	// Create a cron schedule string, e.g., "@every 60s"
	schedule := fmt.Sprintf("@every %ds", m.IntervalSec)

	// Add the job to the cron runner.
	// AddFunc takes a schedule string and a function to run.
	entryID, err := s.cronRunner.AddFunc(schedule, func() {
		slog.Info("-> Running check", "monitor_id", m.ID, "url", m.URL)

		// 1. Perform the check using the checker package we built.
		timeout := time.Duration(s.config.MonitorCheckTimeoutSec) * time.Second
		checkResult := checker.Check(m.URL, timeout)

		// 2. Link the result back to its monitor.
		checkResult.MonitorID = m.ID

		// 3. Save the result to the database.
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

	// Store the job's ID so we can manage it later (e.g., stop or remove it).
	s.activeJobs[m.ID] = entryID
	slog.Info(
		"Scheduled new monitor",
		"monitor_id", m.ID,
		"url", m.URL,
		"interval_sec", m.IntervalSec,
		"job_id", entryID,
	)
}


func (s *Scheduler) RemoveMonitorJob(monitorID uint) {
	// 1. Look up the job's internal cron ID from our map.
	entryID, found := s.activeJobs[monitorID]
	if !found {
		// If it's not in our map, it's not a running job. Nothing to do.
		slog.Warn("Could not find job to remove", "monitor_id", monitorID)
		return
	}

	// 2. Tell the cron runner to remove the job with that ID.
	s.cronRunner.Remove(entryID)

	// 3. IMPORTANT: Remove the entry from our tracking map to keep our state consistent.
	delete(s.activeJobs, monitorID)

	slog.Info("Removed job from scheduler", "monitor_id", monitorID, "job_id", entryID)
}