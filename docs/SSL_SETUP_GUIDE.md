# SSL/TLS Setup Guide - GoNotes

Panduan lengkap untuk setup SSL/TLS gratis menggunakan Let's Encrypt dan integrasi dengan Docker.

## 🔒 Pilihan SSL/TLS Gratis

### 1. Let's Encrypt (Recommended)
- ✅ **Gratis** dan otomatis renewal
- ✅ **Trusted** oleh semua browser
- ✅ **Easy setup** dengan Certbot
- ✅ **Wildcard certificates** tersedia
- ✅ **90 hari** validity, auto-renewal

### 2. Cloudflare SSL
- ✅ **Gratis** dengan Cloudflare account
- ✅ **Easy setup** melalui dashboard
- ✅ **Additional security** features
- ✅ **CDN** dan performance boost

### 3. AWS Certificate Manager
- ✅ **Gratis** untuk AWS services
- ✅ **Auto-renewal**
- ✅ **Integration** dengan AWS Load Balancer

## 🚀 Let's Encrypt Setup (Recommended)

### Method 1: Docker dengan Certbot + Nginx

#### 1. Update Docker Compose Production
```yaml
# docker-compose.prod.yaml - SSL Version
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.prod
    container_name: gonotes_app
    expose:
      - "8080"
    env_file:
      - .env.prod
    depends_on:
      - db
      - redis
    networks:
      - gonotes-prod-network
    restart: unless-stopped

  nginx:
    image: nginx:1.24-alpine
    container_name: gonotes_nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot
    depends_on:
      - app
    networks:
      - gonotes-prod-network
    restart: unless-stopped
    command: "/bin/sh -c 'while :; do sleep 6h & wait $${!}; nginx -s reload; done & nginx -g \"daemon off;\"'"

  certbot:
    image: certbot/certbot
    container_name: gonotes_certbot
    volumes:
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
```

#### 2. Nginx Configuration dengan SSL
```nginx
# nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream app {
        server app:8080;
    }

    # HTTP redirect to HTTPS
    server {
        listen 80;
        server_name yourdomain.com www.yourdomain.com;

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 301 https://$host$request_uri;
        }
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name yourdomain.com www.yourdomain.com;

        ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

        # SSL configuration
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384;
        ssl_prefer_server_ciphers off;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;

        # Security headers
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # Proxy to app
        location / {
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $host;
            proxy_set_header X-Forwarded-Port $server_port;
        }

        # Health check
        location /health {
            proxy_pass http://app/health;
        }
    }
}
```

#### 3. SSL Setup Script
```bash
#!/bin/bash
# scripts/setup_ssl.sh

# SSL/TLS Setup Script for GoNotes
set -e

DOMAIN="yourdomain.com"
EMAIL="your-email@example.com"

echo "🔒 Setting up SSL/TLS for GoNotes"
echo "================================="

# Create necessary directories
mkdir -p nginx certbot/conf certbot/www

# Check if domain is provided
if [ "$1" != "" ]; then
    DOMAIN=$1
fi

if [ "$2" != "" ]; then
    EMAIL=$2
fi

echo "Domain: $DOMAIN"
echo "Email: $EMAIL"

# Download recommended TLS parameters
echo "📥 Downloading TLS parameters..."
curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > nginx/options-ssl-nginx.conf
curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > nginx/ssl-dhparams.pem

# Create initial certificate
echo "🔐 Creating initial certificate..."
docker-compose -f docker-compose.prod.yaml run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d $DOMAIN -d www.$DOMAIN" certbot

echo "✅ SSL/TLS setup complete!"
echo "🔄 Reloading Nginx..."
docker-compose -f docker-compose.prod.yaml exec nginx nginx -s reload

echo "🎉 SSL/TLS is now active for $DOMAIN"
```

### Method 2: Manual Certbot Setup

#### 1. Install Certbot
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install certbot python3-certbot-nginx

# macOS
brew install certbot
```

#### 2. Obtain Certificate
```bash
# For nginx
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# For standalone (stop your app first)
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# For webroot (while app is running)
sudo certbot certonly --webroot -w /var/www/html -d yourdomain.com -d www.yourdomain.com
```

#### 3. Auto-renewal Setup
```bash
# Test auto-renewal
sudo certbot renew --dry-run

# Add to crontab
sudo crontab -e

