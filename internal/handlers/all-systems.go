package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"server-dashboard/internal/services"
)

// AllSystemsHandlerWithTemplates displays a comprehensive view of all systems
func AllSystemsHandlerWithTemplates(w http.ResponseWriter, r *http.Request, templates *template.Template) {
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
	}
	
	if err := templates.ExecuteTemplate(w, "all-systems.html", data); err != nil {
		log.Printf("Error rendering all-systems template: %v", err)
		fmt.Fprintf(w, "Template Error: %v", err)
	}
}
