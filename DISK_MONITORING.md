# Disk Partition Monitoring

## Overview
The dashboard monitors disk usage across all partitions on your servers and VMs, helping you identify exactly which partition is full when disk space issues occur.

## What Gets Monitored

### Primary Partition Display
- **Root partition (`/`)** - Always monitored and displayed as the primary disk usage metric
- Shows: Total size, used space, and usage percentage
- This is the most common partition that affects system operation

### Full Partition Alerts
The system automatically scans **all mounted partitions** and reports any that exceed **90% capacity**.

Common partitions that may fill up:
- `/` - Root filesystem
- `/home` - User home directories
- `/var` - System logs, databases, application data
- `/tmp` - Temporary files
- `/opt` - Optional software installations
- `/data` - Custom data partitions
- `/boot` - Boot files (small, fills up quickly)

## How It Works

### SSH Command Execution
```bash
# Primary partition (root) - Always checked
df -BG / | tail -1 | awk '{print $3,$2,$5}'
# Output: "450G 1000G 45%"

# All partitions over 90% - Critical alerts
df -BG | awk '$5 ~ /^[0-9]+%$/ && $5+0 > 90 {print $6, $5}'
# Output example:
# /var 95%
# /tmp 92%
```

### Data Fields in API Response

```json
{
  "disk_usage": 450.0,
  "disk_total": 1000.0,
  "disk_percent": 45.0,
  "disk_partition": "/",
  "full_partitions": [
    "/var 95%",
    "/tmp 92%"
  ]
}
```

## Troubleshooting Full Partitions

### If `/var` is full:
```bash
# Find largest directories
du -sh /var/* | sort -rh | head -10

# Common culprits:
# - /var/log - Old log files
# - /var/lib/docker - Docker images/containers
# - /var/cache - Package manager cache
# - /var/tmp - Temporary files
```

### If `/home` is full:
```bash
# Find largest user directories
du -sh /home/* | sort -rh | head -10

# Check for large files
find /home -type f -size +1G -exec ls -lh {} \;
```

### If `/tmp` is full:
```bash
# Clean temporary files (safe to remove)
sudo rm -rf /tmp/*

# Or reboot (clears /tmp automatically on most systems)
```

### If `/boot` is full:
```bash
# Remove old kernels (Ubuntu/Debian)
sudo apt autoremove --purge

# List installed kernels
dpkg --list | grep linux-image
```

### If `/` (root) is full:
```bash
# Find largest directories
du -sh /* | sort -rh | head -10

# Common space-eaters:
# - Old kernels
# - Package manager cache
# - Core dumps
# - Orphaned files
```

## Production Best Practices

1. **Set up alerts** when any partition exceeds 90%
2. **Log rotation** - Ensure logs are rotated and compressed
3. **Monitoring** - Check `full_partitions` field regularly
4. **Cleanup automation** - Schedule cleanup jobs for `/tmp`, log files
5. **Disk space planning** - Monitor growth trends to predict when partitions will fill

## Development Mode

In development mode (mock data), the system simulates realistic partition usage:
- Root partition shows 30-70% usage
- No critical partition alerts (empty `full_partitions` array)

## Production Mode

Enable real partition monitoring in `config/config.yaml`:
```yaml
monitoring:
  use_mock_data: false
  ssh:
    enabled: true
    username: "monitor"
    private_key_path: "~/.ssh/id_rsa"
```

The system will then:
1. Query all partitions via SSH
2. Calculate usage for root partition
3. Identify all partitions over 90% full
4. Update `full_partitions` array with critical alerts
5. Allow you to see exactly which partition(s) need attention

## Example Scenarios

### Scenario 1: Everything Normal
```json
{
  "disk_percent": 45.0,
  "disk_partition": "/",
  "full_partitions": []
}
```
✅ All partitions healthy

### Scenario 2: Log Partition Full
```json
{
  "disk_percent": 45.0,
  "disk_partition": "/",
  "full_partitions": ["/var 95%"]
}
```
⚠️ Check `/var/log` for large log files

### Scenario 3: Multiple Partitions Critical
```json
{
  "disk_percent": 92.0,
  "disk_partition": "/",
  "full_partitions": [
    "/ 92%",
    "/var 95%",
    "/tmp 91%"
  ]
}
```
🚨 Critical - immediate action required on multiple partitions
