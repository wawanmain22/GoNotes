# GoNotes Development Guide

Panduan lengkap untuk development dan deployment aplikasi GoNotes.

## 📁 Struktur Project

```
gonotes/
├── cmd/                     # Entry point aplikasi
├── internal/                # Internal packages
│   ├── handler/            # HTTP handlers
│   ├── service/            # Business logic
│   ├── repository/         # Data access layer
│   ├── model/              # Data models
│   ├── middleware/         # HTTP middleware
│   ├── config/             # Configuration
│   └── utils/              # Utility functions
├── migrations/             # Database migrations
├── docs/                   # Documentation
├── scripts/                # Shell scripts
│   ├── test_api.sh         # API testing
│   ├── test_notes_api.sh   # Notes API testing
│   ├── test_profile_api.sh # Profile API testing
│   ├── test_session_security_api.sh # Session testing
│   └── test_integration_docker.sh   # Integration testing
├── .env.dev               # Development environment
├── .env.prod              # Production environment
├── docker-compose.dev.yaml # Development Docker Compose
├── docker-compose.prod.yaml # Production Docker Compose
├── Dockerfile             # Development Dockerfile
├── Dockerfile.prod        # Production Dockerfile
└── Makefile              # Command aliases
```

## 🚀 Quick Start

### 1. Setup Development Environment

```bash
# Clone project
git clone <repository-url>
cd gonotes

# Setup development environment
make dev-setup

# Start database services
make dev-db

# Run database migrations
make dev-migrate

# Start application with hot reload
make dev-run
```

### 2. Using Make Commands

```bash
# Show all available commands
make help

# Start complete development environment
make dev-start

# Run all tests
make test-all

# Deploy to production
make prod-deploy
```

## 🛠️ Development Commands

### Environment Setup
```bash
make dev-setup          # Setup development environment
make dev-db             # Start database and Redis
make dev-tools          # Start development tools (Adminer, Redis Commander)
make dev-migrate        # Run database migrations
make dev-migrate-down   # Rollback migrations
make dev-run            # Run app with hot reload
make dev-start          # Start complete development environment
make dev-stop           # Stop development services
make dev-clean          # Clean development environment
```

### Testing Commands
```bash
make test-unit          # Run unit tests
make test-integration   # Run integration tests
make test-api           # Run API tests
make test-notes         # Run notes API tests
make test-profile       # Run profile API tests
make test-session       # Run session security tests
make test-all           # Run all tests
```

### Build & Utility Commands
```bash
make build              # Build application binary
make clean              # Clean build artifacts
make deps               # Download dependencies
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint
make check              # Run all code checks
```

## 🏭 Production Commands

### Setup & Deployment
```bash
make prod-setup         # Setup production environment
make prod-build         # Build production Docker image
make prod-start         # Start production services
make prod-stop          # Stop production services
make prod-migrate       # Run production migrations
make prod-deploy        # Complete deployment
```

### SSL/TLS Commands
```bash
make ssl-setup          # Setup SSL with Let's Encrypt (requires DOMAIN and EMAIL)
make ssl-start          # Start production services with SSL
make ssl-stop           # Stop SSL production services
make ssl-renew          # Renew SSL certificates
make ssl-check          # Check SSL certificate status
make ssl-test           # Test SSL configuration online
make ssl-logs           # View SSL production logs
```

### Monitoring & Maintenance
```bash
make prod-logs          # View production logs
make prod-status        # Check services status
make prod-backup        # Backup database
make prod-restore       # Restore database (requires BACKUP_FILE)
make health             # Check application health
make stats              # Show container resource usage
```

### Shell Access
```bash
make shell              # Open shell in app container
make db-shell           # Open database shell
make redis-shell        # Open Redis shell
```

## 📝 Development Workflow

### 1. Daily Development
```bash
# Start development environment
make dev-start

# Your code changes are automatically reloaded by Air
# Open your browser to http://localhost:8080

# Run tests during development
make test-unit

# Format and check code
make check
```

### 2. Testing Workflow
```bash
# Run specific test suites
make test-unit          # Unit tests
make test-api           # API tests
make test-notes         # Notes functionality
make test-profile       # Profile functionality
make test-session       # Session security

# Run integration tests (requires Docker)
make test-integration

# Run all tests before committing
make test-all
```

### 3. Production Deployment
```bash
# Setup production environment
make prod-setup

# Edit .env.prod with your production values
vi .env.prod

# Build and deploy
make prod-deploy

# Monitor the deployment
make prod-logs
make prod-status
```

