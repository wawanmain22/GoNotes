#!/bin/bash

# SSL/TLS Setup Script for GoNotes using Let's Encrypt
# Usage: ./scripts/setup_ssl.sh [domain] [email]

set -e

# Default values
DOMAIN="yourdomain.com"
EMAIL="your-email@example.com"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔒 SSL/TLS Setup for GoNotes${NC}"
echo "================================="

# Check if domain is provided
if [ "$1" != "" ]; then
    DOMAIN=$1
fi

if [ "$2" != "" ]; then
    EMAIL=$2
fi

echo -e "${YELLOW}Domain: $DOMAIN${NC}"
echo -e "${YELLOW}Email: $EMAIL${NC}"
echo ""

# Validate domain format
if [[ ! $DOMAIN =~ ^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$ ]]; then
    echo -e "${RED}❌ Invalid domain format. Please provide a valid domain (e.g., example.com)${NC}"
    echo "Usage: ./scripts/setup_ssl.sh yourdomain.com your-email@example.com"
    exit 1
fi

# Validate email format
if [[ ! $EMAIL =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    echo -e "${RED}❌ Invalid email format. Please provide a valid email address${NC}"
    echo "Usage: ./scripts/setup_ssl.sh yourdomain.com your-email@example.com"
    exit 1
fi

# Create necessary directories
echo -e "${GREEN}📁 Creating SSL directories...${NC}"
mkdir -p nginx certbot/conf certbot/www

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Download recommended TLS parameters
echo -e "${GREEN}📥 Downloading TLS parameters...${NC}"
curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > nginx/options-ssl-nginx.conf
curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > nginx/ssl-dhparams.pem

# Create initial nginx config for certificate generation
echo -e "${GREEN}📝 Creating initial Nginx configuration...${NC}"
cat > nginx/nginx-initial.conf << EOF
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name $DOMAIN www.$DOMAIN;

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 301 https://\$host\$request_uri;
        }
    }
}
EOF

# Start initial nginx with basic config
echo -e "${GREEN}🚀 Starting initial Nginx...${NC}"
docker run -d --name nginx-temp \
    -p 80:80 \
    -v $(pwd)/nginx/nginx-initial.conf:/etc/nginx/nginx.conf \
    -v $(pwd)/certbot/www:/var/www/certbot \
    nginx:1.24-alpine

# Wait for nginx to start
sleep 5

# Create initial certificate
echo -e "${GREEN}🔐 Creating SSL certificate...${NC}"
docker run --rm \
    -v $(pwd)/certbot/conf:/etc/letsencrypt \
    -v $(pwd)/certbot/www:/var/www/certbot \
    certbot/certbot \
    certonly --webroot -w /var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --force-renewal \
    -d $DOMAIN -d www.$DOMAIN

# Stop temporary nginx
echo -e "${GREEN}🛑 Stopping temporary Nginx...${NC}"
docker stop nginx-temp
docker rm nginx-temp

# Create final nginx config with SSL
echo -e "${GREEN}📝 Creating SSL Nginx configuration...${NC}"
cat > nginx/nginx.conf << EOF
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
        server_name $DOMAIN www.$DOMAIN;

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 301 https://\$host\$request_uri;
        }
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name $DOMAIN www.$DOMAIN;

        ssl_certificate /etc/letsencrypt/live/$DOMAIN/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/$DOMAIN/privkey.pem;

        # SSL configuration
        include /etc/letsencrypt/options-ssl-nginx.conf;
        ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

        # Security headers
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
        add_header X-Frame-Options DENY always;
        add_header X-Content-Type-Options nosniff always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Referrer-Policy "strict-origin-when-cross-origin" always;

        # Proxy to app
        location / {
            proxy_pass http://app;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
            proxy_set_header X-Forwarded-Host \$host;
            proxy_set_header X-Forwarded-Port \$server_port;
            
            # Timeout settings
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # Health check
        location /health {
            proxy_pass http://app/health;
            access_log off;
        }

        # Optional: API documentation
        location /docs {
            proxy_pass http://app/docs;
        }
    }
}
EOF

# Move SSL parameters to nginx directory
mv nginx/options-ssl-nginx.conf nginx/ssl-dhparams.pem certbot/conf/

