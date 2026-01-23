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
	SwitchesCache []models.Switch
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
		srv.Tags = append([]string{}, srvCfg.Tags...)
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
	
	// Initialize Switches
	SwitchesCache = make([]models.Switch, len(cfg.Switches))
	for i, swCfg := range cfg.Switches {
		sw := models.NewSwitch(swCfg.ID, swCfg.Name, swCfg.IPAddress, swCfg.Hostname, swCfg.Port)
		// Set default values
		sw.Status = "offline"
		sw.PingStatus = "offline"
		sw.Uptime = "N/A"
		sw.Processes = 0
		sw.DiskUsage = 0
		sw.DiskTotal = 50.0 // Switches typically have small storage
		sw.DiskPercent = 0
		sw.DiskPartition = "/"
		sw.FullPartitions = []string{}
		sw.ControllerIP = swCfg.ControllerIP
		sw.OpenFlowVersion = swCfg.OpenFlowVersion
		sw.OpenFlowStatus = "unknown"
		sw.FlowCount = 0
		sw.PortCount = 0
		sw.Tags = append([]string{}, swCfg.Tags...)
		sw.LastChecked = time.Now()
		SwitchesCache[i] = *sw
	}
	
	// Set monitoring interval (default 5 seconds)
	monitoringInterval = 5 * time.Second
	stopMonitoring = make(chan bool, 1)
	isMonitoring = true

	// Start synthetic checks
	InitSynthetic(cfg)
	
	// Run initial monitoring check immediately (non-blocking)
	go func() {
		MonitorAllServers()
		MonitorAllVMs()
		MonitorAllSwitches()
	}()
	
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
				MonitorAllSwitches()
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

