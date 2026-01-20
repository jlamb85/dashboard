package handlers

import (
    "net/http"
    "html/template"
    "log"
    "server-dashboard/internal/services"
    "server-dashboard/internal/models"
    "github.com/gorilla/mux"
)

// ListVMs handles the request to list all virtual machines
func ListVMs(w http.ResponseWriter, r *http.Request) {
    vms, err := services.GetAllVMs()
    if err != nil {
        http.Error(w, "Unable to retrieve VMs", http.StatusInternalServerError)
        return
    }
    _ = vms
}

func VMHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte("<h1>Virtual Machines</h1>"))
}

func VMHandlerWithTemplates(w http.ResponseWriter, r *http.Request, templates *template.Template) {
    // Get all VMs from service
    vms, err := services.GetAllVMs()
    if err != nil {
        http.Error(w, "Error fetching VMs: "+err.Error(), http.StatusInternalServerError)
        return
    }

    servers, err := services.GetAllServers()
    if err != nil {
        http.Error(w, "Error fetching servers: "+err.Error(), http.StatusInternalServerError)
        return
    }

    serverNames := make(map[string]string)
    for _, srv := range servers {
        serverNames[srv.ID] = srv.Name
    }
    
    w.Header().Set("Content-Type", "text/html")
    
    data := map[string]interface{}{
        "vms": vms,
        "serverNames": serverNames,
    }
    
    // Template is defined as "vms" in vms.html
    if err := templates.ExecuteTemplate(w, "vms", data); err != nil {
        log.Printf("Error rendering template: %v", err)
    }
}

// VMDetailHandlerWithTemplates handles the request to get a single VM detail page
func VMDetailHandlerWithTemplates(w http.ResponseWriter, r *http.Request, templates *template.Template) {
    // Get VM ID from URL parameters
    vars := mux.Vars(r)
    vmID := vars["id"]
    
    // Get all VMs
    vms, err := services.GetAllVMs()
    if err != nil {
        http.Error(w, "Error fetching VMs: "+err.Error(), http.StatusInternalServerError)
        return
    }

    servers, err := services.GetAllServers()
    if err != nil {
        http.Error(w, "Error fetching servers: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Find the specific VM
    var vm *models.VM
    for i := range vms {
        if vms[i].ID == vmID {
            vm = &vms[i]
            break
        }
    }
    
    if vm == nil {
        http.Error(w, "VM not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "text/html")
    
    // Identify host server (if linked)
    var hostServer *models.Server
    if vm.HostServerID != "" {
        for i := range servers {
            if servers[i].ID == vm.HostServerID {
                hostServer = &servers[i]
                break
            }
        }
    }

    data := map[string]interface{}{
        "vm": vm,
        "hostServer": hostServer,
    }
    
    // Template is defined as "vm-detail" in vm-detail.html
    if err := templates.ExecuteTemplate(w, "vm-detail", data); err != nil {
        log.Printf("Error rendering template: %v", err)
    }
}