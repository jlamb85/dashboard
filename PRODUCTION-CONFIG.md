# Production Configuration - What You Need to Know

## 🎯 Key Changes Made

Your dashboard now has complete **production-ready configuration support**:

### 1. Environment Variables
All critical settings can be configured via environment variables - **no code changes needed** for different environments:

```bash
# Development
./server-dashboard  # Uses defaults: localhost:8080

# Production
export SERVER_ADDRESS=0.0.0.0:443
export ENVIRONMENT=production
export AUTH_PASSWORD=strong_password
export TLS_ENABLED=true
./server-dashboard
```

### 2. TLS/HTTPS Support
Enable encrypted HTTPS without modifying any code:

```bash
export TLS_ENABLED=true
export TLS_CERT_FILE=/etc/certs/server.crt
export TLS_KEY_FILE=/etc/certs/server.key
```

### 3. Health Check Endpoint
Built-in `/health` endpoint for load balancers and orchestrators:

```bash
curl http://localhost:8080/health
# {"status":"healthy","timestamp":"2024-01-20T12:00:00Z"}
```

### 4. Graceful Shutdown
Handles SIGTERM properly - completes requests before exiting:

```bash
./server-dashboard &
kill -TERM $!  # Clean shutdown
```

### 5. Security Headers
Automatic security headers on all HTTP responses

## 📄 Documentation Files

| File | Purpose |
|------|---------|
| [QUICKSTART.md](QUICKSTART.md) | **Start here** - Quick setup and deployment examples |
| [PRODUCTION.md](PRODUCTION.md) | Comprehensive guide - Systemd, Docker, K8s, security, monitoring |
| [.env.example](.env.example) | Copy this to `.env` and customize for your environment |

## 🚀 30-Second Start Guide

### Development
```bash
./server-dashboard
# http://localhost:8080
```

### Production with One Command
```bash
export SERVER_ADDRESS=0.0.0.0:8080
export ENVIRONMENT=production
export AUTH_PASSWORD=your_password_here
./server-dashboard
```

### Docker
```bash
docker build -t dashboard .
docker run -p 8080:8080 \
  -e AUTH_PASSWORD=secure123 \
  -e ENVIRONMENT=production \
  dashboard
```

### Systemd (Linux)
```bash
# Set up .env file
cp .env.example /etc/dashboard/.env
# Edit /etc/dashboard/.env with your settings

# Copy service file
sudo cp dashboard.service /etc/systemd/system/

# Start
sudo systemctl enable dashboard
sudo systemctl start dashboard
sudo journalctl -u dashboard -f  # View logs
```

## 🔑 Environment Variables Quick Reference

### Essential
```bash
SERVER_ADDRESS=0.0.0.0:8080          # Listen address
ENVIRONMENT=production                # Environment type
AUTH_PASSWORD=your_password           # CHANGE THIS!
```

### Optional
```bash
AUTH_USERNAME=admin                   # Default: admin
AUTH_ENABLED=true                     # Default: true
TLS_ENABLED=false                     # Default: false
TLS_CERT_FILE=/path/to/cert.crt      # Only if TLS_ENABLED=true
TLS_KEY_FILE=/path/to/cert.key       # Only if TLS_ENABLED=true
MONITORING_INTERVAL=30                # Seconds between checks
MONITORING_TIMEOUT=2                  # TCP timeout in seconds
```

## ✅ Minimal Production Setup

1. **Build**
   ```bash
   go build -o server-dashboard ./cmd
   ```

2. **Configure**
   ```bash
   cp .env.example .env
   # Edit .env - change at minimum:
   #   AUTH_PASSWORD=something_secure
   #   ENVIRONMENT=production
   ```

3. **Run**
   ```bash
   $(cat .env | xargs) ./server-dashboard
   ```

4. **Verify**
   ```bash
   curl http://localhost:8080/health
   # Should return: {"status":"healthy",...}
   ```

## 🔒 Security Requirements

**Before deploying to production:**

- [ ] Change `AUTH_PASSWORD` from default
- [ ] Set `ENVIRONMENT=production`
- [ ] Use strong passwords (16+ characters)
- [ ] Enable TLS with valid certificates
- [ ] Run as non-root user
- [ ] Use environment variables for all secrets
- [ ] Don't commit `.env` to git (add to `.gitignore`)

## 📊 Architecture

```
Environment Variables (highest priority)
          ↓
Config File (config/config.yaml)
          ↓
Defaults (hardcoded)
```

This means:
- Config file sets defaults
- Environment variables override config file
- No need to modify files for different environments

## 🧪 Testing

### Quick Test
```bash
# Start
./server-dashboard &
PID=$!
sleep 2

# Test health
curl http://localhost:8080/health

# Test dashboard
curl -s http://localhost:8080/ | head -20

# Clean shutdown
kill -TERM $PID
```

### Test with Custom Configuration
```bash
# Custom port and password
SERVER_PORT=9000 \
AUTH_PASSWORD=testpass123 \
ENVIRONMENT=test \
./server-dashboard &

sleep 2
curl http://localhost:9000/health
pkill -f server-dashboard
```

## 🐛 Troubleshooting

### Port already in use
```bash
lsof -i :8080
# Kill the process using the port
kill -9 <PID>
```

### Server won't start with custom config
```bash
# Verify environment variables are set
env | grep -E "SERVER|AUTH|TLS|MONITORING"

# Run with explicit values
SERVER_ADDRESS=0.0.0.0:8080 \
AUTH_PASSWORD=test123 \
./server-dashboard
```

### Health check fails
```bash
# Check if server is running
curl -v http://localhost:8080/health

# Check logs
./server-dashboard 2>&1 | grep -i error
```

## 📚 For More Information

- **Quick setup**: Read [QUICKSTART.md](QUICKSTART.md)
- **Production deployment**: Read [PRODUCTION.md](PRODUCTION.md)
- **All options**: See [.env.example](.env.example)

## 🎉 You're Ready!

Your dashboard is **production-ready**. Start with the QUICKSTART.md guide and deploy with confidence!

---

**Next Step:** Read [QUICKSTART.md](QUICKSTART.md) for your specific deployment scenario (Docker, Systemd, Kubernetes, etc.)
