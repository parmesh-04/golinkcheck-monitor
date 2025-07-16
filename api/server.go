// api/server.go

package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"github.com/parmesh-04/golinkcheck-monitor/scheduler"
	"gorm.io/gorm"
)

// Server holds all the dependencies our API needs to function.
type Server struct {
	listenAddr string
	db         *gorm.DB
	scheduler  *scheduler.Scheduler
}

// NewServer creates and configures a new API server instance.
func NewServer(cfg config.Config, db *gorm.DB, sched *scheduler.Scheduler) *Server {
	return &Server{
		listenAddr: ":" + cfg.ServerPort,
		db:         db,
		scheduler:  sched,
	}
}

// Start creates the routes and starts listening for HTTP requests.
// This is a blocking call.
func (s *Server) Start() error {
	router := mux.NewRouter()

	// Define our API endpoints.
	router.HandleFunc("/monitors", s.handleListMonitors).Methods("GET")
	router.HandleFunc("/monitors", s.handleCreateMonitor).Methods("POST")

	log.Println("API server listening on", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}

// --- Handler Functions ---

// handleListMonitors fetches all monitors from the database and returns them as JSON.
func (s *Server) handleListMonitors(w http.ResponseWriter, r *http.Request) {
	var monitors []database.Monitor
	if err := s.db.Find(&monitors).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch monitors from database")
		return
	}
	respondWithJSON(w, http.StatusOK, monitors)
}

// handleCreateMonitor reads a new monitor from the request body, saves it,
// and adds it to the running scheduler.
func (s *Server) handleCreateMonitor(w http.ResponseWriter, r *http.Request) {
	var newMonitor database.Monitor

	// Try to decode the JSON from the request's body into our struct.
	if err := json.NewDecoder(r.Body).Decode(&newMonitor); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Simple validation to ensure we have the minimum required data.
	if newMonitor.URL == "" || newMonitor.IntervalSec <= 0 {
		respondWithError(w, http.StatusBadRequest, "URL and a positive IntervalSec are required")
		return
	}

	// Save the new monitor to the database.
	if err := s.db.Create(&newMonitor).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create monitor in database")
		return
	}

	// IMPORTANT: Add the newly created monitor to the live scheduler.
	s.scheduler.AddMonitorJob(newMonitor)
	log.Printf("New monitor for URL [%s] created via API and added to scheduler.", newMonitor.URL)

	respondWithJSON(w, http.StatusCreated, newMonitor)
}

// --- Helper Functions ---

// respondWithError is a helper to send a standardized JSON error message.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON is a helper to marshal data to JSON and write the response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}