# Update docker-compose.prod.yaml with SSL configuration
echo -e "${GREEN}📝 Updating Docker Compose configuration...${NC}"
if [ -f "docker-compose.prod.yaml" ]; then
    # Backup original file
    cp docker-compose.prod.yaml docker-compose.prod.yaml.backup

    # Update the docker-compose file to include certbot
    cat > docker-compose.ssl.yaml << EOF
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
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  nginx:
    image: nginx:1.24-alpine
    container_name: gonotes_nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot
    depends_on:
      - app
    networks:
      - gonotes-prod-network
    restart: unless-stopped
    command: "/bin/sh -c 'while :; do sleep 6h & wait \$\${!}; nginx -s reload; done & nginx -g \"daemon off;\"'"

  certbot:
    image: certbot/certbot
    container_name: gonotes_certbot
    volumes:
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait \$\${!}; done;'"

  db:
    image: postgres:15-alpine
    container_name: gonotes_db_prod
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=your_secure_production_password
      - POSTGRES_DB=gonotes
    ports:
      - "5432:5432"
    volumes:
      - postgres_prod_data:/var/lib/postgresql/data
      - ./backups:/backups
    networks:
      - gonotes-prod-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: gonotes_redis_prod
    command: redis-server --requirepass your_secure_redis_password
    ports:
      - "6379:6379"
    volumes:
      - redis_prod_data:/data
    networks:
      - gonotes-prod-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

volumes:
  postgres_prod_data:
  redis_prod_data:

networks:
  gonotes-prod-network:
    driver: bridge
EOF

    echo -e "${YELLOW}📋 SSL-enabled docker-compose.ssl.yaml created${NC}"
    echo -e "${YELLOW}💡 Use: docker-compose -f docker-compose.ssl.yaml up -d${NC}"
fi

# Create SSL renewal script
echo -e "${GREEN}📝 Creating SSL renewal script...${NC}"
cat > scripts/renew_ssl.sh << 'EOF'
#!/bin/bash

# SSL Certificate Renewal Script
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🔄 Renewing SSL certificates...${NC}"

# Renew certificates
docker-compose -f docker-compose.ssl.yaml run --rm certbot renew

# Reload nginx
echo -e "${GREEN}🔄 Reloading Nginx...${NC}"
docker-compose -f docker-compose.ssl.yaml exec nginx nginx -s reload

echo -e "${GREEN}✅ SSL certificates renewed successfully!${NC}"
EOF

chmod +x scripts/renew_ssl.sh

# Create SSL check script
echo -e "${GREEN}📝 Creating SSL check script...${NC}"
cat > scripts/check_ssl.sh << 'EOF'
#!/bin/bash

# SSL Certificate Check Script
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🔍 Checking SSL certificates...${NC}"

# List certificates
docker-compose -f docker-compose.ssl.yaml run --rm certbot certificates

echo ""
echo -e "${YELLOW}📋 Certificate expiry check:${NC}"

# Check certificate expiry for domain
DOMAIN=${1:-yourdomain.com}
if command -v openssl > /dev/null; then
    echo | openssl s_client -connect $DOMAIN:443 -servername $DOMAIN 2>/dev/null | openssl x509 -noout -dates
else
    echo -e "${RED}❌ OpenSSL not available for expiry check${NC}"
fi
EOF

chmod +x scripts/check_ssl.sh

# Update .env.prod with SSL settings
echo -e "${GREEN}📝 Updating environment variables...${NC}"
if [ -f ".env.prod" ]; then
    # Add SSL settings if not present
    if ! grep -q "SSL_CERT_PATH" .env.prod; then
        cat >> .env.prod << EOF

# SSL/TLS Configuration
DOMAIN=$DOMAIN
SSL_EMAIL=$EMAIL
SSL_CERT_PATH=/etc/letsencrypt/live/$DOMAIN/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/$DOMAIN/privkey.pem
EOF
    fi
fi

echo ""
echo -e "${GREEN}✅ SSL/TLS setup complete!${NC}"
echo ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo "1. Update your domain DNS to point to this server"
echo "2. Start your application:"
echo "   docker-compose -f docker-compose.ssl.yaml up -d"
echo "3. Test SSL: https://$DOMAIN"
echo "4. Check certificate: ./scripts/check_ssl.sh $DOMAIN"
echo ""
echo -e "${YELLOW}📋 Useful commands:${NC}"
echo "- Renew SSL: ./scripts/renew_ssl.sh"
echo "- Check SSL: ./scripts/check_ssl.sh $DOMAIN"
echo "- SSL Labs test: https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"
echo ""
echo -e "${GREEN}🎉 Your GoNotes API is now ready for HTTPS!${NC}" 