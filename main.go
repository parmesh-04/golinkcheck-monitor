// main.go

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/parmesh-04/golinkcheck-monitor/api"       // Import the API package
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"github.com/parmesh-04/golinkcheck-monitor/scheduler"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("GoLinkCheck Monitor starting up...")

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fatal error loading configuration: %v", err)
	}

	// 2. Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Fatal error initializing database: %v", err)
	}
	log.Println("Database initialized successfully.")

	// 3. Create the scheduler
	sched := scheduler.NewScheduler(db, cfg)

	// 4. Create the API Server, giving it the db and scheduler it needs
	apiServer := api.NewServer(cfg, db, sched)

	// 5. Start the scheduler in the background
	sched.Start()

	// 6. Start the API server in a separate, non-blocking goroutine
	go func() {
		log.Println("Starting API server...")
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Fatal error: API server failed to start: %v", err)
		}
	}()

	log.Println("Application startup sequence complete. Services are running.")
	log.Println("Press Ctrl+C to exit.")

	// 7. Wait for a shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received. Shutting down gracefully...")

	// Stop the scheduler, allowing running jobs to complete.
	sched.Stop()

	log.Println("Application has been shut down. Goodbye!")
}