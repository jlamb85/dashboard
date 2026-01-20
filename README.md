# Server Dashboard

## Overview
The Server Dashboard is a web application designed to monitor and manage servers and virtual machines (VMs) connected to a network. It provides a modern user interface built with Bootstrap, allowing users to easily view the status of their infrastructure.

## Features
- Real-time monitoring of servers and VMs
- User-friendly dashboard with an overview of system health
- Detailed views for servers and VMs
- **Stream viewer** - View video/media streams from VMs (configurable ports)
- **Bcrypt password hashing** - Secure authentication with password hash utility
- **Enhanced logging** - Configurable file logging with automatic rotation
- **Collapsible sidebar** - Expandable/collapsible navigation with state persistence
- **Light/Dark theme** - Theme toggle with localStorage persistence
- Configurable settings for monitoring intervals and server addresses
- Authentication middleware to secure access to the dashboard
- **Embedded assets** - Fully portable single binary with embedded web files

## Project Structure
```
server-dashboard
├── cmd
│   └── main.go                # Entry point of the application
├── internal
│   ├── config
│   │   └── config.go          # Configuration loading and parsing
│   ├── handlers
│   │   ├── dashboard.go        # Dashboard request handlers
│   │   ├── server.go           # Server monitoring handlers
│   │   └── vm.go               # VM monitoring handlers
│   ├── models
│   │   ├── server.go           # Server data model
│   │   └── vm.go               # VM data model
│   ├── services
│   │   ├── monitor.go          # Monitoring services for servers and VMs
│   │   └── network.go          # Network discovery services
│   └── middleware
│       └── auth.go             # Authentication middleware
├── web
│   ├── static
│   │   ├── css
│   │   │   └── style.css       # CSS styles for the application
│   │   └── js
│   │       └── dashboard.js     # JavaScript for dynamic interactions
│   └── templates
│       ├── base.html           # Base HTML template
│       ├── dashboard.html       # Dashboard view template
│       ├── servers.html         # Servers list template
│       └── vms.html             # VMs list template
├── config
│   └── config.yaml             # Configuration settings
├── go.mod                       # Module dependencies
├── go.sum                       # Module checksums
├── Makefile                     # Build instructions
└── README.md                   # Project documentation
```

## Installation
1. Clone the repository:
   ```
   git clone <repository-url>
   cd server-dashboard
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Configure the application by editing `config/config.yaml` to set your server addresses and monitoring intervals.

## Configuration

### Development Mode (Mock Data)
By default, the dashboard uses mock data for development and testing. This allows you to work without needing SSH access to real servers.

```yaml
monitoring:
  use_mock_data: true  # Safe for development - uses generated test data
  ssh:
    enabled: false
```

### Production Mode (Real SSH Monitoring)
To enable real server monitoring via SSH in production:

1. **Enable SSH Monitoring** - Edit `config/config.yaml`:
   ```yaml
   monitoring:
     use_mock_data: false  # Use real SSH monitoring
     ssh:
       enabled: true
       username: "monitor"  # SSH username for server access
       private_key_path: "~/.ssh/id_rsa"  # Path to SSH private key
       password: ""  # Optional: use password authentication
       timeout_seconds: 5
   ```

2. **SSH Setup**:
   - Create a dedicated monitoring user on your servers
   - Generate SSH keys: `ssh-keygen -t rsa -b 4096 -f ~/.ssh/monitor_key`
   - Copy public key to servers: `ssh-copy-id -i ~/.ssh/monitor_key.pub monitor@server-ip`
   - Update `private_key_path` in config to match your key location

3. **Required Commands on Remote Servers**:
   The monitoring user needs access to these commands:
   - `df` - Disk usage statistics
   - `uptime` - System uptime
   - `ps` - Process listing
   - `awk`, `tail`, `wc` - Text processing utilities

4. **Security Notes**:
   - Use SSH key authentication (more secure than passwords)
   - Configure `HostKeyCallback` for production (currently using `InsecureIgnoreHostKey` for development)
   - Consider creating a read-only monitoring user with minimal privileges
   - Use SSH agent forwarding or encrypted key files

5. **Fallback Behavior**:
   - If SSH connection fails, the system automatically falls back to mock data
   - Connection errors are logged to help with troubleshooting
   - Dashboard remains functional even if some servers are unreachable

### Monitoring Metrics
The dashboard collects the following real-time metrics via SSH:
- **Disk Usage**: Free and total disk space on root filesystem
- **Uptime**: How long the system has been running
- **Process Count**: Number of running processes
- **Connectivity**: TCP port checks (22, 80, 443, 3306, 5432)

## Running the Application
To run the application, execute the following command:
```
go run cmd/main.go
```
The application will start a web server, and you can access the dashboard at `http://localhost:8080`.

## Building & Releases

### Quick Build
```bash
# Build for current platform
go build -o server-dashboard ./cmd/main.go

# Run the binary
./server-dashboard
```

### Multi-Platform Builds
Build release packages for Linux, macOS, Windows, and Raspberry Pi:

```bash
./build.sh
```

