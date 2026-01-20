# Production Configuration Guide

## Environment Variables

The dashboard supports configuration via environment variables, allowing you to deploy the same binary across different environments without changing code.

### Basic Configuration

```bash
# Server binding (default: 0.0.0.0:8080)
SERVER_ADDRESS=0.0.0.0:8443
# OR separate host and port:
SERVER_HOST=0.0.0.0
SERVER_PORT=8443

# Environment (default: development)
ENVIRONMENT=production

# Monitoring interval in seconds (default: 30)
MONITORING_INTERVAL=60
MONITORING_TIMEOUT=2
```

### Security Configuration

**⚠️ CRITICAL: Change these in production!**

```bash
# Authentication
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your_very_secure_password_123
```

### TLS/HTTPS Configuration

For production, enable HTTPS to encrypt traffic:

```bash
# TLS Settings
TLS_ENABLED=true
TLS_CERT_FILE=/etc/dashboard/certs/server.crt
TLS_KEY_FILE=/etc/dashboard/certs/server.key
```

**Generate self-signed certificate for testing:**

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes
```

## Deployment Methods

### 1. Systemd Service (Linux)

Create `/etc/systemd/system/dashboard.service`:

```ini
[Unit]
Description=Server Dashboard
After=network.target

[Service]
Type=simple
User=dashboard
WorkingDirectory=/opt/dashboard
EnvironmentFile=/etc/dashboard/.env
ExecStart=/opt/dashboard/server-dashboard
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl daemon-reload
sudo systemctl enable dashboard
sudo systemctl start dashboard
sudo journalctl -u dashboard -f     # View logs
```

### 2. Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server-dashboard ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server-dashboard .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./server-dashboard"]
```

Build and run:

```bash
docker build -t dashboard:latest .
docker run -p 8080:8080 \
  -e SERVER_ADDRESS=0.0.0.0:8080 \
  -e AUTH_PASSWORD=your_password \
  -e TLS_ENABLED=false \
  dashboard:latest
```

### 3. Kubernetes Deployment

Create `k8s-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dashboard
spec:
  replicas: 2
  selector:
    matchLabels:
      app: dashboard
  template:
    metadata:
      labels:
        app: dashboard
    spec:
      containers:
      - name: dashboard
        image: dashboard:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVER_ADDRESS
          value: "0.0.0.0:8080"
        - name: ENVIRONMENT
          value: "production"
        - name: AUTH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: dashboard-secrets
              key: auth-password
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: dashboard-service
spec:
  selector:
    app: dashboard
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

Deploy:

```bash
# Create secret
kubectl create secret generic dashboard-secrets \
  --from-literal=auth-password='your_secure_password'

# Deploy
kubectl apply -f k8s-deployment.yaml

# View logs
kubectl logs -f deployment/dashboard
```

## Security Checklist

- [ ] Change `AUTH_PASSWORD` from default
- [ ] Enable `TLS_ENABLED=true` with valid certificates
- [ ] Use strong passwords (minimum 16 characters)
- [ ] Run as non-root user
- [ ] Enable firewall rules (only allow required ports)
- [ ] Use environment variables, not config files, for secrets
- [ ] Regularly update Go and dependencies
- [ ] Monitor health endpoint: `curl http://localhost:8080/health`
- [ ] Enable access logs in production
- [ ] Run behind reverse proxy (Nginx, HAProxy) for additional security

## Health Check

The dashboard exposes a health check endpoint:

```bash
curl http://localhost:8080/health
# Response: {"status":"healthy","timestamp":"2024-01-20T12:00:00Z"}
```

Use this for load balancer health checks and orchestration platforms.

## Reverse Proxy Configuration (Nginx)

For additional security and SSL termination:

```nginx
upstream dashboard {
    server localhost:8080;
    server localhost:8081;
}

server {
    listen 443 ssl http2;
    server_name dashboard.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://dashboard;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        access_log off;
        proxy_pass http://dashboard;
    }
}

server {
    listen 80;
    server_name dashboard.example.com;
    return 301 https://$server_name$request_uri;
}
```

## Monitoring & Alerts

Monitor these metrics:

1. **Health endpoint response time**: Should be <10ms
2. **Authentication failures**: Check logs for brute force attempts
3. **Memory usage**: Monitor for leaks in monitoring loop
4. **Disk space**: For logging and config storage
5. **Network connectivity**: Monitor TCP port checks success rate

Example Prometheus metrics configuration (future enhancement):

```yaml
- job_name: 'dashboard'
  static_configs:
    - targets: ['localhost:8080']
  metrics_path: '/metrics'
```

## Logs

The application uses standard Go logging. For production:

1. **Redirect to file:**
   ```bash
   ./server-dashboard >> /var/log/dashboard/app.log 2>&1 &
   ```

2. **Use journalctl (with systemd):**
   ```bash
   journalctl -u dashboard -f
   ```

3. **Log rotation (add to systemd service):**
   ```ini
   StandardOutput=journal
   StandardError=journal
   ```

## Troubleshooting

### Port already in use
```bash
lsof -i :8080          # Find process using port
kill -9 <PID>          # Kill it
```

### Authentication not working
```bash
# Check env vars are set
echo $AUTH_USERNAME
echo $AUTH_PASSWORD

# Verify config loading
./server-dashboard -v  # (if you add verbose flag)
```

### TLS certificate issues
```bash
# Verify certificate
openssl x509 -in server.crt -text -noout

# Check if cert matches key
openssl x509 -modulus -noout -in server.crt | openssl md5
openssl rsa -modulus -noout -in server.key | openssl md5
```

### Health check failing
```bash
curl -v http://localhost:8080/health
curl -k https://localhost:8080/health  # For self-signed certs
```

## Performance Tuning

For large deployments:

1. **Increase monitoring interval** to reduce load:
   ```bash
   MONITORING_INTERVAL=60
   ```

2. **Adjust TCP timeout** for slower networks:
   ```bash
   MONITORING_TIMEOUT=5
   ```

3. **Run multiple instances** behind a load balancer

4. **Use connection pooling** (future enhancement)

5. **Cache aggressively** for frequently accessed pages

## Backup & Recovery

Backup your configuration:

```bash
cp config/config.yaml backups/config.yaml.$(date +%s)
tar -czf backups/dashboard-$(date +%Y%m%d).tar.gz config/ web/
```

The dashboard stores all state in-memory. To persist configuration:

1. Mount `/opt/dashboard/config` as persistent volume (if containerized)
2. Back up `config.yaml` regularly
3. Store database dumps if you add persistence layer later

---

For questions or issues, check the README.md or contact your system administrator.
