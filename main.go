package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/parmesh-04/golinkcheck-monitor/api"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"github.com/parmesh-04/golinkcheck-monitor/logging"
	"github.com/parmesh-04/golinkcheck-monitor/scheduler"
)

func main() {
	// Initialize logger as the very first step.
	logging.InitLogger()

	slog.Info("GoLinkCheck Monitor starting up...")

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Fatal error loading configuration", "error", err)
		os.Exit(1)
	}

	// 2. Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		slog.Error("Fatal error initializing database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database initialized successfully.")

	// --- THIS IS THE NEWLY ADDED LINE ---
	// Seed the database with initial data if it's empty.
	database.Seed(db)
	// --- END OF NEWLY ADDED LINE ---

	// 3. Create the scheduler
	sched := scheduler.NewScheduler(db, cfg)

	// 4. Create the API Server
	apiServer := api.NewServer(cfg, db, sched)

	// 5. Start the scheduler in the background
	sched.Start()

	// 6. Start the API server in a separate goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			slog.Error("Fatal error starting API server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Application startup sequence complete. Services are running.")

	// 7. Wait for a shutdown signal (e.g., Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 8. Perform graceful shutdown
	slog.Info("Shutdown signal received. Shutting down gracefully...")
	sched.Stop()
	slog.Info("Application has been shut down. Goodbye!")
}