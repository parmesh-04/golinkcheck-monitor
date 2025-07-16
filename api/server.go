

package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/scheduler"
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

	
	// Create a subrouter for all routes that need authentication.
	// We are saying "all routes starting with /monitors will use this subrouter".
	apiRouter := router.PathPrefix("/monitors").Subrouter()

	// Apply our authMiddleware to every single request that goes to this subrouter.
	apiRouter.Use(s.authMiddleware)

	// Now, attach your handlers to the SECURED apiRouter, not the main router.
	apiRouter.HandleFunc("", s.handleListMonitors).Methods("GET")
	apiRouter.HandleFunc("", s.handleCreateMonitor).Methods("POST")
	apiRouter.HandleFunc("/{id}", s.handleGetMonitor).Methods("GET")
	apiRouter.HandleFunc("/{id}", s.handleDeleteMonitor).Methods("DELETE")
	apiRouter.HandleFunc("/{id}", s.handleUpdateMonitor).Methods("PUT")

	slog.Info("API server listening", "address", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}