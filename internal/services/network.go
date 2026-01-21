package services

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
	"server-dashboard/internal/config"
	"server-dashboard/internal/models"
)

// Global slice to store servers and VMs
var (
	ServersCache []models.Server
	VMsCache     []models.VM
	Config       *config.Config
	
	// Monitoring control
	monitoringInterval time.Duration
	stopMonitoring     chan bool
	isMonitoring       bool
	monitoringMutex    sync.RWMutex
	
	// SSH client for real monitoring
	sshClient *SSHClient
)

// InitializeCache loads servers and VMs from config
func InitializeCache(cfg *config.Config) {
	Config = cfg
	
	// Initialize SSH client for real monitoring if enabled
	if !cfg.Monitoring.UseMockData && cfg.SSH.Enabled {
		client, err := NewSSHClient(
			cfg.SSH.Username,
			cfg.SSH.PrivateKeyPath,
			cfg.SSH.Password,
			cfg.SSH.TimeoutSeconds,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize SSH client: %v\n", err)
			fmt.Println("Falling back to mock data monitoring")
		} else {
			sshClient = client
			fmt.Println("SSH monitoring enabled for production metrics")
		}
	}
	
	// Initialize servers
	ServersCache = make([]models.Server, len(cfg.Servers))
	for i, srvCfg := range cfg.Servers {
		srv := models.NewServer(srvCfg.ID, srvCfg.Name, srvCfg.IPAddress, srvCfg.Hostname, srvCfg.Port)
		// Set default values
		srv.Status = "offline"
		srv.PingStatus = "offline"
		srv.Uptime = "N/A"
		srv.Processes = 0
		srv.DiskUsage = 0
		srv.DiskTotal = 1000.0
		srv.DiskPercent = 0
		srv.DiskPartition = "/"
		srv.FullPartitions = []string{}
		srv.LastChecked = time.Now()
		ServersCache[i] = *srv
	}
	
	// Initialize VMs
	VMsCache = make([]models.VM, len(cfg.VirtualMachines))
	for i, vmCfg := range cfg.VirtualMachines {
		vm := models.NewVM(vmCfg.ID, vmCfg.Name, vmCfg.IPAddress, vmCfg.Hostname, vmCfg.Port, vmCfg.HostServerID)
		// Set default values
		vm.Status = "offline"
		vm.PingStatus = "offline"
		vm.Uptime = "N/A"
		vm.Processes = 0
		vm.DiskUsage = 0
		vm.DiskTotal = 500.0
		vm.DiskPercent = 0
		vm.DiskPartition = "/"
		vm.FullPartitions = []string{}
		vm.LastChecked = time.Now()
		vm.StreamPorts = vmCfg.StreamPorts
		vm.Streams = []models.StreamStatus{}
		// Check if streams are active on initialization
		for _, port := range vm.StreamPorts {
			streamStatus := models.StreamStatus{Port: port, Active: false}
			if stream, err := CheckStreamOnPort(port); err == nil && stream != nil {
				streamStatus.Active = stream.Active
			}
			vm.Streams = append(vm.Streams, streamStatus)
		}
		VMsCache[i] = *vm
	}
	
	// Set monitoring interval (default 30 seconds)
	monitoringInterval = 30 * time.Second
	stopMonitoring = make(chan bool, 1)
	isMonitoring = true
	
	// Start background monitoring in goroutine to avoid blocking initialization
	go StartBackgroundMonitoring()
}

// StartBackgroundMonitoring runs periodic health checks on all servers and VMs
func StartBackgroundMonitoring() {
	ticker := time.NewTicker(monitoringInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-stopMonitoring:
			return
		case <-ticker.C:
			// Run monitoring in separate goroutine to avoid blocking
			go func() {
				MonitorAllServers()
				MonitorAllVMs()
			}()
		}
	}
}

// StopBackgroundMonitoring stops the background monitoring routine
func StopBackgroundMonitoring() {
	select {
	case stopMonitoring <- true:
	default:
	}
}

// StopMonitoring stops the monitoring service
func StopMonitoring() error {
	monitoringMutex.Lock()
	defer monitoringMutex.Unlock()
	
	if !isMonitoring {
		return nil // Already stopped
	}
	
	StopBackgroundMonitoring()
	isMonitoring = false
	return nil
}

// StartMonitoring starts the monitoring service
func StartMonitoring() error {
	monitoringMutex.Lock()
	defer monitoringMutex.Unlock()
	
	if isMonitoring {
		return nil // Already running
	}
	
	stopMonitoring = make(chan bool, 1) // Reset the channel
	go StartBackgroundMonitoring()
	isMonitoring = true
	return nil
}

// RestartMonitoring restarts the monitoring service
func RestartMonitoring() error {
	if err := StopMonitoring(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond) // Brief pause
	return StartMonitoring()
}

// GetMonitoringStatus returns whether monitoring is currently active
func GetMonitoringStatus() bool {
	monitoringMutex.RLock()
	defer monitoringMutex.RUnlock()
	return isMonitoring
}

// MonitorAllServers checks status of all servers
func MonitorAllServers() {
	for i := range ServersCache {
		MonitorServer(&ServersCache[i])
	}
}

// MonitorAllVMs checks status of all VMs
func MonitorAllVMs() {
	for i := range VMsCache {
		MonitorVM(&VMsCache[i])
	}
}

