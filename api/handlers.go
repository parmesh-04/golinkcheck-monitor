// api/handlers.go

package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"gorm.io/gorm"
)

// handleListMonitors retrieves all monitors from the database.
// No changes were needed here as it doesn't process an input body.
func (s *Server) handleListMonitors(w http.ResponseWriter, r *http.Request) {
	var monitors []database.Monitor
	if err := s.db.Find(&monitors).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch monitors from database")
		return
	}
	respondWithJSON(w, http.StatusOK, monitors)
}

// handleGetMonitor retrieves a single monitor by its ID.
// No changes were needed here as it doesn't process an input body.
func (s *Server) handleGetMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Monitor ID is missing")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}
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

// handleCreateMonitor validates and creates a new monitor.
// This handler is now cleaner and uses the validation helper.
func (s *Server) handleCreateMonitor(w http.ResponseWriter, r *http.Request) {
	var req CreateMonitorRequest // Use our new API-specific request struct

	// Use the helper to decode the JSON body and run validation in one step.
	if err := parseAndValidate(r, &req); err != nil {
		slog.Error("Validation failed for create monitor request", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Map the validated request data to our database model.
	newMonitor := database.Monitor{
		URL:         req.URL,
		IntervalSec: req.IntervalSec,
		Active:      true, // New monitors are active by default.
	}

	if err := s.db.Create(&newMonitor).Error; err != nil {
		slog.Error("Failed to create monitor in db", "error", err)
		respondWithError(w, http.StatusConflict, "Could not create monitor (perhaps URL already exists?)")
		return
	}

	s.scheduler.AddMonitorJob(newMonitor)
	slog.Info("New monitor created via API", "monitor_id", newMonitor.ID, "url", newMonitor.URL)
	respondWithJSON(w, http.StatusCreated, newMonitor)
}

// handleUpdateMonitor validates and updates an existing monitor.
// This handler is now much cleaner and safer.
func (s *Server) handleUpdateMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}

	// First, fetch the monitor we want to update.
	var existingMonitor database.Monitor
	if err := s.db.First(&existingMonitor, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Monitor not found to update")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Database error while fetching monitor")
		}
		return
	}

	// Use the helper to decode and validate the incoming update data.
	var req UpdateMonitorRequest
	if err := parseAndValidate(r, &req); err != nil {
		slog.Error("Validation failed for update monitor request", "monitor_id", id, "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Apply the validated changes to the existing monitor model.
	existingMonitor.URL = req.URL
	existingMonitor.IntervalSec = req.IntervalSec
	existingMonitor.Active = req.Active

	if err := s.db.Save(&existingMonitor).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save updated monitor")
		return
	}

	// Resynchronize the scheduler with the new state.
	s.scheduler.RemoveMonitorJob(existingMonitor.ID)
	if existingMonitor.Active {
		s.scheduler.AddMonitorJob(existingMonitor)
		slog.Info("Updated and reactivated job", "monitor_id", existingMonitor.ID)
	} else {
		slog.Info("Deactivated job via update", "monitor_id", existingMonitor.ID)
	}

	respondWithJSON(w, http.StatusOK, existingMonitor)
}

// handleDeleteMonitor deletes a monitor by its ID.
// No changes were needed here as it doesn't process an input body.
func (s *Server) handleDeleteMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}

	// We must remove the job from the scheduler first.
	s.scheduler.RemoveMonitorJob(uint(id))

	// Use Unscoped() to perform a hard delete, even if using soft deletes elsewhere.
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