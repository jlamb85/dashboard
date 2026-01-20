# Quick Start Guide

## Development Environment

### Quick Start
```bash
# Build
go build -o server-dashboard ./cmd

# Run with defaults (localhost:8080, auth disabled)
./server-dashboard

# Run with custom configuration
export AUTH_USERNAME=admin
export AUTH_PASSWORD=mypassword
./server-dashboard

# Run on different port
export SERVER_PORT=9000
./server-dashboard
```

### View Dashboard
```
http://localhost:8080
```

### Health Check
```bash
curl http://localhost:8080/health
# Output: {"status":"healthy","timestamp":"2024-01-20T12:00:00Z"}
```

---

## Production Environment

### 1. Create Configuration File

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
# Edit .env with your production values
```

### 2. Generate TLS Certificates (Optional but Recommended)

```bash
openssl req -x509 -newkey rsa:4096 \
  -keyout server.key -out server.crt \
  -days 365 -nodes \
  -subj "/CN=dashboard.example.com"

# Or use Let's Encrypt for production
certbot certonly --standalone -d dashboard.example.com
```

### 3. Start Server with Environment Variables

```bash
# Load from .env file
export $(cat .env | xargs)

# Start server
./server-dashboard &

# Or in one command
$(cat .env | xargs) ./server-dashboard &
```

### 4. Enable HTTPS

In your `.env`:
```
TLS_ENABLED=true
TLS_CERT_FILE=/path/to/server.crt
TLS_KEY_FILE=/path/to/server.key
```

### 5. Systemd Service (Linux)

Create `/etc/systemd/system/dashboard.service`:

```ini
[Unit]
Description=Server Dashboard
After=network.target

[Service]
Type=simple
User=dashboard
WorkingDirectory=/opt/dashboard
EnvironmentFile=/opt/dashboard/.env
ExecStart=/opt/dashboard/server-dashboard
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable dashboard
sudo systemctl start dashboard
sudo journalctl -u dashboard -f
```

### 6. Docker Deployment

Build:
```bash
docker build -t dashboard:latest .
```

Run:
```bash
docker run -d \
  --name dashboard \
  -p 8080:8080 \
  -e SERVER_ADDRESS=0.0.0.0:8080 \
  -e AUTH_PASSWORD=your_password \
  -e ENVIRONMENT=production \
  dashboard:latest
```

---

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDRESS` | `0.0.0.0:8080` | Full server address |
| `SERVER_HOST` | `0.0.0.0` | Server host (override with SERVER_ADDRESS) |
| `SERVER_PORT` | `8080` | Server port (override with SERVER_ADDRESS) |
| `ENVIRONMENT` | `development` | Environment: development, staging, production |
| `AUTH_ENABLED` | `true` | Enable authentication |
| `AUTH_USERNAME` | `admin` | Dashboard username |
| `AUTH_PASSWORD` | `change_me_in_production` | **Change this!** |
| `TLS_ENABLED` | `false` | Enable HTTPS |
| `TLS_CERT_FILE` | - | Path to TLS certificate |
| `TLS_KEY_FILE` | - | Path to TLS key |
| `MONITORING_INTERVAL` | `30` | Monitoring check interval (seconds) |
| `MONITORING_TIMEOUT` | `2` | TCP timeout per port (seconds) |

### Config File (config/config.yaml)

Default values for any environment variable not set.

---

## Common Tasks

### View Logs (Systemd)
```bash
sudo journalctl -u dashboard -f
sudo journalctl -u dashboard --since "10 minutes ago"
```

### Reload Configuration (without restart)
Currently requires restart. Can be enhanced with file watcher.

### Change Password
Update in `.env` file and restart:
```bash
# Edit .env
export AUTH_PASSWORD=new_secure_password
pkill -f server-dashboard
./server-dashboard &
```

### Monitor Health
```bash
watch -n 5 'curl -s http://localhost:8080/health | jq .'
```

### Check Connectivity Issues
The dashboard monitors servers via TCP. Ensure:
1. Network connectivity between dashboard and monitored servers
2. Firewall rules allow TCP on configured ports (SSH 22, HTTP 80, HTTPS 443, MySQL 3306, PostgreSQL 5432)
3. Monitored servers are running and accepting connections

### View Metrics for a Server
```bash
curl http://localhost:8080/servers/srv001
```

---

## Security Best Practices

✅ **DO:**
- [ ] Change `AUTH_PASSWORD` immediately
- [ ] Use strong passwords (16+ characters)
- [ ] Enable TLS in production
- [ ] Run behind reverse proxy (Nginx)
- [ ] Use environment variables for secrets
- [ ] Regularly update Go
- [ ] Monitor health endpoint
- [ ] Restrict firewall access

❌ **DON'T:**
- [ ] Use default passwords in production
- [ ] Run over HTTP in production
- [ ] Commit secrets to git
- [ ] Run as root
- [ ] Expose configuration files
- [ ] Disable authentication in production

---

## Troubleshooting

### Server won't start
```bash
# Check if port is in use
lsof -i :8080

# Check configuration
cat config/config.yaml

# Try verbose logging
./server-dashboard -v 2>&1 | head -20
```

### Authentication not working
```bash
# Verify environment variables
echo $AUTH_USERNAME
echo $AUTH_PASSWORD

# Restart server
pkill -f server-dashboard
./server-dashboard &
```

### High CPU usage
- Reduce `MONITORING_INTERVAL`
- Check for stuck TCP connections
- Monitor goroutine count

### Services showing "Offline"
- Verify network connectivity
- Check firewall rules
- Verify service IP addresses in config
- Increase `MONITORING_TIMEOUT`

---

For detailed production setup, see [PRODUCTION.md](PRODUCTION.md)
