# Production Configuration Quick Reference

## One-Line Commands

### Development (no config)
```bash
./server-dashboard
```

### Production (minimum)
```bash
AUTH_PASSWORD=secure123 ./server-dashboard
```

### Production (full config)
```bash
export SERVER_ADDRESS=0.0.0.0:8443
export ENVIRONMENT=production
export AUTH_PASSWORD=strong_password_here
export TLS_ENABLED=true
export TLS_CERT_FILE=/etc/certs/server.crt
export TLS_KEY_FILE=/etc/certs/server.key
./server-dashboard
```

### Docker Quick Run
```bash
docker build -t dashboard . && \
docker run -p 8080:8080 \
  -e AUTH_PASSWORD=secure123 \
  -e ENVIRONMENT=production \
  dashboard
```

## Environment Variables (Alphabetical)

```bash
AUTH_ENABLED=true|false              # Enable authentication
AUTH_PASSWORD=value                  # CHANGE THIS!
AUTH_USERNAME=admin                  # Default: admin
ENVIRONMENT=development|staging|prod # Environment name
MONITORING_INTERVAL=30               # Check interval (seconds)
MONITORING_TIMEOUT=2                 # TCP timeout (seconds)
SERVER_ADDRESS=host:port             # Full address (e.g., 0.0.0.0:8080)
SERVER_HOST=host                     # Host only (use with SERVER_PORT)
SERVER_PORT=port                     # Port only (use with SERVER_HOST)
TLS_CERT_FILE=/path/to/cert          # SSL certificate path
TLS_ENABLED=true|false               # Enable HTTPS
TLS_KEY_FILE=/path/to/key            # SSL private key path
```

## Common Tasks

### Change Port
```bash
export SERVER_PORT=9000
./server-dashboard
```

### Enable HTTPS
```bash
# Generate certificate
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

# Run with HTTPS
export TLS_ENABLED=true
export TLS_CERT_FILE=./server.crt
export TLS_KEY_FILE=./server.key
./server-dashboard
```

### Set Production Defaults
```bash
export ENVIRONMENT=production
export AUTH_PASSWORD=secure_password_123
export SERVER_ADDRESS=0.0.0.0:8080
./server-dashboard
```

### Test Health Check
```bash
curl http://localhost:8080/health
```

### Graceful Shutdown
```bash
kill -TERM <PID>     # Graceful (default)
kill -SIGTERM <PID>  # Graceful (explicit)
```

## Configuration Priority (Highest to Lowest)

1. **Environment variables** - Override everything
2. **Config file** (config/config.yaml) - Fallback values
3. **Defaults** (hardcoded) - Last resort

Example: If `AUTH_PASSWORD` is set in both `.env` and `config.yaml`, the environment variable wins.

## Default Values

| Setting | Dev Default | Prod Recommended |
|---------|-------------|------------------|
| `SERVER_ADDRESS` | `localhost:8080` | `0.0.0.0:8080` |
| `ENVIRONMENT` | `development` | `production` |
| `AUTH_ENABLED` | `true` | `true` |
| `AUTH_PASSWORD` | `password` | `strong_unique_pass` |
| `TLS_ENABLED` | `false` | `true` |
| `MONITORING_INTERVAL` | `5s` | `30s` |

## Troubleshooting

### Server won't start?
```bash
# 1. Check syntax
go build -o server-dashboard ./cmd

# 2. Check port availability
lsof -i :8080

# 3. Check environment variables
env | grep -E "SERVER|AUTH|TLS"

# 4. Try minimal config
./server-dashboard
```

### Health check fails?
```bash
# Try different port
curl http://localhost:9000/health

# Check if server is running
ps aux | grep server-dashboard

# Check for errors on stdout
./server-dashboard 2>&1 | grep -i error
```

### HTTPS not working?
```bash
# Check cert and key exist
ls -la server.crt server.key

# Verify cert is valid
openssl x509 -in server.crt -text -noout

# Check if cert and key match
openssl x509 -modulus -noout -in server.crt | openssl md5
openssl rsa -modulus -noout -in server.key | openssl md5
# Should match
```

## Configuration Template

Create a `.env` file for easy deployment:

```bash
# Server
SERVER_ADDRESS=0.0.0.0:8080

# Environment
ENVIRONMENT=production

# Authentication (CHANGE THESE!)
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your_secure_password_here

# TLS (optional)
TLS_ENABLED=false
#TLS_CERT_FILE=/etc/certs/server.crt
#TLS_KEY_FILE=/etc/certs/server.key

# Monitoring
MONITORING_INTERVAL=30
MONITORING_TIMEOUT=2
```

Then run:
```bash
$(cat .env | xargs) ./server-dashboard
# OR
export $(cat .env | xargs)
./server-dashboard
```

## Deployment Checklist

- [ ] Build passes: `go build -o server-dashboard ./cmd`
- [ ] Password changed from default
- [ ] Environment set to `production`
- [ ] Port configured correctly
- [ ] HTTPS enabled (recommended)
- [ ] Health check tested: `curl /health`
- [ ] Graceful shutdown tested: `kill -TERM <PID>`
- [ ] Configuration backup created

## Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| "address already in use" | Change `SERVER_PORT` or kill other process |
| "certificate required" | Enable with `TLS_ENABLED=true` or disable TLS |
| "permission denied" | Run as user with permission or use `sudo` |
| "connection refused" | Check `SERVER_ADDRESS` and `SERVER_PORT` |
| "authentication failed" | Verify `AUTH_PASSWORD` is set correctly |

## Security Checklist

- [ ] Password is 16+ characters and unique
- [ ] TLS enabled with valid certificates
- [ ] Running as non-root user
- [ ] Firewall rules configured
- [ ] Reverse proxy in front (recommended)
- [ ] Health checks configured
- [ ] Secrets not in config files
- [ ] Environment variables used for all secrets

## Resources

- Full guide: [QUICKSTART.md](QUICKSTART.md)
- Detailed setup: [PRODUCTION.md](PRODUCTION.md)
- Configuration help: [PRODUCTION-CONFIG.md](PRODUCTION-CONFIG.md)
- Sample config: [.env.example](.env.example)

---

**Quick links:**
- 📖 Learn more → [QUICKSTART.md](QUICKSTART.md)
- 🔧 Configure → [.env.example](.env.example)
- 🚀 Deploy → [PRODUCTION.md](PRODUCTION.md)
