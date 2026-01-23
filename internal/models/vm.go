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
	Tags           []string       `json:"tags"`
	// Memory metrics
	MemoryUsed     float64        `json:"memory_used"`     // Memory used in MB
	MemoryTotal    float64        `json:"memory_total"`    // Total memory in MB
	MemoryPercent  float64        `json:"memory_percent"`  // Memory usage percentage
	// Load average
	LoadAverage    string         `json:"load_average"`    // 1, 5, 15 minute load averages
	// Failed services
	FailedServices int            `json:"failed_services"` // Count of failed systemd services
	// Inode usage
	InodeUsed      int64          `json:"inode_used"`      // Inodes used on root partition
	InodeTotal     int64          `json:"inode_total"`     // Total inodes on root partition
	InodePercent   float64        `json:"inode_percent"`   // Inode usage percentage
	// Network stats
	NetworkRxMB    float64        `json:"network_rx_mb"`   // Network received in MB
	NetworkTxMB    float64        `json:"network_tx_mb"`   // Network transmitted in MB
	// System info
	KernelVersion  string         `json:"kernel_version"`  // Linux kernel version
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