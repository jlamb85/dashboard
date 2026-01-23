package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddress      string              `yaml:"server_address"`
	MonitoringInterval int                 `yaml:"monitoring_interval"`
	Logging            LoggingConfig       `yaml:"logging"`
	Auth               AuthConfig          `yaml:"auth"`
	Servers            []ServerConfig      `yaml:"servers"`
	VirtualMachines    []VirtualMachineConfig `yaml:"virtual_machines"`
	Switches           []SwitchConfig      `yaml:"switches"`
	Monitoring         MonitoringConfig    `yaml:"monitoring"`
	SyntheticChecks    []SyntheticCheckConfig `yaml:"synthetic_checks"`
	SSH                SSHConfig           `yaml:"ssh"`
	TLS                TLSConfig           `yaml:"tls"`
	UI                 UIConfig            `yaml:"ui"`
	Environment        string              `yaml:"environment"`
}

type LoggingConfig struct {
	Directory   string `yaml:"directory"`
	Level       string `yaml:"level"`
	MaxSizeMB   int    `yaml:"max_size_mb"`
	MaxBackups  int    `yaml:"max_backups"`
	MaxAgeDays  int    `yaml:"max_age_days"`
	Compress    bool   `yaml:"compress"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Enabled  bool   `yaml:"enabled"`
}

type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type ServerConfig struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	IPAddress string `yaml:"ip_address"`
	Hostname  string `yaml:"hostname"`
	Port      int    `yaml:"port"`
	Enabled   bool   `yaml:"enabled"`
	Tags      []string `yaml:"tags"`
}

type VirtualMachineConfig struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	IPAddress    string `yaml:"ip_address"`
	Hostname     string `yaml:"hostname"`
	Port         int    `yaml:"port"`
	Enabled      bool   `yaml:"enabled"`
	HostServerID string `yaml:"host_server_id"`
	StreamPorts  []int  `yaml:"stream_ports"` // Optional ports for video/media streaming
	Tags         []string `yaml:"tags"`
}

type SwitchConfig struct {
	ID              string `yaml:"id"`
	Name            string `yaml:"name"`
	IPAddress       string `yaml:"ip_address"`
	Hostname        string `yaml:"hostname"`
	Port            int    `yaml:"port"`
	Enabled         bool   `yaml:"enabled"`
	ControllerIP    string `yaml:"controller_ip"`    // SDN controller IP
	OpenFlowVersion string `yaml:"openflow_version"` // Expected OpenFlow version
	// SSH credentials specific to this switch (optional, falls back to global SSH config)
	SSHUsername     string `yaml:"ssh_username"`      // Switch-specific SSH username
	SSHPassword     string `yaml:"ssh_password"`      // Switch-specific SSH password
	SSHKeyPath      string `yaml:"ssh_key_path"`      // Switch-specific SSH private key path
	Tags            []string `yaml:"tags"`
}

type SyntheticCheckConfig struct {
	ID              string   `yaml:"id"`
	Name            string   `yaml:"name"`
	Type            string   `yaml:"type"`             // http, tcp, dns
	URL             string   `yaml:"url"`              // for http
	Host            string   `yaml:"host"`             // for tcp/dns
	Port            int      `yaml:"port"`             // for tcp
	ExpectedStatus  int      `yaml:"expected_status"`  // for http
	IntervalSeconds int      `yaml:"interval_seconds"`
	TimeoutSeconds  int      `yaml:"timeout_seconds"`
	Enabled         bool     `yaml:"enabled"`
	Tags            []string `yaml:"tags"`
}

type MonitoringConfig struct {
	PingTimeoutSeconds  int  `yaml:"ping_timeout_seconds"`
	DiskThresholdPercent int `yaml:"disk_threshold_percent"`
	CheckProcesses      bool `yaml:"check_processes"`
	CheckDiskSpace      bool `yaml:"check_disk_space"`
	CheckUptime         bool `yaml:"check_uptime"`
	UseMockData         bool `yaml:"use_mock_data"`
}

type UIConfig struct {
	ShowQuickSummary       bool `yaml:"show_quick_summary"`
	ShowMonitoringFeatures bool `yaml:"show_monitoring_features"`
	ShowNavigationButtons  bool `yaml:"show_navigation_buttons"`
	ShowSynthetics         bool `yaml:"show_synthetics"`
	EnableAutoRefresh      bool `yaml:"enable_auto_refresh"`
	AutoRefreshSeconds     int  `yaml:"auto_refresh_seconds"`
}

type SSHConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Username       string `yaml:"username"`
	PrivateKeyPath string `yaml:"private_key_path"`
	Password       string `yaml:"password"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	return &config, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(cfg *Config) {
	// Server configuration
	if host := os.Getenv("SERVER_HOST"); host != "" {
		if port := os.Getenv("SERVER_PORT"); port != "" {
			cfg.ServerAddress = host + ":" + port
		}
	} else if port := os.Getenv("SERVER_PORT"); port != "" {
		// If only port is set, keep current host
		parts := strings.Split(cfg.ServerAddress, ":")
		if len(parts) > 0 {
			cfg.ServerAddress = parts[0] + ":" + port
		}
	}

	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		cfg.ServerAddress = addr
	}

	// Monitoring configuration
	if interval := os.Getenv("MONITORING_INTERVAL"); interval != "" {
		if i, err := strconv.Atoi(interval); err == nil {
			cfg.MonitoringInterval = i
		}
	}

	// Environment
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		cfg.Environment = env
	} else if env := os.Getenv("APP_ENV"); env != "" {
		cfg.Environment = env
	}

	// Authentication
	if user := os.Getenv("AUTH_USERNAME"); user != "" {
		cfg.Auth.Username = user
	}
	if pass := os.Getenv("AUTH_PASSWORD"); pass != "" {
		cfg.Auth.Password = pass
	}
	if authEnabled := os.Getenv("AUTH_ENABLED"); authEnabled != "" {
		cfg.Auth.Enabled = strings.ToLower(authEnabled) == "true"
	}

	// TLS configuration
	if tlsEnabled := os.Getenv("TLS_ENABLED"); tlsEnabled != "" {
		cfg.TLS.Enabled = strings.ToLower(tlsEnabled) == "true"
	}
	if certFile := os.Getenv("TLS_CERT_FILE"); certFile != "" {
		cfg.TLS.CertFile = certFile
	}
	if keyFile := os.Getenv("TLS_KEY_FILE"); keyFile != "" {
		cfg.TLS.KeyFile = keyFile
	}

	// Monitoring options
	if timeout := os.Getenv("MONITORING_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			cfg.Monitoring.PingTimeoutSeconds = t
		}
	}

	// Logging configuration
	if logDir := os.Getenv("LOG_DIRECTORY"); logDir != "" {
		cfg.Logging.Directory = logDir
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if maxSize := os.Getenv("LOG_MAX_SIZE"); maxSize != "" {
		if size, err := strconv.Atoi(maxSize); err == nil {
			cfg.Logging.MaxSizeMB = size
		}
	}
	if maxBackups := os.Getenv("LOG_MAX_BACKUPS"); maxBackups != "" {
		if backups, err := strconv.Atoi(maxBackups); err == nil {
			cfg.Logging.MaxBackups = backups
		}
	}
	if maxAge := os.Getenv("LOG_MAX_AGE"); maxAge != "" {
		if age, err := strconv.Atoi(maxAge); err == nil {
			cfg.Logging.MaxAgeDays = age
		}
	}
}