// api/server.go

package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/scheduler"
	"github.com/prometheus/client_golang/prometheus/promhttp" // Import the Prometheus HTTP handler
	"gorm.io/gorm"
)

// Server holds all the dependencies our API needs to function.
type Server struct {
	listenAddr string
	db         *gorm.DB
	scheduler  *scheduler.Scheduler
	config     config.Config
}

// NewServer creates and configures a new API server instance.
func NewServer(cfg config.Config, db *gorm.DB, sched *scheduler.Scheduler) *Server {
	return &Server{
		listenAddr: ":" + cfg.ServerPort,
		db:         db,
		scheduler:  sched,
		config:     cfg,
	}
}

// Start creates the routes and starts listening for HTTP requests.
func (s *Server) Start() error {
	router := mux.NewRouter()

	// --- THIS IS THE NEW LINE ---
	// Expose the /metrics endpoint for Prometheus scraping.
	// This is attached to the main router because it should NOT be authenticated.
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	// --- END OF NEW LINE ---

	// Create a subrouter for all API routes that need authentication.
	apiRouter := router.PathPrefix("/monitors").Subrouter()

	// Apply our authMiddleware to every single request that goes to this subrouter.
	apiRouter.Use(s.authMiddleware)

	// Attach your handlers to the SECURED apiRouter.
	apiRouter.HandleFunc("", s.handleListMonitors).Methods("GET")
	apiRouter.HandleFunc("", s.handleCreateMonitor).Methods("POST")
	apiRouter.HandleFunc("/{id}", s.handleGetMonitor).Methods("GET")
	apiRouter.HandleFunc("/{id}", s.handleDeleteMonitor).Methods("DELETE")
	apiRouter.HandleFunc("/{id}", s.handleUpdateMonitor).Methods("PUT")

	slog.Info("API server listening", "address", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}