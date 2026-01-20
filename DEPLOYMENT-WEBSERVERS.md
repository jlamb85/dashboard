# Web Server Reverse Proxy Configuration

Configure Apache (httpd) or Nginx as reverse proxy for your dashboard.

---

## Nginx Reverse Proxy

### Basic Setup

Create `/etc/nginx/sites-available/dashboard`:

```nginx
upstream dashboard {
    server localhost:8080;
}

server {
    listen 80;
    server_name dashboard.example.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name dashboard.example.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/dashboard.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dashboard.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/dashboard.access.log;
    error_log /var/log/nginx/dashboard.error.log;

    # Health Check
    location /health {
        access_log off;
        proxy_pass http://dashboard;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }

    # Main Application
    location / {
        proxy_pass http://dashboard;
        proxy_http_version 1.1;
        
        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $server_name;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # Static files (if needed)
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        proxy_pass http://dashboard;
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
}
```

### Enable Site

```bash
sudo ln -s /etc/nginx/sites-available/dashboard /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### Load Balancing (Multiple Instances)

```nginx
upstream dashboard {
    least_conn;
    
    server localhost:8080 weight=1 max_fails=3 fail_timeout=30s;
    server localhost:8081 weight=1 max_fails=3 fail_timeout=30s;
    server localhost:8082 weight=1 max_fails=3 fail_timeout=30s;
    
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name dashboard.example.com;

    # ... SSL config ...

    location / {
        proxy_pass http://dashboard;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        
        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Health check
        access_log /var/log/nginx/dashboard.access.log;
    }
}
```

### Rate Limiting (Optional)

Add before server block:

```nginx
limit_req_zone $binary_remote_addr zone=dashboard:10m rate=10r/s;
limit_conn_zone $binary_remote_addr zone=addr:10m;

server {
    listen 443 ssl http2;
    
    location / {
        limit_req zone=dashboard burst=20 nodelay;
        limit_conn addr 10;
        
        proxy_pass http://dashboard;
        # ... rest of config ...
    }
}
```

### SSL Certificate (Let's Encrypt)

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Get certificate
sudo certbot certonly --nginx -d dashboard.example.com

# Auto-renewal
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```

---

## Apache/httpd Reverse Proxy

### Enable Modules

```bash
sudo a2enmod proxy
sudo a2enmod proxy_http
sudo a2enmod ssl
sudo a2enmod headers
sudo a2enmod rewrite
sudo systemctl restart apache2
```

### Virtual Host Configuration

Create `/etc/apache2/sites-available/dashboard.conf`:

```apache
<VirtualHost *:80>
    ServerName dashboard.example.com
    ServerAdmin admin@example.com
    
    # Redirect HTTP to HTTPS
    RewriteEngine On
    RewriteCond %{HTTPS} off
    RewriteRule ^(.*)$ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]
</VirtualHost>

<VirtualHost *:443>
    ServerName dashboard.example.com
    ServerAdmin admin@example.com
    
    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/dashboard.example.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/dashboard.example.com/privkey.pem
    
    # SSL Settings
    SSLProtocol TLSv1.2 TLSv1.3
    SSLCipherSuite 'HIGH:!aNULL:!MD5'
    SSLHonorCipherOrder on
    
    # Security Headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options "DENY"
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-XSS-Protection "1; mode=block"
    
    # Logging
    CustomLog ${APACHE_LOG_DIR}/dashboard.access.log combined
    ErrorLog ${APACHE_LOG_DIR}/dashboard.error.log
    
    # Reverse Proxy Configuration
    ProxyPreserveHost On
    ProxyRequests Off
    ProxyVia On
    
    # Health Check (no logging)
    <Location /health>
        SetEnvIf Request_URI "^/health$" nolog
        ProxyPass http://localhost:8080/health
        ProxyPassReverse http://localhost:8080/health
    </Location>
    
    # Main Application
    <Location />
        ProxyPass http://localhost:8080/
        ProxyPassReverse http://localhost:8080/
        
        # Connection Settings
        SetEnvIf X-Forwarded-For "^.*\..*\..*\..*" forwarded
        CustomLog ${APACHE_LOG_DIR}/dashboard.access.log combined env=!nolog
    </Location>
    
    # Static Files (Caching)
    <Location ~ "\.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$">
        ProxyPass http://localhost:8080/
        ProxyPassReverse http://localhost:8080/
        Header set Cache-Control "max-age=604800, public"
    </Location>
    
    # Timeouts
    TimeOut 300
    ProxyTimeout 300
</VirtualHost>
```

