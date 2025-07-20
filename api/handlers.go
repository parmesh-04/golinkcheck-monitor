package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"gorm.io/gorm"
)

//crud handlers for monitors


//return a list of all monitors
func (s *Server) handleListMonitors(w http.ResponseWriter, r *http.Request) {
	var monitors []database.Monitor
	if err := s.db.Find(&monitors).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch monitors from database")
		return
	}
	respondWithJSON(w, http.StatusOK, monitors)
}

//fetches a single monitor by ID
func (s *Server) handleGetMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Monitor ID is missing")
		return
	}

	//checks if the id is a valid integer 
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}

	//finds the monitor in the database and then return it or returns an error if not found
	var monitor database.Monitor
	if err := s.db.First(&monitor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Monitor not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}
	respondWithJSON(w, http.StatusOK, monitor)
}

//creates a new monitor
//expects a JSON payload with URL and IntervalSec
func (s *Server) handleCreateMonitor(w http.ResponseWriter, r *http.Request) {
	var req CreateMonitorRequest

	if err := parseAndValidate(r, &req); err != nil {
		slog.Error("Validation failed for create monitor request", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	//add the new monitor to the database
	newMonitor := database.Monitor{
		URL:         req.URL,
		IntervalSec: req.IntervalSec,
		Active:      true,
	}

	//attempts to create the monitor in the database
	if err := s.db.Create(&newMonitor).Error; err != nil {
		slog.Error("Failed to create monitor in db", "error", err)
		respondWithError(w, http.StatusConflict, "Could not create monitor (perhaps URL already exists?)")
		return
	}

	s.scheduler.AddMonitorJob(newMonitor)
	slog.Info("New monitor created via API", "monitor_id", newMonitor.ID, "url", newMonitor.URL)
	respondWithJSON(w, http.StatusCreated, newMonitor)
}

//updates an existing monitor
//expects a JSON payload with URL, IntervalSec, and Active status
func (s *Server) handleUpdateMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}

	var existingMonitor database.Monitor
	if err := s.db.First(&existingMonitor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Monitor not found to update")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database error while fetching monitor")
		}
		return
	}

	var req UpdateMonitorRequest
	if err := parseAndValidate(r, &req); err != nil {
		slog.Error("Validation failed for update monitor request", "monitor_id", id, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	existingMonitor.URL = req.URL
	existingMonitor.IntervalSec = req.IntervalSec
	existingMonitor.Active = req.Active

	if err := s.db.Save(&existingMonitor).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save updated monitor")
		return
	}
	//removes the existing job from the scheduler and re-adds it if active
	//this ensures the job is updated in the scheduler
	s.scheduler.RemoveMonitorJob(existingMonitor.ID)
	if existingMonitor.Active {
		s.scheduler.AddMonitorJob(existingMonitor)
		slog.Info("Updated and reactivated job", "monitor_id", existingMonitor.ID)
	} else {
		slog.Info("Deactivated job via update", "monitor_id", existingMonitor.ID)
	}

	respondWithJSON(w, http.StatusOK, existingMonitor)
}

//deletes a monitor by ID
//removes it from the database and the scheduler
func (s *Server) handleDeleteMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}

	s.scheduler.RemoveMonitorJob(uint(id))

	result := s.db.Unscoped().Delete(&database.Monitor{}, id)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete monitor from database")
		return
	}

	if result.RowsAffected == 0 {
		slog.Warn("Attempted to delete monitor, but it was not found", "monitor_id", id)
	} else {
		slog.Info("Deleted monitor", "monitor_id", id)
	}

	w.WriteHeader(http.StatusNoContent)
}
