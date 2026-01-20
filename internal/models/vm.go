package models

import (
	"time"
)

type StreamStatus struct {
	Port   int  `json:"port"`
	Active bool `json:"active"`
}

type VM struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	IPAddress      string         `json:"ip_address"`
	Hostname       string         `json:"hostname"`
	Port           int            `json:"port"`
	HostServerID   string         `json:"host_server_id"`
	StreamPorts    []int          `json:"stream_ports"`    // Ports for video/media streaming
	Streams        []StreamStatus `json:"streams"`         // Status of each stream port
	Status         string         `json:"status"`
	PingStatus     string         `json:"ping_status"`
	Uptime         string         `json:"uptime"`
	Processes      int            `json:"processes"`
	DiskUsage      float64        `json:"disk_usage"`
	DiskTotal      float64        `json:"disk_total"`
	DiskPercent    float64        `json:"disk_percent"`
	DiskPartition  string         `json:"disk_partition"`  // Primary partition being monitored (usually /)
	FullPartitions []string       `json:"full_partitions"` // List of partitions over 90% full
	LastChecked    time.Time      `json:"last_checked"`
}

func NewVM(id, name, ipAddress, hostname string, port int, hostServerID string) *VM {
	return &VM{
		ID:          id,
		Name:        name,
		IPAddress:   ipAddress,
		Hostname:    hostname,
		Port:        port,
		HostServerID: hostServerID,
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

func (vm *VM) CheckStatus() string {
	return vm.Status
}

func (vm *VM) UpdateStatus(newStatus string) {
	vm.Status = newStatus
	vm.LastChecked = time.Now()
}