package handlers

import (
	"encoding/json"
	"net/http"

	"server-dashboard/internal/services"
)

// MonitoringStatusResponse represents the monitoring status
type MonitoringStatusResponse struct {
	Active bool   `json:"active"`
	Status string `json:"status"`
}

// MonitoringActionResponse represents the response to a monitoring action
type MonitoringActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Active  bool   `json:"active"`
}

// GetMonitoringStatus returns the current monitoring status
func GetMonitoringStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	active := services.GetMonitoringStatus()
	status := "stopped"
	if active {
		status = "running"
	}

	response := MonitoringStatusResponse{
		Active: active,
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartMonitoring starts the monitoring service
func StartMonitoring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := services.StartMonitoring()
	response := MonitoringActionResponse{
		Success: err == nil,
		Active:  services.GetMonitoringStatus(),
	}

	if err != nil {
		response.Message = "Failed to start monitoring: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Message = "Monitoring started successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopMonitoring stops the monitoring service
func StopMonitoring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := services.StopMonitoring()
	response := MonitoringActionResponse{
		Success: err == nil,
		Active:  services.GetMonitoringStatus(),
	}

	if err != nil {
		response.Message = "Failed to stop monitoring: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Message = "Monitoring stopped successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RestartMonitoring restarts the monitoring service
func RestartMonitoring(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := services.RestartMonitoring()
	response := MonitoringActionResponse{
		Success: err == nil,
		Active:  services.GetMonitoringStatus(),
	}

	if err != nil {
		response.Message = "Failed to restart monitoring: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Message = "Monitoring restarted successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
