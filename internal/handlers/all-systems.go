package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/services"
)

// AllSystemsHandlerWithTemplates displays a comprehensive view of all systems
func AllSystemsHandlerWithTemplates(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)

		// Get all devices from service
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

		w.Header().Set("Content-Type", "text/html")

		data := map[string]interface{}{
			"servers":  servers,
			"vms":      vms,
			"switches": switches,
			"IsAdmin":  isAdminUser(cfg, username),
			"Username": username,
		}

		if err := templates.ExecuteTemplate(w, "all-systems.html", data); err != nil {
			log.Printf("Error rendering all-systems template: %v", err)
			fmt.Fprintf(w, "Template Error: %v", err)
		}
	}
}
