package handlers

import (
    "log"
    "net/http"
    "html/template"

    "server-dashboard/internal/services"
)

func SyntheticHandlerWithTemplates(w http.ResponseWriter, r *http.Request, templates *template.Template) {
    results := services.GetSyntheticResults()

    data := map[string]interface{}{
        "synthetics": results,
    }

    if err := templates.ExecuteTemplate(w, "synthetics.html", data); err != nil {
        log.Printf("Error rendering synthetics template: %v", err)
        http.Error(w, "Failed to render synthetics", http.StatusInternalServerError)
    }
}