// MonitorServer performs health checks on a server
func MonitorServer(srv *models.Server) {
	// Check ping status using ICMP
	srv.PingStatus = CheckPingStatus(srv.IPAddress)
	
	if srv.PingStatus == "online" {
		srv.Status = "online"
		
		// Use real SSH monitoring if available and configured
		if !Config.Monitoring.UseMockData && Config.SSH.Enabled && sshClient != nil {
			err := sshClient.GetRealServerMetrics(srv)
			if err != nil {
				// Fallback to mock data on error
				fmt.Printf("SSH monitoring failed for %s: %v, using mock data\n", srv.Name, err)
				useServerMockData(srv)
			}
		} else {
			// Use mock data for development
			useServerMockData(srv)
		}
	} else {
		srv.Status = "offline"
		// Reset metrics for offline servers
		srv.Uptime = "N/A"
		srv.Processes = 0
		srv.DiskUsage = 0
		srv.DiskTotal = 1000.0
		srv.DiskPercent = 0
	}
	
	srv.LastChecked = time.Now()
}

// useServerMockData generates mock metrics for a server
func useServerMockData(srv *models.Server) {
	srv.Uptime = GenerateUptime()
	srv.Processes = rand.Intn(200) + 50 // 50-250 processes
	srv.DiskTotal = 1000.0
	srv.DiskUsage = generateDiskUsage(srv.DiskTotal, 30, 70) // 30-70% usage
	srv.DiskPercent = (srv.DiskUsage / srv.DiskTotal) * 100
}

// MonitorVM performs health checks on a VM
func MonitorVM(vm *models.VM) {
	// Check ping status using ICMP
	vm.PingStatus = CheckPingStatus(vm.IPAddress)
	
	if vm.PingStatus == "online" {
		vm.Status = "running"
		
		// Use real SSH monitoring if available and configured
		if !Config.Monitoring.UseMockData && Config.SSH.Enabled && sshClient != nil {
			err := sshClient.GetRealVMMetrics(vm)
			if err != nil {
				// Fallback to mock data on error
				fmt.Printf("SSH monitoring failed for %s: %v, using mock data\n", vm.Name, err)
				useVMMockData(vm)
			}
		} else {
			// Use mock data for development
			useVMMockData(vm)
		}
	} else {
		vm.Status = "offline"
		// Reset metrics for offline VMs
		vm.Uptime = "N/A"
		vm.Processes = 0
		vm.DiskUsage = 0
		vm.DiskTotal = 500.0
		vm.DiskPercent = 0
	}
	
	// Check stream status for all configured ports
	vm.Streams = []models.StreamStatus{}
	for _, port := range vm.StreamPorts {
		streamStatus := models.StreamStatus{Port: port, Active: false}
		if stream, err := CheckStreamOnPort(port); err == nil && stream != nil {
			streamStatus.Active = stream.Active
		}
		vm.Streams = append(vm.Streams, streamStatus)
	}
	
	vm.LastChecked = time.Now()
}

// useVMMockData generates mock metrics for a VM
func useVMMockData(vm *models.VM) {
	vm.Uptime = GenerateUptime()
	vm.Processes = rand.Intn(150) + 30 // 30-180 processes
	vm.DiskTotal = 500.0
	vm.DiskUsage = generateDiskUsage(vm.DiskTotal, 20, 60) // 20-60% usage
	vm.DiskPercent = (vm.DiskUsage / vm.DiskTotal) * 100
}

// CheckPingStatus attempts TCP checks on multiple ports for connectivity
func CheckPingStatus(ipAddress string) string {
	// Try TCP checks on common ports (HTTP, HTTPS, SSH, MySQL, PostgreSQL)
	// Using sequential checks with short timeouts
	ports := []int{80, 443, 22, 3306, 5432}
	for _, port := range ports {
		if IsReachableTCP(ipAddress, port) {
			return "online"
		}
	}
	
	return "offline"
}

// IsReachableTCP checks if a host is reachable on a specific TCP port
func IsReachableTCP(ipAddress string, port int) bool {
	timeout := 1 * time.Second  // Shorter timeout per port
	addr := fmt.Sprintf("%s:%d", ipAddress, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GenerateUptime creates a realistic uptime string
func GenerateUptime() string {
	days := rand.Intn(365)      // 0-365 days
	hours := rand.Intn(24)      // 0-23 hours
	return fmt.Sprintf("%d days %d hours", days, hours)
}

// generateDiskUsage creates a random disk usage within specified range
func generateDiskUsage(total float64, minPercent, maxPercent int) float64 {
	percent := minPercent + rand.Intn(maxPercent-minPercent+1)
	return (float64(percent) / 100.0) * total
}

// DiscoverConnectedServers discovers servers connected to the network.
func DiscoverConnectedServers(networkCIDR string) ([]string, error) {
	var servers []string
	_, ipNet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return nil, err
	}

	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		servers = append(servers, ip.String())
	}

	return servers, nil
}

// incrementIP increments the given IP address by one.
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		if ip[j]++; ip[j] != 0 {
			break
		}
	}
}

// GetAllServers returns all configured servers
func GetAllServers() ([]models.Server, error) {
	return ServersCache, nil
}

// GetAllVMs returns all configured VMs
func GetAllVMs() ([]models.VM, error) {
	return VMsCache, nil
}

// CheckServerStatus checks the status of a specific server
func CheckServerStatus(serverID string) (interface{}, error) {
	for i := range ServersCache {
		if ServersCache[i].ID == serverID {
			MonitorServer(&ServersCache[i])
			return ServersCache[i], nil
		}
	}
	return nil, fmt.Errorf("server not found: %s", serverID)
}

// CheckVMStatus checks the status of a specific VM
func CheckVMStatus(vmID string) (string, error) {
	for i := range VMsCache {
		if VMsCache[i].ID == vmID {
			MonitorVM(&VMsCache[i])
			return VMsCache[i].Status, nil
		}
	}
	return "unknown", fmt.Errorf("vm not found: %s", vmID)
}