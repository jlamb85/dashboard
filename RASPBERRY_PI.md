# Raspberry Pi Deployment Guide

## Supported Models

### Raspberry Pi 3, 4, 5 (64-bit OS) ✅ Recommended
- **Binary**: `server-dashboard-v1.0.0-linux-arm64.tar.gz`
- **Architecture**: ARM64 (aarch64)
- **Performance**: Excellent
- **OS**: Raspberry Pi OS (64-bit), Ubuntu 64-bit

### Raspberry Pi 2, 3, 4 (32-bit OS)
- **Binary**: `server-dashboard-v1.0.0-linux-armv7.tar.gz`
- **Architecture**: ARMv7
- **Performance**: Good
- **OS**: Raspberry Pi OS (32-bit), Raspbian

### Raspberry Pi Zero, Zero W, Pi 1
- **Binary**: `server-dashboard-v1.0.0-linux-armv6.tar.gz`
- **Architecture**: ARMv6
- **Performance**: Limited (suitable for light monitoring)
- **OS**: Raspberry Pi OS (32-bit)

## Quick Install

### 1. Check Your OS Architecture

```bash
uname -m
```

Output guide:
- `aarch64` → Use **arm64** build (64-bit)
- `armv7l` → Use **armv7** build (32-bit, Pi 2/3/4)
- `armv6l` → Use **armv6** build (32-bit, Pi Zero/1)

### 2. Download Release

```bash
# For 64-bit OS (Recommended for Pi 3/4/5)
wget https://github.com/youruser/server-dashboard/releases/download/v1.0.0/server-dashboard-v1.0.0-linux-arm64.tar.gz
tar -xzf server-dashboard-v1.0.0-linux-arm64.tar.gz
cd server-dashboard-v1.0.0-linux-arm64

# OR for 32-bit OS (Pi 2/3/4)
wget https://github.com/youruser/server-dashboard/releases/download/v1.0.0/server-dashboard-v1.0.0-linux-armv7.tar.gz
tar -xzf server-dashboard-v1.0.0-linux-armv7.tar.gz
cd server-dashboard-v1.0.0-linux-armv7

# OR for Pi Zero/1
wget https://github.com/youruser/server-dashboard/releases/download/v1.0.0/server-dashboard-v1.0.0-linux-armv6.tar.gz
tar -xzf server-dashboard-v1.0.0-linux-armv6.tar.gz
cd server-dashboard-v1.0.0-linux-armv6
```

### 3. Configure

```bash
# Edit configuration
nano config/config.yaml

# Add your servers and VMs
```

### 4. Run

```bash
# Make startup script executable
chmod +x start.sh

# Start the dashboard
./start.sh
```

Access the dashboard at: `http://[raspberry-pi-ip]:8080`

## Running as a Service

To run the dashboard automatically on boot:

### Create systemd service

```bash
sudo nano /etc/systemd/system/server-dashboard.service
```

Add the following content (adjust paths):

```ini
[Unit]
Description=Server Dashboard
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/server-dashboard-v1.0.0-linux-arm64
ExecStart=/home/pi/server-dashboard-v1.0.0-linux-arm64/server-dashboard-linux-arm64
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Enable and start service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable on boot
sudo systemctl enable server-dashboard

# Start now
sudo systemctl start server-dashboard

# Check status
sudo systemctl status server-dashboard

# View logs
sudo journalctl -u server-dashboard -f
```

## Performance Optimization

### For Raspberry Pi 3/4/5

Recommended settings in `config/config.yaml`:

```yaml
monitoring_interval: 30  # Check every 30 seconds
```

These models have sufficient resources for real-time monitoring.

### For Raspberry Pi Zero/1

Conservative settings recommended:

```yaml
monitoring_interval: 60  # Check every 60 seconds (reduce load)
```

Limit the number of servers/VMs monitored:
- Pi Zero: Monitor 5-10 devices max
- Pi 1: Monitor 10-15 devices max

