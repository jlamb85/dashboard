package services

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"server-dashboard/internal/models"
)

// SSHClient wraps SSH connection for remote monitoring
type SSHClient struct {
	config *ssh.ClientConfig
	timeout time.Duration
}

// NewSSHClient creates a new SSH client for monitoring
func NewSSHClient(username, privateKeyPath, password string, timeoutSeconds int) (*SSHClient, error) {
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key verification
		Timeout:         time.Duration(timeoutSeconds) * time.Second,
	}

	// Try private key authentication first
	if privateKeyPath != "" {
		expandedPath := os.ExpandEnv(privateKeyPath)
		if strings.HasPrefix(expandedPath, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				expandedPath = strings.Replace(expandedPath, "~", home, 1)
			}
		}

		key, err := os.ReadFile(expandedPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				config.Auth = append(config.Auth, ssh.PublicKeys(signer))
			}
		}
	}

	// Fallback to password authentication
	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}

	if len(config.Auth) == 0 {
		return nil, fmt.Errorf("no authentication method configured")
	}

	return &SSHClient{
		config:  config,
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

// executeCommand runs a command via SSH and returns the output
func (c *SSHClient) executeCommand(host string, port int, command string) (string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return "", fmt.Errorf("ssh dial failed: %w", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("session creation failed: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

// GetRealServerMetrics queries real server metrics via SSH
func (c *SSHClient) GetRealServerMetrics(srv *models.Server) error {
	// Get disk usage for root partition: df -BG / | tail -1 | awk '{print $3,$2,$5}'
	diskCmd := "df -BG / | tail -1 | awk '{print $3,$2,$5}'"
	diskOutput, err := c.executeCommand(srv.IPAddress, srv.Port, diskCmd)
	if err != nil {
		return fmt.Errorf("disk query failed: %w", err)
	}

	// Parse root partition: "450G 1000G 45%"
	var usedStr, totalStr, percentStr string
	fmt.Sscanf(strings.TrimSpace(diskOutput), "%s %s %s", &usedStr, &totalStr, &percentStr)
	
	srv.DiskUsage = parseGigabytes(usedStr)
	srv.DiskTotal = parseGigabytes(totalStr)
	srv.DiskPercent = parsePercent(percentStr)
	srv.DiskPartition = "/"

	// Check all partitions for critical disk usage (>90%)
	srv.FullPartitions = c.checkFullPartitions(srv.IPAddress, srv.Port)

	// Get uptime
	uptimeCmd := "uptime -p 2>/dev/null || uptime | awk '{print $3, $4}'"
	uptimeOutput, err := c.executeCommand(srv.IPAddress, srv.Port, uptimeCmd)
	if err == nil {
		srv.Uptime = strings.TrimSpace(uptimeOutput)
		if srv.Uptime == "" {
			srv.Uptime = "N/A"
		}
	}

	// Get process count
	processCmd := "ps aux | wc -l"
	processOutput, err := c.executeCommand(srv.IPAddress, srv.Port, processCmd)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(processOutput))
		srv.Processes = count
	}

	// Get memory usage: free -m | grep Mem | awk '{print $3,$2}'
	memoryCmd := "free -m | grep Mem | awk '{print $3,$2}'"
	memoryOutput, err := c.executeCommand(srv.IPAddress, srv.Port, memoryCmd)
	if err == nil {
		var used, total int
		fmt.Sscanf(strings.TrimSpace(memoryOutput), "%d %d", &used, &total)
		srv.MemoryUsed = float64(used)
		srv.MemoryTotal = float64(total)
		if total > 0 {
			srv.MemoryPercent = (float64(used) / float64(total)) * 100
		}
	}

	// Get load average: cat /proc/loadavg | awk '{print $1, $2, $3}'
	loadCmd := "cat /proc/loadavg | awk '{print $1, $2, $3}'"
	loadOutput, err := c.executeCommand(srv.IPAddress, srv.Port, loadCmd)
	if err == nil {
		srv.LoadAverage = strings.TrimSpace(loadOutput)
	}

	// Get failed systemd services count: systemctl --failed --no-pager | grep -c 'loaded.*failed'
	failedCmd := "systemctl --failed --no-pager 2>/dev/null | grep -c 'loaded.*failed' || echo 0"
	failedOutput, err := c.executeCommand(srv.IPAddress, srv.Port, failedCmd)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(failedOutput))
		srv.FailedServices = count
	}

	// Get inode usage: df -i / | tail -1 | awk '{print $3,$2,$5}'
	inodeCmd := "df -i / | tail -1 | awk '{print $3,$2,$5}'"
	inodeOutput, err := c.executeCommand(srv.IPAddress, srv.Port, inodeCmd)
	if err == nil {
		var used, total int64
		var percentStr string
		fmt.Sscanf(strings.TrimSpace(inodeOutput), "%d %d %s", &used, &total, &percentStr)
		srv.InodeUsed = used
		srv.InodeTotal = total
		srv.InodePercent = parsePercent(percentStr)
	}

	// Get network stats: cat /proc/net/dev | grep -E '^ *(eth0|enp|ens|wlan0):' | awk '{rx+=$2; tx+=$10} END {print rx/1048576, tx/1048576}'
	networkCmd := "cat /proc/net/dev | grep -v 'lo:' | grep ':' | awk '{rx+=$2; tx+=$10} END {printf \"%.2f %.2f\", rx/1048576, tx/1048576}'"
	networkOutput, err := c.executeCommand(srv.IPAddress, srv.Port, networkCmd)
	if err == nil {
		var rx, tx float64
		fmt.Sscanf(strings.TrimSpace(networkOutput), "%f %f", &rx, &tx)
		srv.NetworkRxMB = rx
		srv.NetworkTxMB = tx
	}

	// Get kernel version: uname -r
	kernelCmd := "uname -r"
	kernelOutput, err := c.executeCommand(srv.IPAddress, srv.Port, kernelCmd)
	if err == nil {
		srv.KernelVersion = strings.TrimSpace(kernelOutput)
	}

	return nil
}

