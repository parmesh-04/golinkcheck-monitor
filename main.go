// File: main.go

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
	// Initialize the structured logger first.
	logging.InitLogger()

	slog.Info("GoLinkCheck Monitor starting up...")

	// Load configuration from environment variables/file.
	// The convention is to pass the path where the config file (e.g., app.env) is located.
	// "." means the current directory.
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Fatal error loading configuration", "error", err)
		os.Exit(1)
	}

	// Initialize the database connection, passing the loaded config.
	db, err := database.InitDB(cfg)
	if err != nil {
		slog.Error("Fatal error initializing database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database initialized successfully.")

	// Optional: Seed the database with initial data (e.g., an admin user).
	// It's good practice to have this function check if it needs to run.
	database.Seed(db)

	// Create the scheduler instance.
	sched := scheduler.NewScheduler(db, cfg)

	// Create the API server instance.
	apiServer := api.NewServer(cfg, db, sched)

	// Start the scheduler in a background goroutine.
	sched.Start()

	// Start the API server in a background goroutine so it doesn't block.
	go func() {
		if err := apiServer.Start(); err != nil {
			slog.Error("Fatal error starting API server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Application startup sequence complete. Services are running.")

	// Wait for a shutdown signal (e.g., Ctrl+C).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received.

	slog.Info("Shutdown signal received. Shutting down gracefully...")

	// Stop the scheduler, allowing running jobs to finish.
	sched.Stop()

	slog.Info("Application has been shut down. Goodbye!")
}