## Network Configuration

### Access from Other Devices

By default, the dashboard binds to `0.0.0.0:8080` (all interfaces).

Find your Raspberry Pi's IP:
```bash
hostname -I
```

Access from another device:
```
http://192.168.1.100:8080
```

### Static IP (Recommended)

Set a static IP for your Pi:

```bash
sudo nano /etc/dhcpcd.conf
```

Add:
```
interface eth0
static ip_address=192.168.1.100/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1 8.8.8.8
```

Reboot:
```bash
sudo reboot
```

## Firewall Configuration

If using `ufw`:

```bash
# Allow dashboard port
sudo ufw allow 8080/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

## Storage Requirements

Minimum SD card recommendations:
- **8GB**: Sufficient for OS + dashboard
- **16GB**: Recommended for logs and future updates
- **32GB**: Comfortable for production use

Dashboard disk usage:
- Binary: ~12-15 MB
- Configuration: < 1 MB
- Web assets: ~500 KB
- Total: ~20 MB

## SSH Monitoring from Raspberry Pi

The dashboard can monitor other servers via SSH from your Raspberry Pi.

### Generate SSH Key

```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/monitor_key
```

### Copy to Target Servers

```bash
ssh-copy-id -i ~/.ssh/monitor_key.pub user@target-server
```

### Update Configuration

Edit `config/config.yaml`:

```yaml
monitoring:
  use_mock_data: false  # Enable real monitoring
  ssh:
    enabled: true
    username: "monitor"
    private_key_path: "~/.ssh/monitor_key"
    timeout_seconds: 5
```

## Troubleshooting

### Binary Won't Run

```bash
# Check architecture
file server-dashboard-linux-arm64

# Make executable
chmod +x server-dashboard-linux-arm64

# Check for missing libraries (should be none - statically linked)
ldd server-dashboard-linux-arm64
```

### Port Already in Use

```bash
# Find what's using port 8080
sudo netstat -tulpn | grep 8080

# Or change port in config.yaml
server_address: "0.0.0.0:8081"
```

### Out of Memory (Pi Zero/1)

Reduce monitoring frequency and number of monitored devices.

Check memory:
```bash
free -h
```

### Permission Denied

```bash
# Make sure user has permissions
chmod +x server-dashboard-linux-armv6
chmod +x start.sh

# Or run with sudo (not recommended)
sudo ./start.sh
```

## Updating

1. Stop the service:
   ```bash
   sudo systemctl stop server-dashboard
   ```

2. Download new version and extract

3. Copy your configuration:
   ```bash
   cp ~/old-version/config/config.yaml ~/new-version/config/
   ```

4. Update service file if paths changed

5. Restart:
   ```bash
   sudo systemctl start server-dashboard
   ```

## Monitoring the Raspberry Pi Itself

You can have the Pi monitor itself!

Add to `config/config.yaml`:

```yaml
servers:
  - id: "local-pi"
    name: "Raspberry Pi Dashboard Host"
    ip_address: "127.0.0.1"
    hostname: "localhost"
    port: 22
    enabled: true
```

This will show:
- CPU usage
- Disk usage
- Memory usage
- Uptime
- Running processes

## Resource Usage

Typical resource consumption:

| Model | CPU Usage | RAM Usage | Notes |
|-------|-----------|-----------|-------|
| Pi 5 | < 5% | ~50 MB | Excellent performance |
| Pi 4 (4GB) | < 10% | ~50 MB | Great for production |
| Pi 3 | ~15% | ~50 MB | Good for small setups |
| Pi Zero | ~25-40% | ~30 MB | Light monitoring only |

## Questions?

For issues specific to Raspberry Pi deployment, check:
- System logs: `sudo journalctl -u server-dashboard`
- Application logs: Check stdout/stderr in terminal
- Resource usage: `top` or `htop`
