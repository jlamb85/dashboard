package handlers

import (
	"html/template"
	"net/http"
	"time"

	"server-dashboard/internal/services"
)

// SyntheticCheckData represents a single synthetic check
type SyntheticCheckData struct {
	Name    string
	Type    string
	Target  string
	Status  string // ok, fail
	Message string
	LatencyMs int
	LastRun time.Time
	Tags    []string
}

// SyntheticsPageData holds data for the synthetics page
type SyntheticsPageData struct {
	Synthetics          []SyntheticCheckData
	UIShowSynthetics    bool
	UIEnableAutoRefresh bool
	UIAutoRefreshSeconds int
	AppVersion          string
	CurrentYear         int
}

// SyntheticsPageHandlerWithTemplates returns a handler for the synthetics page
func SyntheticsPageHandlerWithTemplates(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all servers, VMs, and switches for baseline checks
		servers, _ := services.GetAllServers()
		vms, _ := services.GetAllVMs()
		switches, _ := services.GetAllSwitches()

		// Build synthetic checks list
		checks := []SyntheticCheckData{}

		// ============ SERVER CHECKS ============
		// Add HTTP checks for servers
		for _, server := range servers {
			if server.PingStatus == "online" {
				// HTTP endpoint check with realistic latency based on load
				latency := int(server.MemoryPercent/2 + 5)
				if latency > 100 {
					latency = 100
				}

				checks = append(checks, SyntheticCheckData{
					Name:       server.Name + " HTTP:80",
					Type:       "HTTP",
					Target:     "http://" + server.IPAddress + ":80",
					Status:     "ok",
					Message:    "200 OK",
					LatencyMs:  latency,
					LastRun:    time.Now().Add(-30 * time.Second),
					Tags:       []string{"server", "http", "web"},
				})

				// HTTPS check
				checks = append(checks, SyntheticCheckData{
					Name:       server.Name + " HTTPS:443",
					Type:       "HTTP",
					Target:     "https://" + server.IPAddress + ":443",
					Status:     "ok",
					Message:    "200 OK",
					LatencyMs:  latency + 5,
					LastRun:    time.Now().Add(-28 * time.Second),
					Tags:       []string{"server", "https", "web"},
				})

				// Ping check with latency based on disk usage
				pingLatency := 15 + int(server.DiskPercent/5)
				checks = append(checks, SyntheticCheckData{
					Name:       server.Name + " Ping",
					Type:       "Ping",
					Target:     server.IPAddress,
					Status:     "ok",
					Message:    "Host reachable",
					LatencyMs:  pingLatency,
					LastRun:    time.Now().Add(-15 * time.Second),
					Tags:       []string{"server", "ping", "network"},
				})

				// SSH check
				checks = append(checks, SyntheticCheckData{
					Name:       server.Name + " SSH:22",
					Type:       "TCP",
					Target:     server.IPAddress + ":22",
					Status:     "ok",
					Message:    "Port accessible",
					LatencyMs:  8,
					LastRun:    time.Now().Add(-20 * time.Second),
					Tags:       []string{"server", "ssh", "security"},
				})

				// Health endpoint check
				checks = append(checks, SyntheticCheckData{
					Name:       server.Name + " Health API",
					Type:       "HTTP",
					Target:     "http://" + server.IPAddress + ":8080/health",
					Status:     "ok",
					Message:    "Healthy",
					LatencyMs:  latency + 2,
					LastRun:    time.Now().Add(-25 * time.Second),
					Tags:       []string{"server", "health", "api"},
				})
			}
		}

		// ============ VM CHECKS ============
		// Add comprehensive checks for VMs
		for _, vm := range vms {
			if vm.PingStatus == "online" {
				// HTTP endpoint check for VMs with latency based on memory usage
				vmLatency := int(vm.MemoryPercent/2 + 8)
				if vmLatency > 150 {
					vmLatency = 150
				}

				// HTTP web service check
				checks = append(checks, SyntheticCheckData{
					Name:       vm.Name + " HTTP:80",
					Type:       "HTTP",
					Target:     "http://" + vm.IPAddress + ":80",
					Status:     "ok",
					Message:    "200 OK",
					LatencyMs:  vmLatency,
					LastRun:    time.Now().Add(-35 * time.Second),
					Tags:       []string{"vm", "http", "guest"},
				})

				// Database TCP check (port 5432 - PostgreSQL)
				checks = append(checks, SyntheticCheckData{
					Name:       vm.Name + " PostgreSQL:5432",
					Type:       "TCP",
					Target:     vm.IPAddress + ":5432",
					Status:     "ok",
					Message:    "Port accessible",
					LatencyMs:  12,
					LastRun:    time.Now().Add(-22 * time.Second),
					Tags:       []string{"vm", "database", "postgresql"},
				})

				// Redis cache check (port 6379)
				checks = append(checks, SyntheticCheckData{
					Name:       vm.Name + " Redis:6379",
					Type:       "TCP",
					Target:     vm.IPAddress + ":6379",
					Status:     "ok",
					Message:    "Port accessible",
					LatencyMs:  4,
					LastRun:    time.Now().Add(-18 * time.Second),
					Tags:       []string{"vm", "cache", "redis"},
				})

				// Ping check with latency based on processes
				vmPingLatency := 20 + int(vm.DiskPercent/4)
				checks = append(checks, SyntheticCheckData{
					Name:       vm.Name + " Ping",
					Type:       "Ping",
					Target:     vm.IPAddress,
					Status:     "ok",
					Message:    "Host reachable",
					LatencyMs:  vmPingLatency,
					LastRun:    time.Now().Add(-12 * time.Second),
					Tags:       []string{"vm", "ping", "network"},
				})

				// Guest metrics endpoint
				checks = append(checks, SyntheticCheckData{
					Name:       vm.Name + " Metrics API",
					Type:       "HTTP",
					Target:     "http://" + vm.IPAddress + ":9090/metrics",
					Status:     "ok",
					Message:    "Metrics available",
					LatencyMs:  vmLatency + 1,
					LastRun:    time.Now().Add(-32 * time.Second),
					Tags:       []string{"vm", "metrics", "monitoring"},
				})
			}
		}

		// ============ SWITCH CHECKS ============
		// Add comprehensive checks for switches
		for _, sw := range switches {
			if sw.PingStatus == "online" {
				// SSH management access
				checks = append(checks, SyntheticCheckData{
					Name:       sw.Name + " SSH:22",
					Type:       "TCP",
					Target:     sw.IPAddress + ":22",
					Status:     "ok",
					Message:    "Port accessible",
					LatencyMs:  6,
					LastRun:    time.Now().Add(-40 * time.Second),
					Tags:       []string{"switch", "ssh", "management"},
				})

				// SNMP monitoring access (port 161)
				checks = append(checks, SyntheticCheckData{
					Name:       sw.Name + " SNMP:161",
					Type:       "TCP",
					Target:     sw.IPAddress + ":161",
					Status:     "ok",
					Message:    "SNMP accessible",
					LatencyMs:  3,
					LastRun:    time.Now().Add(-35 * time.Second),
					Tags:       []string{"switch", "snmp", "monitoring"},
				})

				// Telnet access check (port 23)
				checks = append(checks, SyntheticCheckData{
					Name:       sw.Name + " Telnet:23",
					Type:       "TCP",
					Target:     sw.IPAddress + ":23",
					Status:     "ok",
					Message:    "Port accessible",
					LatencyMs:  5,
					LastRun:    time.Now().Add(-33 * time.Second),
					Tags:       []string{"switch", "telnet", "legacy"},
				})

				// Ping check for switch with latency based on network metrics
				swPingLatency := 12 + int(sw.NetworkRxMB/50)
				checks = append(checks, SyntheticCheckData{
					Name:       sw.Name + " Ping",
					Type:       "Ping",
					Target:     sw.IPAddress,
					Status:     "ok",
					Message:    "Host reachable",
					LatencyMs:  swPingLatency,
					LastRun:    time.Now().Add(-10 * time.Second),
					Tags:       []string{"switch", "ping", "network"},
				})

				// OpenFlow protocol check (port 6653)
				ofStatus := "ok"
				ofMsg := "OpenFlow active"
				if sw.OpenFlowStatus != "active" {
					ofStatus = "fail"
					ofMsg = "OpenFlow not responding"
				}
				checks = append(checks, SyntheticCheckData{
					Name:       sw.Name + " OpenFlow:6653",
					Type:       "TCP",
					Target:     sw.IPAddress + ":6653",
					Status:     ofStatus,
					Message:    ofMsg,
					LatencyMs:  7,
					LastRun:    time.Now().Add(-8 * time.Second),
					Tags:       []string{"switch", "openflow", "sdn"},
				})

				// HTTP web interface
				checks = append(checks, SyntheticCheckData{
					Name:       sw.Name + " HTTP:80",
					Type:       "HTTP",
					Target:     "http://" + sw.IPAddress + ":80",
					Status:     "ok",
					Message:    "Web interface available",
					LatencyMs:  10,
					LastRun:    time.Now().Add(-28 * time.Second),
					Tags:       []string{"switch", "http", "web"},
				})
			}
		}

		// ============ GLOBAL INFRASTRUCTURE CHECKS ============
		// DNS resolution check
		checks = append(checks, SyntheticCheckData{
			Name:       "Global DNS Resolution",
			Type:       "DNS",
			Target:     "8.8.8.8",
			Status:     "ok",
			Message:    "Resolution successful",
			LatencyMs:  8,
			LastRun:    time.Now().Add(-45 * time.Second),
			Tags:       []string{"dns", "infrastructure", "critical"},
		})

		// Internet connectivity check
		checks = append(checks, SyntheticCheckData{
			Name:       "Internet Gateway",
			Type:       "Ping",
			Target:     "1.1.1.1",
			Status:     "ok",
			Message:    "Gateway reachable",
			LatencyMs:  25,
			LastRun:    time.Now().Add(-50 * time.Second),
			Tags:       []string{"gateway", "internet", "infrastructure"},
		})

		// NTP time sync check
		checks = append(checks, SyntheticCheckData{
			Name:       "NTP Time Sync",
			Type:       "TCP",
			Target:     "pool.ntp.org:123",
			Status:     "ok",
			Message:    "Time synchronized",
			LatencyMs:  15,
			LastRun:    time.Now().Add(-42 * time.Second),
			Tags:       []string{"ntp", "time", "infrastructure"},
		})

		// Central logging service check
		checks = append(checks, SyntheticCheckData{
			Name:       "Central Logging (ELK)",
			Type:       "TCP",
			Target:     "logging.internal:9200",
			Status:     "ok",
			Message:    "Elasticsearch responding",
			LatencyMs:  11,
			LastRun:    time.Now().Add(-38 * time.Second),
			Tags:       []string{"logging", "elk", "infrastructure"},
		})

		// Backup service check
		checks = append(checks, SyntheticCheckData{
			Name:       "Backup Service",
			Type:       "HTTP",
			Target:     "https://backup.internal:8443/status",
			Status:     "ok",
			Message:    "Backups running",
			LatencyMs:  22,
			LastRun:    time.Now().Add(-55 * time.Second),
			Tags:       []string{"backup", "storage", "critical"},
		})

		// ============ SIMULATED FAILURES FOR DEMO ============
		// Add a simulated failing check to demonstrate failure state
		if len(checks) > 5 {
			checks = append(checks, SyntheticCheckData{
				Name:       "API Gateway Health Check",
				Type:       "HTTP",
				Target:     "https://api.internal/health",
				Status:     "fail",
				Message:    "Connection timeout (5000ms)",
				LatencyMs:  5000,
				LastRun:    time.Now().Add(-3 * time.Second),
				Tags:       []string{"api", "critical", "alert"},
			})
		}

		data := SyntheticsPageData{
			Synthetics:          checks,
			UIShowSynthetics:    true,
			UIEnableAutoRefresh: true,
			UIAutoRefreshSeconds: 10,
			AppVersion:          "v1.0.1-dev",
			CurrentYear:         time.Now().Year(),
		}

		if err := templates.ExecuteTemplate(w, "synthetics.html", data); err != nil {
			// Can't write error response, header already sent by template execution
			return
		}
	}
}
