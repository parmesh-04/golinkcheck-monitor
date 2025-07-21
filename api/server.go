// File: api/server.go

package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/handlers"
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

	// Public Prometheus metrics endpoint - no auth needed.
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	//  Public Authentication Routes ---
	// These endpoints are for registering and logging in, so they must NOT have auth middleware.
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", s.handleRegister).Methods("POST")
	authRouter.HandleFunc("/login", s.handleLogin).Methods("POST")

	//  Protected Monitor Routes ---
	// All routes under /monitors now require a valid JWT.
	apiRouter := router.PathPrefix("/monitors").Subrouter()
	// Use the new jwtAuthMiddleware to protect these endpoints.
	apiRouter.Use(s.jwtAuthMiddleware)

	// These handlers will be updated next to be user-aware.
	apiRouter.HandleFunc("", s.handleListMonitors).Methods("GET")
	apiRouter.HandleFunc("", s.handleCreateMonitor).Methods("POST")
	apiRouter.HandleFunc("/{id}", s.handleGetMonitor).Methods("GET")
	apiRouter.HandleFunc("/{id}", s.handleDeleteMonitor).Methods("DELETE")
	apiRouter.HandleFunc("/{id}", s.handleUpdateMonitor).Methods("PUT")

	slog.Info("API server listening", "address", s.listenAddr)

	// CORS Configuration
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	corsRouter := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router)

	// Start the server with the CORS-wrapped router
	return http.ListenAndServe(s.listenAddr, corsRouter)
}