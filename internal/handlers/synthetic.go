package handlers

import (
	"html/template"
	"log"
	"net/http"

	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/services"
)

func SyntheticHandlerWithTemplates(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)

		results := services.GetSyntheticResults()

		data := map[string]interface{}{
			"Synthetics": results,
			"IsAdmin":    isAdminUser(cfg, username),
			"Username":   username,
		}

		if err := templates.ExecuteTemplate(w, "synthetics.html", data); err != nil {
			log.Printf("Error rendering synthetics template: %v", err)
			http.Error(w, "Failed to render synthetics", http.StatusInternalServerError)
		}
	}
}