This creates distribution bundles in the `dist/` directory ready for deployment on:
- Linux (amd64, arm64)
- **Raspberry Pi** (armv6, armv7, arm64)
- macOS (Intel, Apple Silicon)
- Windows (64-bit)

### Version Management
```bash
# Update version (patch/minor/major)
./update_version.sh patch
```

For complete build and release documentation, see [BUILD_GUIDE.md](BUILD_GUIDE.md).

## Security

### Password Hashing Utility

The dashboard includes a password hashing utility to securely hash passwords using bcrypt.

**Location:** `tools/hashpass/`

#### Quick Usage:

**Generate and display a hash:**
```bash
cd tools/hashpass
./hashpass
```

**Update config.yaml automatically:**
```bash
./hashpass --update
```
This will prompt for a password and automatically update `config/config.yaml` with the bcrypt hash.

**Generate environment variable format:**
```bash
./hashpass --env
```

**Custom config path:**
```bash
./hashpass --config /path/to/config.yaml --update
```

#### Features:
- ✅ Secure password input (hidden while typing)
- ✅ Password confirmation
- ✅ Bcrypt hashing (industry standard, cost factor 12)
- ✅ Preserves config.yaml formatting and comments
- ✅ Environment variable output format

For complete documentation, see [tools/hashpass/README.md](tools/hashpass/README.md).

## Logging

The dashboard includes comprehensive logging with automatic rotation and configurable output.

### Configuration

In `config/config.yaml`:
```yaml
logging:
  directory: "./logs"  # Development: relative path, Production: /var/log/server-dashboard
  level: "info"  # Log level: debug, info, warn, error
  max_size_mb: 100  # Max file size before rotation
  max_backups: 5  # Number of old log files to keep
  max_age_days: 30  # Max age of log files
  compress: true  # Compress rotated logs
```

### Environment Variables
Override config with environment variables:
- `LOG_DIRECTORY` - Log file directory
- `LOG_LEVEL` - Logging level
- `LOG_MAX_SIZE` - Max size in MB before rotation
- `LOG_MAX_BACKUPS` - Number of backups to retain
- `LOG_MAX_AGE` - Max age in days

### Features
- ✅ Dual output: Console and file simultaneously
- ✅ Automatic log rotation by size
- ✅ Request logging with timing (method, path, status, duration)
- ✅ Detailed startup/shutdown logging
- ✅ Development mode: logs in `./logs/` directory
- ✅ Production mode: logs in `/var/log/server-dashboard/`
- ✅ Compressed old log files

### Log Format
```
2026/01/20 17:40:20 main.go:256: Logging initialized - writing to logs/server-dashboard.log
2026/01/20 17:40:20 main.go:60: Server Dashboard v1.0.0
2026/01/20 17:40:37 main.go:269: Started GET /health from [::1]:62918
2026/01/20 17:40:37 main.go:275: Completed GET /health with 200 in 209.916µs
```

## User Interface

### Collapsible Sidebar
The navigation sidebar can be collapsed to maximize content area:

- **Toggle button** in navbar (desktop only)
- **Smooth animations** - 0.3s transition
- **State persistence** - Preference saved to localStorage
- **Icon changes**: Sidebar inset ↔ Sidebar inset reverse
- **Full-width mode** when collapsed

### Light/Dark Theme
Switch between light and dark themes:

- **Theme toggle button** in navbar (moon/sun icon)
- **Persistent preference** - Saved to localStorage
- **System preference detection** - Respects `prefers-color-scheme`
- **Instant switching** across all pages

## Stream Viewer

The dashboard supports viewing video/media streams from VMs. Configure stream ports in `config/config.yaml`:

```yaml
virtual_machines:
  - id: "vm001"
    name: "Web Server VM"
    stream_ports: [6501, 6502]  # Multiple streams supported
  - id: "vm002"
    name: "App Server VM"
    stream_ports: [6503]  # Single stream
```

**Features:**
- Multiple streams per VM
- Auto-detection of active streams
- Collapsible accordion interface
- Embedded iframe viewer with 16:9 aspect ratio
- "Open in New Window" functionality
- Live/Offline status indicators

Streams appear in the VM detail view when configured ports are detected as active.

## Raspberry Pi Deployment

The dashboard fully supports Raspberry Pi! 🥧

Supported models:
- ✅ Raspberry Pi 5, 4, 3 (64-bit OS recommended)
- ✅ Raspberry Pi 3, 2 (32-bit OS)
- ✅ Raspberry Pi Zero, Zero W, Pi 1

For detailed Raspberry Pi installation and optimization guide, see [RASPBERRY_PI.md](RASPBERRY_PI.md).

## Documentation

- **[BUILD_GUIDE.md](BUILD_GUIDE.md)** - Complete build and release process
- **[RASPBERRY_PI.md](RASPBERRY_PI.md)** - Raspberry Pi deployment guide
- **[DISK_MONITORING.md](DISK_MONITORING.md)** - Disk partition monitoring guide
- **[config/config.yaml](config/config.yaml)** - Configuration reference

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for details.