# Enhanced Docker Production Setup

This guide covers the enhanced Docker production deployment for GoNotes with comprehensive monitoring, security, and scalability features.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Prerequisites](#prerequisites)
4. [Quick Start](#quick-start)
5. [Services Overview](#services-overview)
6. [Security Features](#security-features)
7. [Monitoring & Observability](#monitoring--observability)
8. [Backup & Recovery](#backup--recovery)
9. [Scaling & Performance](#scaling--performance)
10. [Troubleshooting](#troubleshooting)

## Overview

The enhanced Docker production setup provides:

- **🔒 Enterprise Security**: Multi-layer security with SSL/TLS, rate limiting, and security headers
- **📊 Complete Monitoring**: Prometheus, Grafana, and Loki for metrics and logs
- **🔄 Automated Backups**: Scheduled database backups with compression and retention
- **⚡ High Performance**: Optimized containers with resource limits and caching
- **🚀 Easy Scaling**: Horizontal scaling support with load balancing
- **🛡️ Health Monitoring**: Comprehensive health checks and alerting

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Enhanced Production Setup                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────┐    ┌──────────┐    ┌─────────┐    ┌─────────────┐  │
│  │ Nginx   │    │ GoNotes  │    │ Postgres│    │ Redis       │  │
│  │ Proxy   │───▶│ App      │───▶│ DB      │    │ Cache       │  │
│  │         │    │          │    │         │    │             │  │
│  └─────────┘    └──────────┘    └─────────┘    └─────────────┘  │
│       │                                                          │
│       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                 Monitoring Stack                            │ │
│  │  ┌──────────┐  ┌────────┐  ┌──────┐  ┌──────────────────┐  │ │
│  │  │Prometheus│  │Grafana │  │ Loki │  │ Promtail         │  │ │
│  │  │Metrics   │  │Dashboard│  │ Logs │  │ Log Collection   │  │ │
│  │  └──────────┘  └────────┘  └──────┘  └──────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                 Support Services                            │ │
│  │  ┌──────────┐  ┌─────────────┐  ┌─────────────────────────┐ │ │
│  │  │ Backup   │  │ Health      │  │ SSL/TLS                 │ │ │
│  │  │ Service  │  │ Checker     │  │ Management              │ │ │
│  │  └──────────┘  └─────────────┘  └─────────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### System Requirements

- **CPU**: 4+ cores recommended
- **RAM**: 8GB+ recommended
- **Storage**: 50GB+ available space
- **Network**: Stable internet connection for SSL certificates

### Software Requirements

- Docker 20.10+
- Docker Compose 2.0+
- GNU Make
- Git

### Domain Requirements

- Registered domain name
- DNS pointing to your server
- Email for SSL certificate registration

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd GoNotes
```

### 2. Configure Environment

```bash
# Copy enhanced production environment
cp .env.prod.enhanced .env.prod

# Edit production values
nano .env.prod
```

**Critical values to change:**
```bash
# Database
DB_PASSWORD=your_strong_database_password

# Redis
REDIS_PASSWORD=your_strong_redis_password

# JWT
JWT_SECRET=your_very_strong_jwt_secret_key

# Grafana
GRAFANA_PASSWORD=your_grafana_admin_password

# Email/SMTP
SMTP_PASSWORD=your_smtp_password

# Domain
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 3. Deploy Enhanced Production

```bash
# Option 1: Using Make
make prod-enhanced

# Option 2: Using Aliases
source aliases.sh
gn-prod-enhanced

# Option 3: Manual Docker Compose
docker-compose -f docker-compose.prod.yaml up -d
```

### 4. Verify Deployment

```bash
# Check service health
make prod-health

# View service status
make prod-status

# Access services
echo "Application: http://localhost:8080"
echo "Prometheus: http://localhost:9090"
echo "Grafana: http://localhost:3000"
```

## Services Overview

### Core Application Services

#### 1. GoNotes Application (`app`)
- **Image**: Custom built from `Dockerfile.prod`
- **Port**: 8080
- **Features**: 
  - Multi-stage optimized build
  - Non-root user execution
  - Health checks
  - Resource limits (512MB RAM, 1 CPU)

#### 2. PostgreSQL Database (`db`)
- **Image**: `postgres:15-alpine`
- **Port**: 5432 (localhost only)
- **Features**:
  - Enhanced security configuration
  - Performance optimization
  - SSL/TLS encryption
  - Connection pooling

#### 3. Redis Cache (`redis`)
- **Image**: `redis:7-alpine`
- **Port**: 6379 (localhost only)
- **Features**:
  - Password protection
  - Memory optimization (256MB limit)
  - AOF persistence
  - Performance tuning

#### 4. Nginx Reverse Proxy (`nginx`)
- **Image**: `nginx:1.24-alpine`
- **Ports**: 80, 443
- **Features**:
  - SSL/TLS termination
  - Rate limiting
  - Security headers
  - Gzip compression
  - Caching

### Monitoring Services

#### 1. Prometheus (`prometheus`)
- **Port**: 9090 (localhost only)
- **Features**: Metrics collection and alerting
- **Retention**: 15 days
- **Storage**: 10GB limit

#### 2. Grafana (`grafana`)
- **Port**: 3000 (localhost only)
- **Features**: Metrics visualization and dashboards
- **Default Login**: admin / (from GRAFANA_PASSWORD)

#### 3. Loki (`loki`)
- **Port**: 3100 (localhost only)
- **Features**: Log aggregation and querying

#### 4. Promtail (`promtail`)
- **Features**: Log collection from all services

### Support Services

#### 1. Backup Service (`backup`)
- **Schedule**: Daily at 2 AM (configurable)
- **Retention**: 30 days (configurable)
- **Features**:
  - Automated compression
  - Integrity verification
  - Email/Slack notifications

#### 2. Health Checker (`healthcheck`)
- **Features**: Continuous health monitoring
- **Interval**: 60 seconds
- **Alerts**: Log failures

## Security Features

### Network Security

- **Isolated Networks**: Custom Docker bridge network
- **Port Binding**: Critical services bound to localhost only
- **Firewall Ready**: External access only through Nginx

### Application Security

- **Non-root Execution**: All containers run as non-root users
- **Read-only Filesystems**: Containers use read-only root filesystems
- **Security Options**: `no-new-privileges` enabled
- **Resource Limits**: CPU and memory limits prevent resource exhaustion

### SSL/TLS Security

- **Modern Protocols**: TLS 1.2 and 1.3 only
- **Strong Ciphers**: ECDHE and ChaCha20-Poly1305
- **HSTS**: HTTP Strict Transport Security enabled
- **Perfect Forward Secrecy**: ECDHE key exchange

### Data Security

- **Database Encryption**: SSL connections required
- **Password Hashing**: SCRAM-SHA-256 for PostgreSQL
- **JWT Security**: Strong secret keys and short expiration
- **Rate Limiting**: Multiple layers of rate limiting

## Monitoring & Observability

### Metrics Collection

Prometheus collects metrics from:
- **Application**: Custom GoNotes metrics
- **Database**: Connection pools, query performance
- **Cache**: Redis hit rates, memory usage
- **Web Server**: Request rates, response times
- **System**: CPU, memory, disk usage

### Key Dashboards

#### Application Dashboard
- Request rate and response times
- Error rates and status codes
- Active users and sessions
- Database query performance

#### Infrastructure Dashboard
- CPU and memory usage
- Disk space and I/O
- Network traffic
- Container health

#### Security Dashboard
- Failed authentication attempts
- Rate limiting triggers
- Suspicious IP addresses
- Security alert notifications

### Log Management

#### Structured Logging
- JSON format for easy parsing
- Correlation IDs for request tracing
- Different log levels per service
- Centralized log aggregation

#### Log Sources
- Application logs
- Access logs (Nginx)
- Error logs
- Security audit logs
- System logs

### Alerting

#### Critical Alerts
- Service downtime
- Database connection failures
- High error rates (>5%)
- Resource exhaustion (>90%)

#### Warning Alerts
- High response times (>2s)
- Memory usage (>85%)
- Disk space (>80%)
- Failed backup jobs

## Backup & Recovery

### Automated Backups

#### Schedule
- **Daily**: 2:00 AM (configurable with `BACKUP_SCHEDULE`)
- **Retention**: 30 days (configurable with `BACKUP_RETENTION_DAYS`)
- **Compression**: Level 6 gzip compression

#### Features
- Database size calculation
- Compression ratio reporting
- Integrity verification
- Email/Slack notifications
- Automatic cleanup of old backups

#### Manual Backup

```bash
# Run immediate backup
make prod-backup-manual

# Or using aliases
gn-prod-backup-manual
```

### Backup Verification

Each backup includes:
- File existence check
- Gzip integrity test
- SQL content validation
- Size comparison

### Recovery Process

#### Database Recovery

```bash
# Stop application
docker-compose -f docker-compose.prod.yaml stop app

# Restore from backup
docker exec -i gonotes_db_prod psql -U postgres gonotes < backups/backup_file.sql

# Start application
docker-compose -f docker-compose.prod.yaml start app
```

#### Point-in-Time Recovery

PostgreSQL WAL archiving is enabled for point-in-time recovery:

```bash
# Restore to specific timestamp
docker exec gonotes_db_prod pg_basebackup -D /var/lib/postgresql/restore \
  -Ft -z -P -U postgres -h localhost -p 5432
```

## Scaling & Performance

### Horizontal Scaling

#### Application Scaling

```bash
# Scale to 3 app instances
make prod-scale APP_REPLICAS=3

# Or using environment variable
APP_REPLICAS=5 make prod-scale
```

#### Load Balancer Configuration

Nginx automatically load balances between app instances:

```nginx
upstream gonotes_backend {
    server app_1:8080;
    server app_2:8080;
    server app_3:8080;
    
    keepalive 32;
    keepalive_requests 1000;
    keepalive_timeout 60s;
}
```

### Performance Optimization

#### Database Optimization
- Connection pooling (25 connections)
- Query performance logging
- Index optimization
- Autovacuum tuning

#### Cache Optimization
- Redis memory limits
- LRU eviction policy
- AOF persistence
- Key expiration strategies

#### Web Server Optimization
- Gzip compression
- Static file caching
- HTTP/2 support
- Connection keep-alive

### Resource Monitoring

#### CPU Usage
- Application: 50-100% of 1 CPU
- Database: 50-200% of 2 CPUs
- Redis: 25-50% of 0.5 CPU
- Nginx: 25-50% of 0.5 CPU

#### Memory Usage
- Application: 256-512MB
- Database: 512MB-1GB
- Redis: 128-256MB
- Nginx: 128-256MB

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

```bash
# Check service logs
docker-compose -f docker-compose.prod.yaml logs service_name

# Check service status
docker-compose -f docker-compose.prod.yaml ps

# Restart specific service
docker-compose -f docker-compose.prod.yaml restart service_name
```

#### 2. Database Connection Issues

```bash
# Check database logs
docker-compose -f docker-compose.prod.yaml logs db

# Test database connectivity
docker-compose -f docker-compose.prod.yaml exec app \
  psql -h db -U postgres -d gonotes -c "SELECT version();"

# Check network connectivity
docker-compose -f docker-compose.prod.yaml exec app ping db
```

#### 3. High Memory Usage

```bash
# Check container resource usage
docker stats

# Check application metrics
curl http://localhost:9090/api/v1/query?query=container_memory_usage_bytes

# Restart services with memory issues
docker-compose -f docker-compose.prod.yaml restart service_name
```

#### 4. SSL Certificate Issues

```bash
# Check certificate expiration
openssl x509 -in nginx/ssl/fullchain.pem -noout -dates

# Verify certificate chain
openssl verify -CAfile nginx/ssl/chain.pem nginx/ssl/fullchain.pem

# Test SSL configuration
curl -I https://yourdomain.com
```

### Health Check Commands

```bash
# Application health
curl -f http://localhost:8080/health

# Prometheus health
curl -f http://localhost:9090/-/healthy

# Nginx health
curl -f http://localhost/health

# Database health
docker-compose -f docker-compose.prod.yaml exec db pg_isready -U postgres

# Redis health
docker-compose -f docker-compose.prod.yaml exec redis redis-cli ping
```

### Performance Debugging

#### Slow Queries
```bash
# Enable slow query logging
docker-compose -f docker-compose.prod.yaml exec db \
  psql -U postgres -d gonotes -c "
    ALTER SYSTEM SET log_min_duration_statement = 100;
    SELECT pg_reload_conf();
  "

# View slow queries
docker-compose -f docker-compose.prod.yaml logs db | grep "duration:"
```

#### Memory Leaks
```bash
# Monitor memory usage over time
watch 'docker stats --no-stream | grep gonotes'

# Check for memory leaks in application
curl http://localhost:8080/debug/pprof/heap
```

#### High CPU Usage
```bash
# Check CPU usage per container
docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"

# Profile application CPU usage
curl http://localhost:8080/debug/pprof/profile?seconds=30
```

### Log Analysis

#### Application Logs
```bash
# View recent application logs
docker-compose -f docker-compose.prod.yaml logs --tail=100 -f app

# Search for errors
docker-compose -f docker-compose.prod.yaml logs app | grep ERROR

# Monitor authentication failures
docker-compose -f docker-compose.prod.yaml logs app | grep "authentication failed"
```

#### Access Logs
```bash
# View Nginx access logs
docker-compose -f docker-compose.prod.yaml logs nginx | grep "GET\|POST"

# Monitor 4xx/5xx errors
docker-compose -f docker-compose.prod.yaml logs nginx | grep " [45][0-9][0-9] "

# Check rate limiting
docker-compose -f docker-compose.prod.yaml logs nginx | grep "limiting"
```

## Maintenance

### Regular Maintenance Tasks

#### Daily
- Check service health: `make prod-health`
- Monitor resource usage: `docker stats`
- Review error logs: `make prod-logs | grep ERROR`

#### Weekly
- Update Docker images: `docker-compose pull && make prod-deploy`
- Review backup reports: `cat backups/backup_report.txt`
- Check SSL certificate expiration: `make ssl-check`

#### Monthly
- Update base images and rebuild: `make prod-build prod-deploy`
- Review and optimize database indexes
- Analyze performance metrics and trends
- Update security configurations

### Updates and Upgrades

#### Application Updates
```bash
# Pull latest code
git pull origin main

# Rebuild and deploy
make prod-build prod-deploy
```

#### System Updates
```bash
# Update Docker images
docker-compose -f docker-compose.prod.yaml pull

# Recreate containers with new images
docker-compose -f docker-compose.prod.yaml up -d --force-recreate
```

This enhanced Docker production setup provides enterprise-grade reliability, security, and observability for your GoNotes application. For additional customization and advanced configurations, refer to the individual service documentation files. 