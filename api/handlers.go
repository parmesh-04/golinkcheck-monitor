
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

func (s *Server) handleListMonitors(w http.ResponseWriter, r *http.Request) {
	var monitors []database.Monitor
	if err := s.db.Find(&monitors).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch monitors from database")
		return
	}
	respondWithJSON(w, http.StatusOK, monitors)
}

func (s *Server) handleCreateMonitor(w http.ResponseWriter, r *http.Request) {
	var newMonitor database.Monitor
	if err := json.NewDecoder(r.Body).Decode(&newMonitor); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if newMonitor.URL == "" || newMonitor.IntervalSec <= 0 {
		respondWithError(w, http.StatusBadRequest, "URL and a positive IntervalSec are required")
		return
	}
	if err := s.db.Create(&newMonitor).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create monitor in database")
		return
	}
	s.scheduler.AddMonitorJob(newMonitor)
	slog.Info("New monitor created via API", "monitor_id", newMonitor.ID, "url", newMonitor.URL)
	respondWithJSON(w, http.StatusCreated, newMonitor)
}

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
	var updateData database.Monitor
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	existingMonitor.URL = updateData.URL
	existingMonitor.IntervalSec = updateData.IntervalSec
	existingMonitor.Active = updateData.Active
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