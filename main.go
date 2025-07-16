// main.go

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	

	// 3. Create and start the scheduler
	sched := scheduler.NewScheduler(db, cfg)
	sched.Start()

	log.Println("Application startup sequence complete. Monitoring is active.")
	log.Println("Press Ctrl+C to exit.")

	// 4. Implement graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received. Shutting down gracefully...")
	sched.Stop()
	log.Println("Application has been shut down. Goodbye!")
}