package handlers

import (
	"html/template"
	"log"
	"net/http"
	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/models"
	"server-dashboard/internal/services"

	"github.com/gorilla/mux"
)

// ListServers handles the request to list all servers
func ListServers(c interface{}) {
	servers, err := services.GetAllServers()
	if err != nil {
		return
	}
	_ = servers
}

func ServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Servers</h1>"))
}

func ServerHandlerWithTemplates(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)

		// Get all servers from service
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

		vmCounts := make(map[string]int)
		for _, vm := range vms {
			if vm.HostServerID != "" {
				vmCounts[vm.HostServerID]++
			}
		}

		w.Header().Set("Content-Type", "text/html")

		data := map[string]interface{}{
			"servers":  servers,
			"vmCounts": vmCounts,
			"IsAdmin":  isAdminUser(cfg, username),
			"Username": username,
		}

		// Template is defined as "servers" in servers.html
		if err := templates.ExecuteTemplate(w, "servers", data); err != nil {
			log.Printf("Error rendering template: %v", err)
		}
	}
}

// ServerDetailHandlerWithTemplates handles the request to get a single server detail page
func ServerDetailHandlerWithTemplates(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)

		// Get server ID from URL parameters
		vars := mux.Vars(r)
		serverID := vars["id"]

		// Get all servers
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

		// Find the specific server
		var server *models.Server
		for i := range servers {
			if servers[i].ID == serverID {
				server = &servers[i]
				break
			}
		}

		if server == nil {
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html")

		// Attach VMs assigned to this server (if any)
		var attachedVMs []models.VM
		for _, vm := range vms {
			if vm.HostServerID == server.ID {
				attachedVMs = append(attachedVMs, vm)
			}
		}

		data := map[string]interface{}{
			"server":   server,
			"vms":      attachedVMs,
			"IsAdmin":  isAdminUser(cfg, username),
			"Username": username,
		}

		// Template is defined as "server-detail" in server-detail.html
		if err := templates.ExecuteTemplate(w, "server-detail", data); err != nil {
			log.Printf("Error rendering template: %v", err)
		}
	}
}
