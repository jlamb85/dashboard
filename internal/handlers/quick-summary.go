package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"server-dashboard/internal/config"
	"server-dashboard/internal/middleware"
	"server-dashboard/internal/services"
)

const itemsPerPage = 25

func QuickSummaryHandlerWithTemplates(cfg *config.Config, tmplExec func(string, interface{}) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, _ := middleware.GetUsername(r)

		servers, err := services.GetAllServers()
		if err != nil {
			fmt.Fprintf(w, "Error fetching servers: %v\n", err)
			return
		}

		vms, err := services.GetAllVMs()
		if err != nil {
			fmt.Fprintf(w, "Error fetching VMs: %v\n", err)
			return
		}

		switches, err := services.GetAllSwitches()
		if err != nil {
			fmt.Fprintf(w, "Error fetching switches: %v\n", err)
			return
		}

		log.Printf("QuickSummary: servers=%d, vms=%d, switches=%d", len(servers), len(vms), len(switches))

		// Get pagination parameters from query string
		serversPage := parsePageParam(r.URL.Query().Get("servers_page"), 1)
		vmsPage := parsePageParam(r.URL.Query().Get("vms_page"), 1)
		switchesPage := parsePageParam(r.URL.Query().Get("switches_page"), 1)

		// Normalize tab selection and fallback to first available tab with data
		activeTab := normalizeTab(r.URL.Query().Get("tab"))
		if activeTab == "servers" && len(servers) == 0 {
			if len(vms) > 0 {
				activeTab = "vms"
			} else if len(switches) > 0 {
				activeTab = "switches"
			}
		}
		if activeTab == "vms" && len(vms) == 0 {
			if len(servers) > 0 {
				activeTab = "servers"
			} else if len(switches) > 0 {
				activeTab = "switches"
			}
		}
		if activeTab == "switches" && len(switches) == 0 {
			if len(servers) > 0 {
				activeTab = "servers"
			} else if len(vms) > 0 {
				activeTab = "vms"
			}
		}

		log.Printf("QuickSummary: activeTab=%s", activeTab)

		// Calculate pagination for each category
		serversStart, serversEnd := getPaginationBounds(serversPage, len(servers))
		vmsStart, vmsEnd := getPaginationBounds(vmsPage, len(vms))
		switchesStart, switchesEnd := getPaginationBounds(switchesPage, len(switches))

		// Calculate total pages
		serversTotalPages := max(1, (len(servers)+itemsPerPage-1)/itemsPerPage)
		vmsTotalPages := max(1, (len(vms)+itemsPerPage-1)/itemsPerPage)
		switchesTotalPages := max(1, (len(switches)+itemsPerPage-1)/itemsPerPage)

		// Clamp pages so out-of-range queries don’t bleed between tabs
		serversPage = clampPage(serversPage, serversTotalPages)
		vmsPage = clampPage(vmsPage, vmsTotalPages)
		switchesPage = clampPage(switchesPage, switchesTotalPages)

		// Get paginated data
		serversPageData := servers[serversStart:serversEnd]
		vmsPageData := vms[vmsStart:vmsEnd]
		switchesPageData := switches[switchesStart:switchesEnd]

		data := map[string]interface{}{
			"ActiveTab":           activeTab,
			"servers":             servers,
			"serversPage":         serversPageData,
			"ServersCurrentPage":  serversPage,
			"ServersTotalPages":   serversTotalPages,
			"vms":                 vms,
			"vmsPage":             vmsPageData,
			"VMsCurrentPage":      vmsPage,
			"VMsTotalPages":       vmsTotalPages,
			"switches":            switches,
			"switchesPage":        switchesPageData,
			"SwitchesCurrentPage": switchesPage,
			"SwitchesTotalPages":  switchesTotalPages, "IsAdmin": isAdminUser(cfg, username), "Username": username}

		html, err := tmplExec("quick-summary.html", data)
		if err != nil {
			fmt.Fprintf(w, "Template error: %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	}
}

// parsePageParam parses the page parameter and returns a valid page number (1-based)
func parsePageParam(param string, defaultPage int) int {
	if param == "" {
		return defaultPage
	}
	page, err := strconv.Atoi(param)
	if err != nil || page < 1 {
		return defaultPage
	}
	return page
}

// getPaginationBounds returns the start and end indices for pagination
func getPaginationBounds(page, totalItems int) (int, int) {
	if page < 1 {
		page = 1
	}

	start := (page - 1) * itemsPerPage
	end := start + itemsPerPage

	if start > totalItems {
		start = totalItems
	}
	if end > totalItems {
		end = totalItems
	}

	return start, end
}

func clampPage(page, totalPages int) int {
	if totalPages < 1 {
		return 1
	}
	if page < 1 {
		return 1
	}
	if page > totalPages {
		return totalPages
	}
	return page
}

func normalizeTab(tab string) string {
	t := strings.ToLower(strings.TrimSpace(tab))
	switch t {
	case "servers", "vms", "switches":
		return t
	default:
		return "servers"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