### 4. SSL/TLS Setup (FREE!)
```bash
# Setup SSL with Let's Encrypt (FREE!)
make ssl-setup DOMAIN=yourdomain.com EMAIL=your-email@example.com

# Start production with SSL
make ssl-start

# Check SSL certificate
make ssl-check

# Test SSL online
make ssl-test
```

## 🔧 Manual Commands (Alternative)

### Development
```bash
# Start database manually
docker-compose -f docker-compose.dev.yaml up -d db redis

# Run migrations manually
migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/gonotes_dev?sslmode=disable" up

# Run app with Air
air

# Run specific test script
./scripts/test_api.sh
```

### Production
```bash
# Build production image
docker build -f Dockerfile.prod -t gonotes:latest .

# Start production services
docker-compose -f docker-compose.prod.yaml up -d

# View logs
docker-compose -f docker-compose.prod.yaml logs -f
```

## 🌐 Service URLs

### Development
- **API**: http://localhost:8080
- **Database**: localhost:5432
- **Redis**: localhost:6379
- **Adminer**: http://localhost:8082
- **Redis Commander**: http://localhost:8081

### Production
- **API**: http://localhost:8080 (or https://yourdomain.com with SSL)
- **Database**: localhost:5432
- **Redis**: localhost:6379

### SSL/TLS Testing
- **SSL Labs**: https://www.ssllabs.com/ssltest/
- **Security Headers**: https://securityheaders.com/
- **Certificate Info**: `make ssl-check`

## 🔑 Environment Variables

### Development (.env.dev)
```env
APP_PORT=8080
APP_ENV=development
DB_HOST=localhost
DB_NAME=gonotes_dev
JWT_SECRET=dev_supersecretkey_for_testing_only
LOG_LEVEL=debug
```

### Production (.env.prod)
```env
APP_PORT=8080
APP_ENV=production
DB_HOST=db
DB_NAME=gonotes
JWT_SECRET=your_very_secure_jwt_secret_key_at_least_32_characters_long
LOG_LEVEL=info
```

## 🧪 Testing

### Unit Tests
```bash
# Run all unit tests
go test -v ./internal/service/...

# Run specific test file
go test -v ./internal/service/user_test.go

# Run with coverage
go test -v -cover ./internal/service/...
```

### Integration Tests
```bash
# Run integration tests with Docker
./scripts/test_integration_docker.sh

# Run API tests
./scripts/test_api.sh

# Run specific feature tests
./scripts/test_notes_api.sh
./scripts/test_profile_api.sh
./scripts/test_session_security_api.sh
```

## 📊 Monitoring

### Health Check
```bash
# Check application health
curl http://localhost:8080/health

# Using make command
make health
```

### Container Stats
```bash
# Show resource usage
make stats

# View logs
make logs ENV=dev    # Development logs
make logs ENV=prod   # Production logs
```

## 🔒 Security

### Development
- Less strict rate limiting
- Debug logging enabled
- CORS allows all origins
- Weaker JWT secret (for testing)

### Production
- Strict rate limiting
- Info-level logging
- Restricted CORS origins
- Strong JWT secret required
- SSL/TLS ready
- Health checks enabled

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Error**
   ```bash
   # Check if database is running
   docker ps
   
   # Restart database
   make dev-db
   ```

2. **Migration Issues**
   ```bash
   # Check migration status
   migrate -path ./migrations -database "$(MIGRATE_URL)" version
   
   # Force migration version
   migrate -path ./migrations -database "$(MIGRATE_URL)" force <version>
   ```

3. **Port Already in Use**
   ```bash
   # Find process using port
   lsof -i :8080
   
   # Kill process
   kill -9 <PID>
   ```

4. **Docker Issues**
   ```bash
   # Clean Docker system
   make dev-clean
   
   # Remove all containers and images
   docker system prune -a
   ```

## 📚 Additional Resources

- [API Documentation](docs/API_Documentation.md)
- [Postman Collection](docs/GoNotes_API_Collection.postman_collection.json)
- [Collection Guide](docs/README_Collection.md)

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Make changes and run tests: `make test-all`
4. Format code: `make check`
5. Commit changes: `git commit -m 'Add amazing feature'`
6. Push to branch: `git push origin feature/amazing-feature`
7. Create Pull Request

## 📞 Support

### Development Issues
- Check this guide first
- Run `make help` for available commands
- Check logs with `make logs`
- Run health check with `make health`

### Production Issues
- Check service status: `make prod-status`
- View logs: `make prod-logs`
- Check health: `make health`
- Monitor resources: `make stats` 