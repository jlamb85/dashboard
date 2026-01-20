package models

import (
	"time"
)

// Server represents a server in the network.
type Server struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	IPAddress      string    `json:"ip_address"`
	Hostname       string    `json:"hostname"`
	Port           int       `json:"port"`
	Status         string    `json:"status"`
	PingStatus     string    `json:"ping_status"`
	Uptime         string    `json:"uptime"`
	Processes      int       `json:"processes"`
	DiskUsage      float64   `json:"disk_usage"`
	DiskTotal      float64   `json:"disk_total"`
	DiskPercent    float64   `json:"disk_percent"`
	DiskPartition  string    `json:"disk_partition"`  // Primary partition being monitored (usually /)
	FullPartitions []string  `json:"full_partitions"` // List of partitions over 90% full
	LastChecked    time.Time `json:"last_checked"`
}

// NewServer creates a new Server instance.
func NewServer(id, name, ipAddress, hostname string, port int) *Server {
	return &Server{
		ID:          id,
		Name:        name,
		IPAddress:   ipAddress,
		Hostname:    hostname,
		Port:        port,
		Status:      "unknown",
		PingStatus:  "unknown",
		Uptime:      "N/A",
		Processes:   0,
		DiskUsage:   0,
		DiskTotal:   0,
		DiskPercent: 0,
		LastChecked: time.Now(),
	}
}

// CheckStatus updates the server's status.
func (s *Server) CheckStatus(status string) {
	s.Status = status
	s.LastChecked = time.Now()
}