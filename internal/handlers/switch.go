package handlers

import (
	"html/template"
	"net/http"
	"server-dashboard/internal/services"

	"github.com/gorilla/mux"
)

// SwitchesHandler shows the list of all switches
func SwitchesHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switches, err := services.GetAllSwitches()
		if err != nil {
			http.Error(w, "Failed to retrieve switches", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"switches": switches,
		}

		if err := templates.ExecuteTemplate(w, "switches.html", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

// SwitchDetailHandler shows details for a specific switch
func SwitchDetailHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get switch ID from URL path parameters
		vars := mux.Vars(r)
		switchID := vars["id"]
		if switchID == "" {
			http.Error(w, "Switch ID is required", http.StatusBadRequest)
			return
		}

		switches, err := services.GetAllSwitches()
		if err != nil {
			http.Error(w, "Failed to retrieve switches", http.StatusInternalServerError)
			return
		}

		var targetSwitch *interface{}
		for i := range switches {
			if switches[i].ID == switchID {
				var iface interface{} = switches[i]
				targetSwitch = &iface
				break
			}
		}

		if targetSwitch == nil {
			http.Error(w, "Switch not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{
			"switch": *targetSwitch,
		}

		if err := templates.ExecuteTemplate(w, "switch-detail.html", data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}
