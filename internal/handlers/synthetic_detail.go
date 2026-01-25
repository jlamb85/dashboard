package handlers

import (
	"html/template"
	"log"
	"net/http"

	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/services"

	"github.com/gorilla/mux"
)

func SyntheticDetailHandler(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		username, _ := middleware.GetUsername(r)

		results := services.GetSyntheticResults()
		var found interface{}
		for _, s := range results {
			if s.ID == id {
				found = s
				break
			}
		}
		if found == nil {
			http.NotFound(w, r)
			return
		}
		data := map[string]interface{}{
			"Synthetic": found,
			"IsAdmin":   isAdminUser(cfg, username),
			"Username":  username,
		}
		if err := templates.ExecuteTemplate(w, "synthetic-detail.html", data); err != nil {
			log.Printf("Error rendering synthetic detail: %v", err)
			http.Error(w, "Failed to render detail", http.StatusInternalServerError)
		}
	}
}
