package handlers

import (
	"html/template"
	"net/http"
	"time"

	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/services"
)

// SyntheticCheckData represents a single synthetic check
type SyntheticCheckData struct {
	ID        string
	Name      string
	Type      string
	Target    string
	Status    string // ok, fail
	Message   string
	LatencyMs int
	LastRun   time.Time
	Tags      []string
}

// SyntheticsPageData holds data for the synthetics page
type SyntheticsPageData struct {
	Synthetics           []SyntheticCheckData
	UIShowSynthetics     bool
	UIEnableAutoRefresh  bool
	UIAutoRefreshSeconds int
	AppVersion           string
	CurrentYear          int
	IsAdmin              bool
	Username             string
	ServerCount          int
	VMCount              int
	SwitchCount          int
}

// SyntheticsPageHandlerWithTemplates returns a handler for the synthetics page
func SyntheticsPageHandlerWithTemplates(cfg *config.Config, templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)
		results := services.GetSyntheticResults()
		servers, _ := services.GetAllServers()
		vms, _ := services.GetAllVMs()
		switches, _ := services.GetAllSwitches()
		data := SyntheticsPageData{
			Synthetics:           nil,
			UIShowSynthetics:     true,
			UIEnableAutoRefresh:  true,
			UIAutoRefreshSeconds: 10,
			AppVersion:           "v1.0.1-dev",
			CurrentYear:          time.Now().Year(),
			IsAdmin:              isAdminUser(cfg, username),
			Username:             username,
			ServerCount:          len(servers),
			VMCount:              len(vms),
			SwitchCount:          len(switches),
		}
		// Convert real results to SyntheticCheckData for template compatibility
		for _, s := range results {
			data.Synthetics = append(data.Synthetics, SyntheticCheckData{
				ID:        s.ID,
				Name:      s.Name,
				Type:      s.Type,
				Target:    s.Target,
				Status:    s.Status,
				Message:   s.Message,
				LatencyMs: int(s.LatencyMs),
				LastRun:   s.LastRun,
				Tags:      s.Tags,
			})
		}
		if err := templates.ExecuteTemplate(w, "synthetics.html", data); err != nil {
			// Can't write error response, header already sent by template execution
			return
		}
	}
}