# Add this line for auto-renewal
0 12 * * * /usr/bin/certbot renew --quiet
```

## 🌐 Cloudflare SSL Setup

### 1. Cloudflare Account Setup
1. Buat account di [Cloudflare](https://cloudflare.com)
2. Add domain ke Cloudflare
3. Update nameserver domain ke Cloudflare
4. Wait for activation (biasanya 5-10 menit)

### 2. SSL Configuration
```bash
# Cloudflare Dashboard
1. Go to SSL/TLS tab
2. Choose "Full (strict)" mode
3. Enable "Always Use HTTPS"
4. Enable "HSTS"
5. Enable "Minimum TLS Version 1.2"
```

### 3. Origin Certificate
```bash
# Generate Origin Certificate di Cloudflare
1. Go to SSL/TLS → Origin Server
2. Click "Create Certificate"
3. Choose "Let Cloudflare generate a private key and CSR"
4. Select validity (15 years)
5. Download certificate dan private key
```

### 4. Nginx Configuration untuk Cloudflare
```nginx
# nginx/nginx-cloudflare.conf
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/cloudflare.pem;
    ssl_certificate_key /etc/nginx/ssl/cloudflare.key;

    # Cloudflare IP ranges only
    set_real_ip_from 103.21.244.0/22;
    set_real_ip_from 103.22.200.0/22;
    set_real_ip_from 103.31.4.0/22;
    # ... add more Cloudflare IP ranges
    real_ip_header CF-Connecting-IP;

    location / {
        proxy_pass http://app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔧 Updated Makefile Commands

```makefile
# Add to Makefile
.PHONY: ssl-setup
ssl-setup: ## Setup SSL/TLS with Let's Encrypt
	@echo "$(GREEN)Setting up SSL/TLS...$(NC)"
	@./scripts/setup_ssl.sh
	@echo "$(GREEN)SSL/TLS setup complete!$(NC)"

.PHONY: ssl-renew
ssl-renew: ## Renew SSL certificates
	@echo "$(GREEN)Renewing SSL certificates...$(NC)"
	@docker-compose -f docker-compose.prod.yaml run --rm certbot renew
	@docker-compose -f docker-compose.prod.yaml exec nginx nginx -s reload
	@echo "$(GREEN)SSL certificates renewed!$(NC)"

.PHONY: ssl-check
ssl-check: ## Check SSL certificate expiry
	@echo "$(GREEN)Checking SSL certificate...$(NC)"
	@docker-compose -f docker-compose.prod.yaml run --rm certbot certificates
```

## 🔍 SSL Testing & Verification

### 1. Online SSL Checkers
```bash
# SSL Labs Test
https://www.ssllabs.com/ssltest/analyze.html?d=yourdomain.com

# SSL Checker
https://www.sslchecker.com/sslchecker

# Security Headers
https://securityheaders.com/?q=yourdomain.com
```

### 2. Command Line Testing
```bash
# Test SSL connection
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com

# Check certificate expiry
echo | openssl s_client -connect yourdomain.com:443 -servername yourdomain.com 2>/dev/null | openssl x509 -noout -dates

# Test with curl
curl -I https://yourdomain.com
```

### 3. Browser Testing
```bash
# Check browser address bar
- Green lock icon
- "Secure" or "Connection is secure"
- Valid certificate details

# Check developer tools
- Security tab should show "Secure"
- No mixed content warnings
```

## 📋 Production Deployment dengan SSL

### 1. Complete Production Setup
```bash
# 1. Setup production environment
make prod-setup

# 2. Edit .env.prod dengan domain yang benar
vi .env.prod

# 3. Update nginx config dengan domain
vi nginx/nginx.conf

# 4. Build dan start services
make prod-build
make prod-start

# 5. Setup SSL
make ssl-setup

# 6. Verify SSL
make ssl-check
```

### 2. Environment Variables untuk SSL
```env
# .env.prod - SSL Configuration
DOMAIN=yourdomain.com
SSL_EMAIL=your-email@example.com
SSL_CERT_PATH=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/yourdomain.com/privkey.pem
```

## 🔄 Automated SSL Renewal

### 1. Docker Compose Auto-renewal
```yaml
# Sudah included di docker-compose.prod.yaml
certbot:
  image: certbot/certbot
  volumes:
    - ./certbot/conf:/etc/letsencrypt
    - ./certbot/www:/var/www/certbot
  entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
```

### 2. Crontab Setup (Alternative)
```bash
# Add to crontab
0 12 * * * /usr/bin/make ssl-renew
```

## 🚨 Troubleshooting

### Common Issues

1. **Domain not pointing to server**
   ```bash
   # Check DNS resolution
   nslookup yourdomain.com
   dig yourdomain.com
   ```

2. **Port 80/443 not accessible**
   ```bash
   # Check firewall
   sudo ufw status
   sudo ufw allow 80
   sudo ufw allow 443
   ```

3. **Certificate renewal fails**
   ```bash
   # Check logs
   docker-compose -f docker-compose.prod.yaml logs certbot
   
   # Manual renewal
   docker-compose -f docker-compose.prod.yaml run --rm certbot renew --dry-run
   ```

4. **Nginx fails to start**
   ```bash
   # Check nginx config
   docker-compose -f docker-compose.prod.yaml exec nginx nginx -t
   
   # Check logs
   docker-compose -f docker-compose.prod.yaml logs nginx
   ```

## 💰 Cost Comparison

| Solution | Cost | Renewal | Difficulty |
|----------|------|---------|------------|
| Let's Encrypt | **Free** | Auto | Easy |
| Cloudflare | **Free** | Auto | Very Easy |
| AWS ACM | **Free** | Auto | Medium |
| Commercial CA | $10-100/year | Manual | Easy |

## 🎯 Recommendation

**Untuk GoNotes project, saya recommend:**

1. **Let's Encrypt** - Jika punya VPS/server sendiri
2. **Cloudflare** - Jika mau tambahan CDN dan security
3. **AWS ACM** - Jika deploy di AWS

**Let's Encrypt adalah pilihan terbaik karena:**
- ✅ Completely free
- ✅ Auto-renewal
- ✅ Trusted by all browsers
- ✅ Easy Docker integration
- ✅ Wildcard certificates available

## 🔗 Useful Links

- [Let's Encrypt](https://letsencrypt.org/)
- [Certbot](https://certbot.eff.org/)
- [Cloudflare](https://cloudflare.com)
- [SSL Labs Test](https://www.ssllabs.com/ssltest/)
- [Security Headers](https://securityheaders.com/)

---

**Total setup time: 10-15 minutes**
**Annual cost: $0 (FREE!)**
**Maintenance: Fully automated** 