### Enable Site

```bash
sudo a2ensite dashboard
sudo apache2ctl configtest
sudo systemctl restart apache2
```

### Load Balancing (Multiple Instances)

Add to `/etc/apache2/sites-available/dashboard.conf`:

```apache
# Define upstream pool
<Proxy balancer://dashboard>
    BalancerMember http://localhost:8080 loadfactor=1
    BalancerMember http://localhost:8081 loadfactor=1
    BalancerMember http://localhost:8082 loadfactor=1
    
    # Load balancing algorithm
    ProxySet lbmethod=byrequests
    
    # Health check
    ProxySet timeout=300
    ProxySet retry=3
</Proxy>

<VirtualHost *:443>
    # ... SSL config ...
    
    <Location />
        ProxyPass balancer://dashboard/
        ProxyPassReverse balancer://dashboard/
        
        # Headers
        ProxyAddHeaders On
        SetEnvIf X-Forwarded-For "^.*\..*\..*\..*" forwarded
    </Location>
</VirtualHost>
```

### SSL Certificate (Let's Encrypt)

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-apache

# Get certificate
sudo certbot --apache -d dashboard.example.com

# Auto-renewal (automatic with certbot)
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer
```

---

## Deployment with Web Server

### Option 1: Web Server + Dashboard Service

**Start Dashboard:**
```bash
export AUTH_PASSWORD=secure123
export SERVER_ADDRESS=127.0.0.1:8080
./server-dashboard &
```

**Start Web Server:**
```bash
# Nginx
sudo systemctl start nginx

# Apache
sudo systemctl start apache2
```

### Option 2: Docker + Web Server

**Run Dashboard in Docker:**
```bash
docker run -d --name dashboard \
  -p 127.0.0.1:8080:8080 \
  -e AUTH_PASSWORD=secure123 \
  dashboard:latest
```

**Start Web Server:**
```bash
sudo systemctl start nginx  # or apache2
```

### Option 3: Multiple Dashboard Instances

**Terminal 1:**
```bash
export SERVER_PORT=8080
export AUTH_PASSWORD=secure123
./server-dashboard
```

**Terminal 2:**
```bash
export SERVER_PORT=8081
export AUTH_PASSWORD=secure123
./server-dashboard
```

**Terminal 3:**
```bash
export SERVER_PORT=8082
export AUTH_PASSWORD=secure123
./server-dashboard
```

Then configure web server with load balancing (see above).

---

## Monitoring Web Server Health

### Nginx Logs
```bash
# Real-time access log
sudo tail -f /var/log/nginx/dashboard.access.log

# Error log
sudo tail -f /var/log/nginx/dashboard.error.log

# Count requests
sudo tail -f /var/log/nginx/dashboard.access.log | grep -c "GET"
```

### Apache Logs
```bash
# Real-time access log
sudo tail -f /var/log/apache2/dashboard.access.log

# Error log
sudo tail -f /var/log/apache2/dashboard.error.log

# Monitor status
sudo apachectl status
```

### Check Backend Health
```bash
# Test health endpoint through proxy
curl https://dashboard.example.com/health

# Test direct backend
curl http://localhost:8080/health

# Monitor upstream
# Nginx: upstream status (requires separate module)
# Apache: mod_proxy_balancer status page
```

---

## Performance Tuning

### Nginx

```nginx
# In http block:
upstream dashboard {
    keepalive 32;
    server localhost:8080 max_fails=3 fail_timeout=30s;
}

