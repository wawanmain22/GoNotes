# Development Dockerfile Guide

## Overview

`Dockerfile.dev` is specifically designed for development workflow, providing a complete development environment with hot reload capabilities and debugging tools.

## Key Features

### 🔧 Development Tools
- **Air**: Hot reload for Go applications
- **Migrate**: Database migration tool
- **Go Tools**: goimports, golangci-lint, swag
- **Build Tools**: gcc, make, bash
- **Debug Support**: Delve debugging port (2345)

### 📦 Optimized for Development
- **Not Multi-stage**: Single stage for faster builds
- **Volume Mounts**: Source code mounted for instant changes
- **Go Module Cache**: Cached between builds
- **Debug Mode**: GIN_MODE=debug, full logging

### 🚀 Hot Reload
- **Air Integration**: Automatic restart on code changes
- **Fast Rebuilds**: Cached dependencies and build artifacts
- **Debug Mode**: Full error traces and logging

## Usage

### Docker Compose (Recommended)
```bash
# Start complete development environment
make dev-start

# Or start individual services
make dev-db          # Database and Redis
make dev-app         # Application container
make dev-tools       # Redis Commander, Adminer
```

### Direct Docker Commands
```bash
# Build development image
docker build -f Dockerfile.dev -t gonotes:dev .

# Run development container
docker run -p 8080:8080 -p 2345:2345 \
  -v $(pwd):/app:cached \
  -v go_mod_cache:/go/pkg/mod \
  gonotes:dev
```

### Local Development
```bash
# Traditional local development (no container)
make dev-run         # Uses air for hot reload

# Or directly
air -c .air.toml
```

## Container Configuration

### Environment Variables
```bash
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gonotes_dev
REDIS_HOST=redis
REDIS_PORT=6379
JWT_SECRET=dev_secret_key_change_in_production
GIN_MODE=debug
LOG_LEVEL=debug
```

### Exposed Ports
- **8080**: HTTP server
- **2345**: Delve debug port

### Volume Mounts
- **Source Code**: `.:/app:cached`
- **Go Modules**: `go_mod_cache:/go/pkg/mod`
- **Build Cache**: `go_build_cache:/root/.cache/go-build`

## Development Workflow

### 1. Initial Setup
```bash
# Clone repository
git clone <repository-url>
cd gonotes

# Start development environment
make dev-start
```

### 2. Development Process
```bash
# Make code changes
# Files are automatically reloaded via Air

# Run tests
make test-unit

# Check code quality
make check
```

### 3. Database Operations
```bash
# Access database
make db-shell

# Run migrations
make dev-migrate

# Access Redis
make redis-shell
```

### 4. Debugging
```bash
# View logs
make logs

# Access application shell
make shell

# Debug with Delve (port 2345)
dlv connect localhost:2345
```

## Development Tools

### Database Management
- **Adminer**: http://localhost:8082
- **Direct Access**: `make db-shell`

### Redis Management
- **Redis Commander**: http://localhost:8081
- **Direct Access**: `make redis-shell`

### Code Quality
```bash
# Format code
make fmt

# Run linters
make lint

# Import organization
goimports -w .
```

## Comparison with Production

| Feature | Development | Production |
|---------|-------------|------------|
| **Build Type** | Single stage | Multi-stage |
| **Size** | ~800MB | ~50MB |
| **Build Time** | ~2 minutes | ~3 minutes |
| **Hot Reload** | ✅ Yes | ❌ No |
| **Debug Tools** | ✅ Yes | ❌ No |
| **Optimization** | ❌ No | ✅ Yes |
| **Security** | Basic | Enhanced |

## Best Practices

### 1. Volume Mounting
```yaml
volumes:
  - .:/app:cached          # Source code
  - go_mod_cache:/go/pkg/mod  # Go modules
  - go_build_cache:/root/.cache/go-build  # Build cache
```

### 2. Environment Variables
```yaml
environment:
  - GIN_MODE=debug
  - LOG_LEVEL=debug
  - DB_NAME=gonotes_dev    # Different from production
```

### 3. Port Configuration
```yaml
ports:
  - "8080:8080"    # HTTP
  - "2345:2345"    # Debug
```

## Troubleshooting

### Common Issues

**Container Won't Start:**
```bash
# Check logs
docker-compose -f docker-compose.dev.yaml logs app

# Rebuild image
make dev-build
```

**Hot Reload Not Working:**
```bash
# Check air configuration
cat .air.toml

# Verify volume mounts
docker-compose -f docker-compose.dev.yaml config
```

**Database Connection Issues:**
```bash
# Check database status
make dev-db

# Verify connection
make db-shell
```

**Port Conflicts:**
```bash
# Check port usage
lsof -i :8080

# Stop conflicting services
make dev-stop
```

## Performance Tips

### 1. Build Performance
- Use `.dockerignore` to exclude unnecessary files
- Leverage Go module cache
- Use cached volume mounts

### 2. Runtime Performance
- Use `cached` volume mount option
- Limit container resources appropriately
- Use SSD storage for better I/O

### 3. Development Efficiency
- Use Air for hot reload
- Configure IDE for container debugging
- Use development database with sample data

## Security Considerations

### Development Security
- Use different secrets from production
- Limit network exposure
- Use development-specific database
- Regular dependency updates

### Container Security
- Non-root user in container
- Limited capabilities
- Isolated network
- Regular base image updates

## Integration with IDE

### Visual Studio Code
```json
{
  "go.toolsGopath": "/go",
  "go.gopath": "/go",
  "go.goroot": "/usr/local/go"
}
```

### GoLand
- Configure Docker integration
- Set up remote debugging
- Use Go modules support

## Conclusion

`Dockerfile.dev` provides a complete development environment with:
- ✅ Hot reload capabilities
- ✅ Debug tools and ports
- ✅ Database and Redis integration
- ✅ Code quality tools
- ✅ Fast development workflow

This setup enables efficient development while maintaining consistency with production environment. 