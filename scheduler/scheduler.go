// scheduler/scheduler.go

package scheduler

import (
	"fmt"
	"log"
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
	log.Println("Scheduler starting...")

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
	log.Printf("Scheduler started with %d active jobs.", len(s.activeJobs))
}

// Stop shuts down the cron runner.
func (s *Scheduler) Stop() {
	log.Println("Scheduler stopping...")
	// The Stop method waits for any running jobs to complete.
	ctx := s.cronRunner.Stop()
	<-ctx.Done() // Wait until the stop is complete.
	log.Println("Scheduler stopped.")
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
		log.Printf("-> Running check for monitor #%d: %s", m.ID, m.URL)

		// 1. Perform the check using the checker package we built.
		timeout := time.Duration(s.config.MonitorCheckTimeoutSec) * time.Second
		checkResult := checker.Check(m.URL, timeout)

		// 2. Link the result back to its monitor.
		checkResult.MonitorID = m.ID

		// 3. Save the result to the database.
		if dbErr := s.db.Create(&checkResult).Error; dbErr != nil {
			log.Printf("Error saving check result for monitor #%d: %v", m.ID, dbErr)
		} else {
			log.Printf("<- Check for monitor #%d successful. Status: %d, Time: %dms", m.ID, checkResult.StatusCode, checkResult.DurationMs)
		}
	})

	if err != nil {
		log.Printf("Error adding monitor #%d to scheduler: %v", m.ID, err)
		return
	}

	// Store the job's ID so we can manage it later (e.g., stop or remove it).
	s.activeJobs[m.ID] = entryID
	log.Printf("Scheduled monitor #%d (%s) to run every %d seconds. Job ID: %d", m.ID, m.URL, m.IntervalSec, entryID)
}