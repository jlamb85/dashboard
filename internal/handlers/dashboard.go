package handlers

import (
    "net/http"
    "html/template"
    "log"
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

func DashboardHandlerWithTemplates(w http.ResponseWriter, r *http.Request, templates *template.Template) {
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
    
    w.Header().Set("Content-Type", "text/html")
    
    data := map[string]interface{}{
        "servers": servers,
        "vms": vms,
    }
    
    // Template is defined with name "dashboard" in dashboard.html
    if err := templates.ExecuteTemplate(w, "dashboard", data); err != nil {
        log.Printf("Error rendering template: %v", err)
    }
}