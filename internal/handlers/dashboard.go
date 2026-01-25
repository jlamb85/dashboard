package handlers

import (
	"html/template"
	"log"
	"net/http"
	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/services"
)

type DashboardHandlerType struct {
	Templates *template.Template
}

func NewDashboardHandler(templates *template.Template) *DashboardHandlerType {
	return &DashboardHandlerType{Templates: templates}
}

func (h *DashboardHandlerType) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use template name defined in dashboard.html
	if err := h.Templates.ExecuteTemplate(w, "dashboard", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Dashboard</h1>"))
}

func DashboardHandlerWithTemplates(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)
		// Get servers and VMs from service
		servers, err := services.GetAllServers()
		if err != nil {
			http.Error(w, "Error fetching servers: "+err.Error(), http.StatusInternalServerError)
			return
		}

		vms, err := services.GetAllVMs()
		if err != nil {
			http.Error(w, "Error fetching VMs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switches, err := services.GetAllSwitches()
		if err != nil {
			http.Error(w, "Error fetching switches: "+err.Error(), http.StatusInternalServerError)
			return
		}

		synthetics := services.GetSyntheticResults()
		okCount := 0
		worstLatency := int64(0)
		for _, s := range synthetics {
			if s.Status == "ok" {
				okCount++
			}
			if s.LatencyMs > worstLatency {
				worstLatency = s.LatencyMs
			}
		}

		w.Header().Set("Content-Type", "text/html")

		data := map[string]interface{}{
			"servers":                servers,
			"vms":                    vms,
			"switches":               switches,
			"synthetics":             synthetics,
			"syntheticsOK":           okCount,
			"syntheticsTotal":        len(synthetics),
			"syntheticsWorstLatency": worstLatency,
			"IsAdmin":                isAdminUser(cfg, username),
			"Username":               username,
		}

		// Template is defined with name "dashboard" in dashboard.html
		if err := templates.ExecuteTemplate(w, "dashboard", data); err != nil {
			log.Printf("Error rendering template: %v", err)
		}
	}
}
