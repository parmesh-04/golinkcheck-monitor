package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/config"
	"github.com/parmesh-04/golinkcheck-monitor/scheduler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

type Server struct {
	listenAddr string
	db         *gorm.DB
	scheduler  *scheduler.Scheduler
	config     config.Config
}

// NewServer creates a new API server instance.
func NewServer(cfg config.Config, db *gorm.DB, sched *scheduler.Scheduler) *Server {
	return &Server{
		listenAddr: ":" + cfg.ServerPort,
		db:         db,
		scheduler:  sched,
		config:     cfg,
	}
}

// Start initializes the API server and starts listening for requests.
func (s *Server) Start() error {
	router := mux.NewRouter()

	// Register Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	apiRouter := router.PathPrefix("/monitors").Subrouter()

	apiRouter.Use(s.authMiddleware)

	apiRouter.HandleFunc("", s.handleListMonitors).Methods("GET")
	apiRouter.HandleFunc("", s.handleCreateMonitor).Methods("POST")
	apiRouter.HandleFunc("/{id}", s.handleGetMonitor).Methods("GET")
	apiRouter.HandleFunc("/{id}", s.handleDeleteMonitor).Methods("DELETE")
	apiRouter.HandleFunc("/{id}", s.handleUpdateMonitor).Methods("PUT")

	slog.Info("API server listening", "address", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}
