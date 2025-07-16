// api/server.go

package api

import (
	"log"
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
func (s *Server) Start() error {
	router := mux.NewRouter()

	// Define our API endpoints.
	router.HandleFunc("/monitors", s.handleListMonitors).Methods("GET")
	router.HandleFunc("/monitors", s.handleCreateMonitor).Methods("POST")
	router.HandleFunc("/monitors/{id}", s.handleGetMonitor).Methods("GET")
	router.HandleFunc("/monitors/{id}", s.handleDeleteMonitor).Methods("DELETE")
	router.HandleFunc("/monitors/{id}", s.handleUpdateMonitor).Methods("PUT")

	log.Println("API server listening on", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}