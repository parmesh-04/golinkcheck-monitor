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
	logging.InitLogger()

	slog.Info("GoLinkCheck Monitor starting up...")
	//reads the env variables and loads the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Fatal error loading configuration", "error", err)
		os.Exit(1)
	}
	//initializes the database connection
	db, err := database.InitDB(cfg)
	if err != nil {
		slog.Error("Fatal error initializing database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database initialized successfully.")

	database.Seed(db)

	sched := scheduler.NewScheduler(db, cfg)

	apiServer := api.NewServer(cfg, db, sched)
	// Start the scheduler in a separate goroutine
	sched.Start()
	// Start the API server in a separate goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			slog.Error("Fatal error starting API server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Application startup sequence complete. Services are running.")
		//gracefully handle shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutdown signal received. Shutting down gracefully...")
	sched.Stop()
	slog.Info("Application has been shut down. Goodbye!")
}
