

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

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use slog for fatal errors too.
		slog.Error("Fatal error loading configuration", "error", err)
		os.Exit(1) // Exit the program
	}

	// 2. Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		slog.Error("Fatal error initializing database", "error", err)
		os.Exit(1)
	}
	slog.Info("Database initialized successfully.")


	// 3. Create the scheduler
	sched := scheduler.NewScheduler(db, cfg)

	// 4. Create the API Server
	apiServer := api.NewServer(cfg, db, sched)

	// 5. Start the scheduler
	sched.Start()

	// 6. Start the API server
	go func() {
		if err := apiServer.Start(); err != nil {
			slog.Error("Fatal error starting API server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Application startup sequence complete. Services are running.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutdown signal received. Shutting down gracefully...")
	sched.Stop()
	slog.Info("Application has been shut down. Goodbye!")
}