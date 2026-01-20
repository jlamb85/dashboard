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