server {
    # Gzip compression
    gzip on;
    gzip_types text/html text/css text/javascript application/javascript;
    gzip_min_length 1000;
    
    # Connection settings
    keepalive_timeout 65;
    
    location / {
        proxy_pass http://dashboard;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        
        # Performance
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
}
```

### Apache

```apache
# Global settings
KeepAlive On
KeepAliveTimeout 5
MaxKeepAliveRequests 100

# Gzip compression
<IfModule mod_deflate.c>
    AddOutputFilterByType DEFLATE text/html text/css text/javascript
    DeflateCompressionLevel 6
</IfModule>

<VirtualHost *:443>
    # Connection pooling
    <Proxy balancer://dashboard>
        BalancerMember http://localhost:8080 connectiontimeout=5 timeout=300
        ProxySet timeout=300 retry=3
    </Proxy>
    
    <Location />
        ProxyPass balancer://dashboard/
        ProxyPassReverse balancer://dashboard/
    </Location>
</VirtualHost>
```

---

## Troubleshooting

### Nginx Issues

**502 Bad Gateway**
```bash
# Check backend is running
curl http://localhost:8080/health

# Check nginx error log
sudo tail -f /var/log/nginx/dashboard.error.log

# Check upstream connectivity
sudo nginx -t
sudo systemctl restart nginx
```

**SSL Certificate Issues**
```bash
# Verify certificate
openssl x509 -in /etc/letsencrypt/live/dashboard.example.com/fullchain.pem -text -noout

# Check certificate expiry
sudo certbot certificates

# Force renewal
sudo certbot renew --force-renewal
```

### Apache Issues

**Connection refused**
```bash
# Check if Apache is running
sudo systemctl status apache2

# Check if dashboard is running
curl http://localhost:8080/health

# Check Apache error log
sudo tail -f /var/log/apache2/dashboard.error.log

# Test configuration
sudo apache2ctl configtest
```

**SSL Issues**
```bash
# Check certificate
sudo openssl x509 -in /etc/letsencrypt/live/dashboard.example.com/fullchain.pem -text

# Reload Apache
sudo systemctl reload apache2

# Check mod_ssl
sudo a2enmod ssl
sudo systemctl restart apache2
```

**Proxy not working**
```bash
# Enable required modules
sudo a2enmod proxy
sudo a2enmod proxy_http
sudo a2enmod headers

# Test configuration
sudo apache2ctl configtest

# Restart
sudo systemctl restart apache2
```

---

## Security Best Practices

### Both Nginx and Apache

1. **Use HTTPS Always**
   - Obtain certificate from Let's Encrypt (free)
   - Enforce HTTPS redirect from HTTP

2. **Security Headers**
   - Already configured in templates above
   - Verify with: `curl -I https://dashboard.example.com`

3. **Rate Limiting**
   - Nginx: Use `limit_req` (shown above)
   - Apache: Use `mod_ratelimit`

4. **Access Control**
   - Restrict to internal IP ranges if possible
   - Configure firewall rules

5. **Logging & Monitoring**
   - Monitor access and error logs
   - Set up log rotation
   - Alert on error spikes

### Nginx Example

```nginx
# Restrict to internal network
location / {
    # Allow dashboard admins only
    satisfy any;
    
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    
    proxy_pass http://dashboard;
}
```

### Apache Example

```apache
# Restrict to internal network
<Location />
    Require ip 192.168.1.0/24
    Require ip 10.0.0.0/8
    
    ProxyPass http://localhost:8080/
</Location>
```

---

## Full Production Setup Example

### Step 1: Start Dashboard Instances
```bash
for port in 8080 8081 8082; do
    export SERVER_PORT=$port
    export AUTH_PASSWORD=secure123
    ./server-dashboard &
done
```

### Step 2: Configure Nginx (or Apache)
```bash
# Copy template from above
sudo vi /etc/nginx/sites-available/dashboard

# Enable
sudo ln -s /etc/nginx/sites-available/dashboard /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### Step 3: Set Up HTTPS
```bash
sudo certbot certonly --nginx -d dashboard.example.com
sudo certbot renew --dry-run  # Test auto-renewal
```

### Step 4: Verify
```bash
# Test external access
curl https://dashboard.example.com/health

# Monitor logs
sudo tail -f /var/log/nginx/dashboard.access.log

# Check status
systemctl status nginx
systemctl status server-dashboard  # if using systemd
```

### Step 5: Monitor
```bash
# Set up log rotation
sudo vi /etc/logrotate.d/dashboard
# Add:
#   /var/log/nginx/dashboard*.log {
#       daily
#       rotate 7
#       compress
#       delaycompress
#       notifempty
#       create 0640 www-data www-data
#   }
```

---

## Summary

- **Nginx**: Better for high-performance, modern setups
- **Apache**: More flexible, widely used in enterprise
- **Both support**: Load balancing, SSL/TLS, security headers, caching
- **Both integrate**: With Let's Encrypt for free HTTPS

Choose based on your infrastructure and preferences!