// GetRealVMMetrics queries real VM metrics via SSH
func (c *SSHClient) GetRealVMMetrics(vm *models.VM) error {
	// Get disk usage for root partition
	diskCmd := "df -BG / | tail -1 | awk '{print $3,$2,$5}'"
	diskOutput, err := c.executeCommand(vm.IPAddress, vm.Port, diskCmd)
	if err != nil {
		return fmt.Errorf("disk query failed: %w", err)
	}

	var usedStr, totalStr, percentStr string
	fmt.Sscanf(strings.TrimSpace(diskOutput), "%s %s %s", &usedStr, &totalStr, &percentStr)
	
	vm.DiskUsage = parseGigabytes(usedStr)
	vm.DiskTotal = parseGigabytes(totalStr)
	vm.DiskPercent = parsePercent(percentStr)
	vm.DiskPartition = "/"

	// Check all partitions for critical disk usage (>90%)
	vm.FullPartitions = c.checkFullPartitions(vm.IPAddress, vm.Port)

	// Get uptime
	uptimeCmd := "uptime -p 2>/dev/null || uptime | awk '{print $3, $4}'"
	uptimeOutput, err := c.executeCommand(vm.IPAddress, vm.Port, uptimeCmd)
	if err == nil {
		vm.Uptime = strings.TrimSpace(uptimeOutput)
		if vm.Uptime == "" {
			vm.Uptime = "N/A"
		}
	}

	// Get process count
	processCmd := "ps aux | wc -l"
	processOutput, err := c.executeCommand(vm.IPAddress, vm.Port, processCmd)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(processOutput))
		vm.Processes = count
	}

	// Get memory usage: free -m | grep Mem | awk '{print $3,$2}'
	memoryCmd := "free -m | grep Mem | awk '{print $3,$2}'"
	memoryOutput, err := c.executeCommand(vm.IPAddress, vm.Port, memoryCmd)
	if err == nil {
		var used, total int
		fmt.Sscanf(strings.TrimSpace(memoryOutput), "%d %d", &used, &total)
		vm.MemoryUsed = float64(used)
		vm.MemoryTotal = float64(total)
		if total > 0 {
			vm.MemoryPercent = (float64(used) / float64(total)) * 100
		}
	}

	// Get load average: cat /proc/loadavg | awk '{print $1, $2, $3}'
	loadCmd := "cat /proc/loadavg | awk '{print $1, $2, $3}'"
	loadOutput, err := c.executeCommand(vm.IPAddress, vm.Port, loadCmd)
	if err == nil {
		vm.LoadAverage = strings.TrimSpace(loadOutput)
	}

	// Get failed systemd services count: systemctl --failed --no-pager | grep -c 'loaded.*failed'
	failedCmd := "systemctl --failed --no-pager 2>/dev/null | grep -c 'loaded.*failed' || echo 0"
	failedOutput, err := c.executeCommand(vm.IPAddress, vm.Port, failedCmd)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(failedOutput))
		vm.FailedServices = count
	}

	// Get inode usage: df -i / | tail -1 | awk '{print $3,$2,$5}'
	inodeCmd := "df -i / | tail -1 | awk '{print $3,$2,$5}'"
	inodeOutput, err := c.executeCommand(vm.IPAddress, vm.Port, inodeCmd)
	if err == nil {
		var used, total int64
		var percentStr string
		fmt.Sscanf(strings.TrimSpace(inodeOutput), "%d %d %s", &used, &total, &percentStr)
		vm.InodeUsed = used
		vm.InodeTotal = total
		vm.InodePercent = parsePercent(percentStr)
	}

	// Get network stats: cat /proc/net/dev | grep -v 'lo:' | grep ':' | awk '{rx+=$2; tx+=$10} END {printf "%.2f %.2f", rx/1048576, tx/1048576}'
	networkCmd := "cat /proc/net/dev | grep -v 'lo:' | grep ':' | awk '{rx+=$2; tx+=$10} END {printf \"%.2f %.2f\", rx/1048576, tx/1048576}'"
	networkOutput, err := c.executeCommand(vm.IPAddress, vm.Port, networkCmd)
	if err == nil {
		var rx, tx float64
		fmt.Sscanf(strings.TrimSpace(networkOutput), "%f %f", &rx, &tx)
		vm.NetworkRxMB = rx
		vm.NetworkTxMB = tx
	}

	// Get kernel version: uname -r
	kernelCmd := "uname -r"
	kernelOutput, err := c.executeCommand(vm.IPAddress, vm.Port, kernelCmd)
	if err == nil {
		vm.KernelVersion = strings.TrimSpace(kernelOutput)
	}

	return nil
}

