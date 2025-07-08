
package main

import (
"fmt"
"log"
"github.com/parmesh-04/golinkcheck-monitor/config"
)

func main() {
	fmt.Println("GoLinkCheck Monitor starting up...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err) // Use log.Fatalf to exit if config fails
	}

	
	fmt.Printf("Server will run on port: %s\n", cfg.ServerPort)
	fmt.Printf("Database will use URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("Default monitor interval: %d seconds\n", cfg.MonitorDefaultInterval)
	fmt.Printf("Scheduler concurrency limit: %d\n", cfg.SchedulerConcurrency)

	fmt.Println("Configuration loaded successfully!")
}