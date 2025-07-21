
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/parmesh-04/golinkcheck-monitor/database"
	"gorm.io/gorm"
)

//  A helper function to safely get user claims from the request context.
func getClaimsFromContext(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(userContextKey).(*Claims)
	return claims, ok
}

// handleListMonitors returns a list of monitors.
// Admins see all monitors; regular users only see their own.
func (s *Server) handleListMonitors(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user claims from token")
		return
	}

	var monitors []database.Monitor
	
	//  --- Authorization Logic ---
	// Start with a base query.
	query := s.db
	// If the user is NOT an admin, add a condition to only fetch their monitors.
	if claims.Role != "admin" {
		query = query.Where("user_id = ?", claims.UserID)
	}

	// Execute the final query.
	if err := query.Order("created_at desc").Find(&monitors).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch monitors from database")
		return
	}
	respondWithJSON(w, http.StatusOK, monitors)
}

// handleGetMonitor fetches a single monitor by ID, checking for ownership.
func (s *Server) handleGetMonitor(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user claims from token")
		return
	}

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

	//  --- Authorization Check ---
	// Check if the monitor belongs to the user, or if the user is an admin.
	if monitor.UserID != claims.UserID && claims.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "You are not authorized to access this monitor")
		return
	}

	respondWithJSON(w, http.StatusOK, monitor)
}

// handleCreateMonitor creates a new monitor and assigns it to the current user.
func (s *Server) handleCreateMonitor(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user claims from token")
		return
	}

	// Using a simple struct here as your request/validation logic may vary.
	var req struct {
		URL         string `json:"url"`
		IntervalSec int    `json:"intervalSec"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to parse create monitor request", "error", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	
	newMonitor := database.Monitor{
		URL:         req.URL,
		IntervalSec: req.IntervalSec,
		Active:      true,
		UserID:      claims.UserID, 
	}

	if err := s.db.Create(&newMonitor).Error; err != nil {
		slog.Error("Failed to create monitor in db", "error", err)
		respondWithError(w, http.StatusConflict, "Could not create monitor (URL may already exist for this user)")
		return
	}

	s.scheduler.AddMonitorJob(newMonitor)
	slog.Info("New monitor created via API", "monitor_id", newMonitor.ID, "user_id", newMonitor.UserID, "url", newMonitor.URL)
	respondWithJSON(w, http.StatusCreated, newMonitor)
}

// handleUpdateMonitor updates an existing monitor after checking for ownership.
func (s *Server) handleUpdateMonitor(w http.ResponseWriter, r *http.Request) {
	// NEW: Get user claims for the auth check.
	claims, ok := getClaimsFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user claims from token")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
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

	//  --- Authorization Check ---
	if existingMonitor.UserID != claims.UserID && claims.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "You are not authorized to update this monitor")
		return
	}

	var req struct { 
		URL         string `json:"url"`
		IntervalSec int    `json:"intervalSec"`
		Active      bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	
	s.scheduler.RemoveMonitorJob(existingMonitor.ID)
	if existingMonitor.Active {
		s.scheduler.AddMonitorJob(existingMonitor)
		slog.Info("Updated and reactivated job", "monitor_id", existingMonitor.ID)
	} else {
		slog.Info("Deactivated job via update", "monitor_id", existingMonitor.ID)
	}

	respondWithJSON(w, http.StatusOK, existingMonitor)
}

// handleDeleteMonitor deletes a monitor by ID after checking for ownership.
func (s *Server) handleDeleteMonitor(w http.ResponseWriter, r *http.Request) {
	// NEW: Get user claims for the auth check.
	claims, ok := getClaimsFromContext(r)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve user claims from token")
		return
	}
	
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Monitor ID")
		return
	}

	// Fetch the monitor first to check for ownership before deleting.
	var monitorToDelete database.Monitor
	if err := s.db.First(&monitorToDelete, id).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Monitor not found to delete")
		return
	}

	// --- Authorization Check ---
	if monitorToDelete.UserID != claims.UserID && claims.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "You are not authorized to delete this monitor")
		return
	}

	// Now that user is authorized, proceed with deletion.
	s.scheduler.RemoveMonitorJob(uint(id))

	result := s.db.Delete(&database.Monitor{}, id)
	if result.Error != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete monitor from database")
		return
	}

	slog.Info("Deleted monitor", "monitor_id", id, "user_id", claims.UserID)
	w.WriteHeader(http.StatusNoContent)
}