// getSwitchSSHClient creates a switch-specific SSH client or returns the global client
func getSwitchSSHClient(sw *models.Switch) *SSHClient {
	// Find the switch config to get credentials
	var swCfg *config.SwitchConfig
	for _, cfg := range Config.Switches {
		if cfg.ID == sw.ID {
			swCfg = &cfg
			break
		}
	}
	
	if swCfg == nil {
		// Config not found, use global client
		return sshClient
	}
	
	// Check if switch has custom SSH credentials
	hasCustomCreds := swCfg.SSHUsername != "" || swCfg.SSHPassword != "" || swCfg.SSHKeyPath != ""
	
	if !hasCustomCreds {
		// No custom credentials, use global client
		return sshClient
	}
	
	// Use switch-specific credentials, fall back to global for missing values
	username := swCfg.SSHUsername
	if username == "" {
		username = Config.SSH.Username
	}
	
	password := swCfg.SSHPassword
	if password == "" {
		password = Config.SSH.Password
	}
	
	keyPath := swCfg.SSHKeyPath
	if keyPath == "" {
		keyPath = Config.SSH.PrivateKeyPath
	}
	
	timeout := Config.SSH.TimeoutSeconds
	
	// Create switch-specific SSH client
	client, err := NewSSHClient(username, keyPath, password, timeout)
	if err != nil {
		fmt.Printf("Warning: Failed to create SSH client for switch %s: %v, falling back to global client\n", sw.Name, err)
		return sshClient
	}
	
	return client
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

// MonitorAllSwitches checks status of all switches
func MonitorAllSwitches() {
	for i := range SwitchesCache {
		MonitorSwitch(&SwitchesCache[i])
	}
}

// MonitorServer performs health checks on a server
func MonitorServer(srv *models.Server) {
	// When using mock data, always show as online (no network checks needed)
	if Config.Monitoring.UseMockData {
		srv.PingStatus = "online"
		srv.Status = "online"
		useServerMockData(srv)
		srv.LastChecked = time.Now()
		return
	}
	
	// Production mode: Check ping status using TCP connectivity
	srv.PingStatus = CheckPingStatus(srv.IPAddress)
	
	if srv.PingStatus == "online" {
		srv.Status = "online"
		
		// Use real SSH monitoring if available and configured
		if Config.SSH.Enabled && sshClient != nil {
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
	
	// Memory metrics
	srv.MemoryTotal = float64(rand.Intn(48000) + 16000) // 16-64 GB in MB
	srv.MemoryUsed = generateDiskUsage(srv.MemoryTotal, 40, 80) // 40-80% usage
	srv.MemoryPercent = (srv.MemoryUsed / srv.MemoryTotal) * 100
	
	// Load average - generate realistic values
	load1 := float64(rand.Intn(400)+50) / 100.0  // 0.5-4.5
	load5 := float64(rand.Intn(350)+60) / 100.0  // 0.6-4.1
	load15 := float64(rand.Intn(300)+70) / 100.0 // 0.7-3.7
	srv.LoadAverage = fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)
	
	// Failed services - usually 0, occasionally 1-2
	if rand.Float64() < 0.2 { // 20% chance
		srv.FailedServices = rand.Intn(2) + 1
	} else {
		srv.FailedServices = 0
	}
	
	// Inode usage
	srv.InodeTotal = int64(rand.Intn(5000000) + 1000000) // 1M-6M inodes
	srv.InodeUsed = int64(float64(srv.InodeTotal) * (float64(rand.Intn(40)+20) / 100.0)) // 20-60%
	if srv.InodeTotal > 0 {
		srv.InodePercent = (float64(srv.InodeUsed) / float64(srv.InodeTotal)) * 100
	}
	
	// Network stats - cumulative MB since boot
	srv.NetworkRxMB = float64(rand.Intn(500000) + 10000) // 10GB-500GB
	srv.NetworkTxMB = float64(rand.Intn(200000) + 5000)  // 5GB-200GB
	
	// Kernel version - common versions
	kernels := []string{"5.15.0-91-generic", "6.1.0-17-amd64", "5.10.0-27-arm64", "6.5.0-14-generic"}
	srv.KernelVersion = kernels[rand.Intn(len(kernels))]
}

// MonitorVM performs health checks on a VM
func MonitorVM(vm *models.VM) {
	// When using mock data, always show as online (no network checks needed)
	if Config.Monitoring.UseMockData {
		vm.PingStatus = "online"
		vm.Status = "running"
		useVMMockData(vm)
		
		// Check stream status for configured ports
		for j, port := range vm.StreamPorts {
			if stream, err := CheckStreamOnPort(port); err == nil && stream != nil {
				if j < len(vm.Streams) {
					vm.Streams[j].Active = stream.Active
				}
			}
		}
		
		vm.LastChecked = time.Now()
		return
	}
	
	// Production mode: Check ping status using TCP connectivity
	vm.PingStatus = CheckPingStatus(vm.IPAddress)
	
	if vm.PingStatus == "online" {
		vm.Status = "running"
		
		// Use real SSH monitoring if available and configured
		if Config.SSH.Enabled && sshClient != nil {
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
	
	// Memory metrics
	vm.MemoryTotal = float64(rand.Intn(24000) + 4000) // 4-28 GB in MB
	vm.MemoryUsed = generateDiskUsage(vm.MemoryTotal, 35, 75) // 35-75% usage
	vm.MemoryPercent = (vm.MemoryUsed / vm.MemoryTotal) * 100
	
	// Load average - generate realistic values
	load1 := float64(rand.Intn(250)+30) / 100.0  // 0.3-2.8
	load5 := float64(rand.Intn(220)+40) / 100.0  // 0.4-2.6
	load15 := float64(rand.Intn(200)+50) / 100.0 // 0.5-2.5
	vm.LoadAverage = fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)
	
	// Failed services - usually 0, occasionally 1
	if rand.Float64() < 0.15 { // 15% chance
		vm.FailedServices = 1
	} else {
		vm.FailedServices = 0
	}
	
	// Inode usage
	vm.InodeTotal = int64(rand.Intn(3000000) + 500000) // 500K-3.5M inodes
	vm.InodeUsed = int64(float64(vm.InodeTotal) * (float64(rand.Intn(35)+15) / 100.0)) // 15-50%
	if vm.InodeTotal > 0 {
		vm.InodePercent = (float64(vm.InodeUsed) / float64(vm.InodeTotal)) * 100
	}
	
	// Network stats - cumulative MB since boot
	vm.NetworkRxMB = float64(rand.Intn(200000) + 5000)  // 5GB-200GB
	vm.NetworkTxMB = float64(rand.Intn(100000) + 2000)  // 2GB-100GB
	
	// Kernel version - common versions
	kernels := []string{"5.15.0-91-generic", "6.1.0-17-amd64", "5.10.0-27-arm64", "6.5.0-14-generic"}
	vm.KernelVersion = kernels[rand.Intn(len(kernels))]
}

// MonitorSwitch performs health checks on a switch
func MonitorSwitch(sw *models.Switch) {
	// When using mock data, always show as online (no network checks needed)
	if Config.Monitoring.UseMockData {
		sw.PingStatus = "online"
		sw.Status = "online"
		useSwitchMockData(sw)
		sw.LastChecked = time.Now()
		return
	}
	
	// Production mode: Check ping status using TCP connectivity
	sw.PingStatus = CheckPingStatus(sw.IPAddress)
	
	if sw.PingStatus == "online" {
		sw.Status = "online"
		
		// Use real SSH monitoring if available and configured
		if Config.SSH.Enabled {
			// Get switch-specific SSH client or use global client
			client := getSwitchSSHClient(sw)
			if client != nil {
				err := client.GetRealSwitchMetrics(sw)
				if err != nil {
					// Fallback to mock data on error
					fmt.Printf("SSH monitoring failed for %s: %v, using mock data\n", sw.Name, err)
					useSwitchMockData(sw)
				}
			} else {
				// Use mock data if no client available
				useSwitchMockData(sw)
			}
		} else {
			// Use mock data for development
			useSwitchMockData(sw)
		}
	} else {
		sw.Status = "offline"
		// Reset metrics for offline switches
		sw.Uptime = "N/A"
		sw.Processes = 0
		sw.DiskUsage = 0
		sw.DiskTotal = 50.0
		sw.DiskPercent = 0
		sw.OpenFlowStatus = "offline"
		sw.FlowCount = 0
	}
	
	sw.LastChecked = time.Now()
}

// useSwitchMockData generates mock metrics for a switch
func useSwitchMockData(sw *models.Switch) {
	sw.Uptime = GenerateUptime()
	sw.Processes = rand.Intn(80) + 20 // 20-100 processes (lighter than servers)
	sw.DiskTotal = 50.0 // Small Debian typically 8-64GB
	sw.DiskUsage = generateDiskUsage(sw.DiskTotal, 15, 40) // 15-40% usage
	sw.DiskPercent = (sw.DiskUsage / sw.DiskTotal) * 100
	
	// Memory metrics - switches have less RAM
	sw.MemoryTotal = float64(rand.Intn(3000) + 1000) // 1-4 GB in MB
	sw.MemoryUsed = generateDiskUsage(sw.MemoryTotal, 30, 60) // 30-60% usage
	sw.MemoryPercent = (sw.MemoryUsed / sw.MemoryTotal) * 100
	
	// Load average - generate realistic low values for switches
	load1 := float64(rand.Intn(100)+10) / 100.0  // 0.1-1.1
	load5 := float64(rand.Intn(90)+15) / 100.0   // 0.15-1.05
	load15 := float64(rand.Intn(80)+20) / 100.0  // 0.2-1.0
	sw.LoadAverage = fmt.Sprintf("%.2f %.2f %.2f", load1, load5, load15)
	
	// Failed services - usually 0
	if rand.Float64() < 0.05 { // 5% chance
		sw.FailedServices = 1
	} else {
		sw.FailedServices = 0
	}
	
	// Inode usage - small for switches
	sw.InodeTotal = int64(rand.Intn(500000) + 200000) // 200K-700K inodes
	sw.InodeUsed = int64(float64(sw.InodeTotal) * (float64(rand.Intn(25)+10) / 100.0)) // 10-35%
	if sw.InodeTotal > 0 {
		sw.InodePercent = (float64(sw.InodeUsed) / float64(sw.InodeTotal)) * 100
	}
	
	// Network stats - switches handle lots of traffic
	sw.NetworkRxMB = float64(rand.Intn(500000) + 50000)  // 50GB-500GB
	sw.NetworkTxMB = float64(rand.Intn(500000) + 50000)  // 50GB-500GB
	
	// Kernel version - Debian versions
	kernels := []string{"6.1.0-17-amd64", "5.10.0-27-amd64", "6.5.0-14-amd64"}
	sw.KernelVersion = kernels[rand.Intn(len(kernels))]
	
	// OpenFlow specific metrics
	sw.OpenFlowStatus = "active"
	sw.FlowCount = rand.Intn(500) + 50 // 50-550 flow rules
	sw.PortCount = rand.Intn(16) + 8   // 8-24 ports
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

// GetAllSwitches returns all configured switches
func GetAllSwitches() ([]models.Switch, error) {
	return SwitchesCache, nil
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