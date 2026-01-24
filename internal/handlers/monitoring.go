package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"server-dashboard/internal/models"
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

// MonitoringPageData represents data for the monitoring page template
type MonitoringPageData struct {
	Servers                   []models.Server
	VMs                       []models.VM
	Switches                  []models.Switch
	UIEnableAutoRefresh       bool
	UIAutoRefreshSeconds      int
	UIShowSynthetics          bool
	UIShowMonitoringFeatures  bool
	UIShowNavigationButtons   bool
	UIShowQuickSummary        bool
	CurrentYear               int
	AppVersion                string
}

// MonitoringPageHandlerWithTemplates returns a handler for the monitoring page
func MonitoringPageHandlerWithTemplates(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all servers, VMs, and switches
		servers, _ := services.GetAllServers()
		vms, _ := services.GetAllVMs()
		switches, _ := services.GetAllSwitches()

		// Prepare data for template
		data := MonitoringPageData{
			Servers:                  servers,
			VMs:                      vms,
			Switches:                 switches,
			UIEnableAutoRefresh:      true,
			UIAutoRefreshSeconds:     30,
			UIShowSynthetics:         true,
			UIShowMonitoringFeatures: true,
			UIShowNavigationButtons:  true,
			UIShowQuickSummary:       true,
			CurrentYear:              2026,
			AppVersion:               "v1.0.1-dev",
		}

		if err := templates.ExecuteTemplate(w, "monitoring.html", data); err != nil {
			http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