// parseGigabytes converts "450G" to 450.0
func parseGigabytes(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "G")
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// parsePercent converts "45%" to 45.0
func parsePercent(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// checkFullPartitions identifies all partitions over 90% full
func (c *SSHClient) checkFullPartitions(host string, port int) []string {
	// Get all filesystem info: df -BG | awk '$5 ~ /^[0-9]+%$/ && $5+0 > 90 {print $6, $5}'
	// This finds partitions over 90% and returns: "/var 95%"
	fullCmd := "df -BG | awk '$5 ~ /^[0-9]+%$/ && $5+0 > 90 {print $6, $5}'"
	output, err := c.executeCommand(host, port, fullCmd)
	if err != nil {
		return []string{} // Return empty on error
	}

	var fullPartitions []string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fullPartitions = append(fullPartitions, line)
		}
	}

	return fullPartitions
}
// GetRealSwitchMetrics queries real switch metrics via SSH
func (c *SSHClient) GetRealSwitchMetrics(sw *models.Switch) error {
	// Get disk usage for root partition
	diskCmd := "df -BG / | tail -1 | awk '{print $3,$2,$5}'"
	diskOutput, err := c.executeCommand(sw.IPAddress, sw.Port, diskCmd)
	if err != nil {
		return fmt.Errorf("disk query failed: %w", err)
	}

	var usedStr, totalStr, percentStr string
	fmt.Sscanf(strings.TrimSpace(diskOutput), "%s %s %s", &usedStr, &totalStr, &percentStr)
	
	sw.DiskUsage = parseGigabytes(usedStr)
	sw.DiskTotal = parseGigabytes(totalStr)
	sw.DiskPercent = parsePercent(percentStr)
	sw.DiskPartition = "/"

	// Check all partitions for critical disk usage (>90%)
	sw.FullPartitions = c.checkFullPartitions(sw.IPAddress, sw.Port)

	// Get uptime
	uptimeCmd := "uptime -p 2>/dev/null || uptime | awk '{print $3, $4}'"
	uptimeOutput, err := c.executeCommand(sw.IPAddress, sw.Port, uptimeCmd)
	if err == nil {
		sw.Uptime = strings.TrimSpace(uptimeOutput)
		if sw.Uptime == "" {
			sw.Uptime = "N/A"
		}
	}

	// Get process count
	processCmd := "ps aux | wc -l"
	processOutput, err := c.executeCommand(sw.IPAddress, sw.Port, processCmd)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(processOutput))
		sw.Processes = count
	}

	// Get memory usage
	memoryCmd := "free -m | grep Mem | awk '{print $3,$2}'"
	memoryOutput, err := c.executeCommand(sw.IPAddress, sw.Port, memoryCmd)
	if err == nil {
		var used, total int
		fmt.Sscanf(strings.TrimSpace(memoryOutput), "%d %d", &used, &total)
		sw.MemoryUsed = float64(used)
		sw.MemoryTotal = float64(total)
		if total > 0 {
			sw.MemoryPercent = (float64(used) / float64(total)) * 100
		}
	}

	// Get load average
	loadCmd := "cat /proc/loadavg | awk '{print $1, $2, $3}'"
	loadOutput, err := c.executeCommand(sw.IPAddress, sw.Port, loadCmd)
	if err == nil {
		sw.LoadAverage = strings.TrimSpace(loadOutput)
	}

	// Get failed systemd services count
	failedCmd := "systemctl --failed --no-pager 2>/dev/null | grep -c 'loaded.*failed' || echo 0"
	failedOutput, err := c.executeCommand(sw.IPAddress, sw.Port, failedCmd)
	if err == nil {
		count, _ := strconv.Atoi(strings.TrimSpace(failedOutput))
		sw.FailedServices = count
	}

	// Get inode usage
	inodeCmd := "df -i / | tail -1 | awk '{print $3,$2,$5}'"
	inodeOutput, err := c.executeCommand(sw.IPAddress, sw.Port, inodeCmd)
	if err == nil {
		var used, total int64
		var percentStr string
		fmt.Sscanf(strings.TrimSpace(inodeOutput), "%d %d %s", &used, &total, &percentStr)
		sw.InodeUsed = used
		sw.InodeTotal = total
		sw.InodePercent = parsePercent(percentStr)
	}

	// Get network stats
	networkCmd := "cat /proc/net/dev | grep -v 'lo:' | grep ':' | awk '{rx+=$2; tx+=$10} END {printf \"%.2f %.2f\", rx/1048576, tx/1048576}'"
	networkOutput, err := c.executeCommand(sw.IPAddress, sw.Port, networkCmd)
	if err == nil {
		var rx, tx float64
		fmt.Sscanf(strings.TrimSpace(networkOutput), "%f %f", &rx, &tx)
		sw.NetworkRxMB = rx
		sw.NetworkTxMB = tx
	}

	// Get kernel version
	kernelCmd := "uname -r"
	kernelOutput, err := c.executeCommand(sw.IPAddress, sw.Port, kernelCmd)
	if err == nil {
		sw.KernelVersion = strings.TrimSpace(kernelOutput)
	}

	// Get OpenFlow status (check if ovs-vsctl is available)
	ovsStatusCmd := "command -v ovs-vsctl >/dev/null 2>&1 && echo 'installed' || echo 'not_installed'"
	ovsStatusOutput, err := c.executeCommand(sw.IPAddress, sw.Port, ovsStatusCmd)
	if err == nil && strings.Contains(strings.TrimSpace(ovsStatusOutput), "installed") {
		// Get OpenFlow controller connection status
		controllerCmd := "ovs-vsctl get-controller br0 2>/dev/null || echo 'N/A'"
		controllerOutput, err := c.executeCommand(sw.IPAddress, sw.Port, controllerCmd)
		if err == nil {
			controller := strings.TrimSpace(controllerOutput)
			if controller != "" && controller != "N/A" {
				sw.OpenFlowStatus = "active"
				// Extract IP from tcp:192.168.1.250:6653 format
				if strings.Contains(controller, ":") {
					parts := strings.Split(controller, ":")
					if len(parts) >= 2 {
						sw.ControllerIP = parts[1]
					}
				}
			} else {
				sw.OpenFlowStatus = "inactive"
			}
		}

		// Get flow count
		flowCmd := "ovs-ofctl dump-flows br0 2>/dev/null | grep -c 'cookie=' || echo 0"
		flowOutput, err := c.executeCommand(sw.IPAddress, sw.Port, flowCmd)
		if err == nil {
			count, _ := strconv.Atoi(strings.TrimSpace(flowOutput))
			sw.FlowCount = count
		}

		// Get port count
		portCmd := "ovs-ofctl show br0 2>/dev/null | grep -c ' addr:' || echo 0"
		portOutput, err := c.executeCommand(sw.IPAddress, sw.Port, portCmd)
		if err == nil {
			count, _ := strconv.Atoi(strings.TrimSpace(portOutput))
			sw.PortCount = count
		}

		// Get OpenFlow version
		versionCmd := "ovs-ofctl -V 2>/dev/null | head -1 | awk '{print $NF}' || echo 'N/A'"
		versionOutput, err := c.executeCommand(sw.IPAddress, sw.Port, versionCmd)
		if err == nil {
			sw.OpenFlowVersion = strings.TrimSpace(versionOutput)
		}
	} else {
		sw.OpenFlowStatus = "not_installed"
	}

	return nil
}