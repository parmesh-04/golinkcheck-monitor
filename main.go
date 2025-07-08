
package main

import (
	"fmt"
	"log"

	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/database" // Import our database package
)

func main() {
	fmt.Println("GoLinkCheck Monitor starting up...")

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 2. Initialize database
	_, err = database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	
	log.Println("Database initialized successfully.")

	
	fmt.Printf("Server will run on port: %s\n", cfg.ServerPort)
	fmt.Printf("Database will use URL: %s\n", cfg.DatabaseURL)

	fmt.Println("Application startup sequence complete!")
}