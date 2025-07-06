# GoNotes Makefile - Development and Production Commands
# Default environment
ENV ?= dev

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Help target
.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)GoNotes - Available Commands$(NC)"
	@echo "================================="
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(GREEN)Environment Variables:$(NC)"
	@echo "  ENV=dev|prod    Set environment (default: dev)"
	@echo ""

# ======================
# DEVELOPMENT COMMANDS
# ======================

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@cp .env.dev .env
	@go mod download
	@go install github.com/air-verse/air@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)Development environment setup complete!$(NC)"

.PHONY: dev-db
dev-db: ## Start development database and Redis
	@echo "$(GREEN)Starting development database and Redis...$(NC)"
	@docker-compose -f docker-compose.dev.yaml up -d db redis
	@echo "$(GREEN)Development services started!$(NC)"
	@echo "$(YELLOW)Database: localhost:5432$(NC)"
	@echo "$(YELLOW)Redis: localhost:6379$(NC)"

.PHONY: dev-tools
dev-tools: ## Start development tools (Redis Commander, Adminer)
	@echo "$(GREEN)Starting development tools...$(NC)"
	@docker-compose -f docker-compose.dev.yaml up -d redis-commander adminer
	@echo "$(GREEN)Development tools started!$(NC)"
	@echo "$(YELLOW)Redis Commander: http://localhost:8081$(NC)"
	@echo "$(YELLOW)Adminer: http://localhost:8082$(NC)"

.PHONY: dev-migrate
dev-migrate: ## Run database migrations for development
	@echo "$(GREEN)Running database migrations...$(NC)"
	@migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/gonotes_dev?sslmode=disable" up

.PHONY: dev-migrate-down
dev-migrate-down: ## Rollback database migrations for development
	@echo "$(YELLOW)Rolling back database migrations...$(NC)"
	@migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/gonotes_dev?sslmode=disable" down

.PHONY: dev-run
dev-run: ## Run application in development mode with hot reload
	@echo "$(GREEN)Starting application in development mode...$(NC)"
	@air

.PHONY: dev-build
dev-build: ## Build development Docker image
	@echo "$(GREEN)Building development Docker image...$(NC)"
	@docker build -f Dockerfile.dev -t gonotes:dev .
	@echo "$(GREEN)Development image built successfully!$(NC)"

.PHONY: dev-app
dev-app: ## Start development application in container
	@echo "$(GREEN)Starting development application in container...$(NC)"
	@docker-compose -f docker-compose.dev.yaml up -d app
	@echo "$(GREEN)Development application started!$(NC)"

.PHONY: dev-start
dev-start: dev-db dev-migrate dev-run ## Start complete development environment

.PHONY: dev-stop
dev-stop: ## Stop development services
	@echo "$(YELLOW)Stopping development services...$(NC)"
	@docker-compose -f docker-compose.dev.yaml down

.PHONY: dev-clean
dev-clean: ## Clean development environment (remove volumes)
	@echo "$(RED)Cleaning development environment...$(NC)"
	@docker-compose -f docker-compose.dev.yaml down -v
	@docker system prune -f

# ======================
# TESTING COMMANDS
# ======================

.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v ./internal/service/...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	@./scripts/test_integration_docker.sh

.PHONY: test-api
test-api: ## Run API tests
	@echo "$(GREEN)Running API tests...$(NC)"
	@./scripts/test_api.sh

.PHONY: test-notes
test-notes: ## Run notes API tests
	@echo "$(GREEN)Running notes API tests...$(NC)"
	@./scripts/test_notes_api.sh

.PHONY: test-profile
test-profile: ## Run profile API tests
	@echo "$(GREEN)Running profile API tests...$(NC)"
	@./scripts/test_profile_api.sh

.PHONY: test-session
test-session: ## Run session security tests
	@echo "$(GREEN)Running session security tests...$(NC)"
	@./scripts/test_session_security_api.sh

.PHONY: test-performance
test-performance: ## Run comprehensive performance tests
	@echo "$(GREEN)Running comprehensive performance tests...$(NC)"
	@./scripts/performance_test.sh

.PHONY: test-perf-quick
test-perf-quick: ## Run quick performance test
	@echo "$(GREEN)Running quick performance test...$(NC)"
	@./scripts/quick_performance_test.sh

.PHONY: test-perf-stress
test-perf-stress: ## Run stress performance test
	@echo "$(GREEN)Running stress performance test...$(NC)"
	@./scripts/performance_test.sh --stress

.PHONY: test-all
test-all: test-unit test-integration test-api ## Run all tests

# ======================
# PRODUCTION COMMANDS
# ======================

.PHONY: prod-setup
prod-setup: ## Setup production environment
	@echo "$(GREEN)Setting up production environment...$(NC)"
	@cp .env.prod .env
	@echo "$(YELLOW)Please update .env.prod with your production values!$(NC)"
	@echo "$(GREEN)Production environment setup complete!$(NC)"

.PHONY: prod-build
prod-build: ## Build production Docker image
	@echo "$(GREEN)Building production Docker image...$(NC)"
	@docker build -f Dockerfile.prod -t gonotes:latest .
	@echo "$(GREEN)Production image built successfully!$(NC)"

.PHONY: prod-start
prod-start: ## Start production services
	@echo "$(GREEN)Starting production services...$(NC)"
	@docker-compose -f docker-compose.prod.yaml up -d
	@echo "$(GREEN)Production services started!$(NC)"

.PHONY: prod-stop
prod-stop: ## Stop production services
	@echo "$(YELLOW)Stopping production services...$(NC)"
	@docker-compose -f docker-compose.prod.yaml down

.PHONY: prod-logs
prod-logs: ## View production logs
	@docker-compose -f docker-compose.prod.yaml logs -f

.PHONY: prod-status
prod-status: ## Check production services status
	@docker-compose -f docker-compose.prod.yaml ps

.PHONY: prod-backup
prod-backup: ## Backup production database
	@echo "$(GREEN)Creating database backup...$(NC)"
	@mkdir -p backups
	@docker exec gonotes_db_prod pg_dump -U postgres gonotes > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Database backup created!$(NC)"

.PHONY: prod-restore
prod-restore: ## Restore production database (requires BACKUP_FILE variable)
	@echo "$(GREEN)Restoring database from backup...$(NC)"
	@docker exec -i gonotes_db_prod psql -U postgres gonotes < $(BACKUP_FILE)
	@echo "$(GREEN)Database restored successfully!$(NC)"

.PHONY: prod-migrate
prod-migrate: ## Run database migrations for production
	@echo "$(GREEN)Running production database migrations...$(NC)"
	@docker exec gonotes_app migrate -path ./migrations -database "$(MIGRATE_URL)" up

.PHONY: prod-deploy
prod-deploy: prod-build prod-start prod-migrate ## Deploy to production

.PHONY: prod-enhanced
prod-enhanced: ## Deploy enhanced production setup with monitoring
	@echo "$(GREEN)Deploying enhanced production setup...$(NC)"
	@cp .env.prod.enhanced .env.prod
	@echo "$(YELLOW)Please update .env.prod with your production values!$(NC)"
	@docker-compose -f docker-compose.prod.yaml up -d
	@echo "$(GREEN)Enhanced production services started!$(NC)"
	@echo "$(YELLOW)Services available:$(NC)"
	@echo "$(YELLOW)  - Application: http://localhost:8080$(NC)"
	@echo "$(YELLOW)  - Prometheus: http://localhost:9090$(NC)"
	@echo "$(YELLOW)  - Grafana: http://localhost:3000$(NC)"
	@echo "$(YELLOW)  - Nginx: http://localhost$(NC)"

.PHONY: prod-monitoring
prod-monitoring: ## Start only monitoring services
	@echo "$(GREEN)Starting monitoring services...$(NC)"
	@docker-compose -f docker-compose.prod.yaml up -d prometheus grafana loki promtail
	@echo "$(GREEN)Monitoring services started!$(NC)"

.PHONY: prod-backup-manual
prod-backup-manual: ## Run manual database backup
	@echo "$(GREEN)Running manual database backup...$(NC)"
	@docker-compose -f docker-compose.prod.yaml exec backup /backup.sh
	@echo "$(GREEN)Manual backup completed!$(NC)"

.PHONY: prod-health
prod-health: ## Check health of all production services
	@echo "$(GREEN)Checking production services health...$(NC)"
	@docker-compose -f docker-compose.prod.yaml exec app curl -f http://localhost:8080/health || echo "$(RED)App health check failed$(NC)"
	@docker-compose -f docker-compose.prod.yaml exec nginx curl -f http://localhost/health || echo "$(RED)Nginx health check failed$(NC)"
	@docker-compose -f docker-compose.prod.yaml exec prometheus curl -f http://localhost:9090/-/healthy || echo "$(RED)Prometheus health check failed$(NC)"

.PHONY: prod-logs-all
prod-logs-all: ## View logs from all production services
	@docker-compose -f docker-compose.prod.yaml logs -f

.PHONY: prod-scale
prod-scale: ## Scale production services (APP_REPLICAS=3)
	@echo "$(GREEN)Scaling production services...$(NC)"
	@docker-compose -f docker-compose.prod.yaml up -d --scale app=$(or $(APP_REPLICAS),2)
	@echo "$(GREEN)Services scaled successfully!$(NC)"

# ======================
# UTILITY COMMANDS
# ======================

.PHONY: build
build: ## Build the application binary
	@echo "$(GREEN)Building application...$(NC)"
	@go build -o bin/gonotes cmd/main.go
	@echo "$(GREEN)Build complete: bin/gonotes$(NC)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@go clean

.PHONY: cleanup
cleanup: ## Comprehensive project cleanup
	@echo "$(GREEN)Running comprehensive cleanup...$(NC)"
	@chmod +x scripts/cleanup.sh
	@./scripts/cleanup.sh
	@echo "$(GREEN)Cleanup completed!$(NC)"

.PHONY: optimize
optimize: ## Optimize project (build, docker, security, performance)
	@echo "$(GREEN)Running comprehensive optimization...$(NC)"
	@chmod +x scripts/optimize.sh
	@./scripts/optimize.sh
	@echo "$(GREEN)Optimization completed!$(NC)"

.PHONY: cleanup-temp
cleanup-temp: ## Quick cleanup (temp files only)
	@echo "$(GREEN)Cleaning temporary files...$(NC)"
	@chmod +x scripts/cleanup.sh
	@./scripts/cleanup.sh temp
	@echo "$(GREEN)Temporary files cleanup completed!$(NC)"

.PHONY: optimize-security
optimize-security: ## Security optimization
	@echo "$(GREEN)Running security optimization...$(NC)"
	@chmod +x scripts/optimize.sh
	@./scripts/optimize.sh security
	@echo "$(GREEN)Security optimization completed!$(NC)"

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy

.PHONY: fmt
fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@golangci-lint run

.PHONY: check
check: fmt vet lint ## Run all code checks

.PHONY: scripts-executable
scripts-executable: ## Make all scripts executable
	@chmod +x scripts/*.sh

.PHONY: logs
logs: ## View application logs
	@docker-compose -f docker-compose.$(ENV).yaml logs -f app

.PHONY: shell
shell: ## Open shell in application container
	@docker-compose -f docker-compose.$(ENV).yaml exec app sh

.PHONY: db-shell
db-shell: ## Open database shell
	@docker-compose -f docker-compose.$(ENV).yaml exec db psql -U postgres -d gonotes$(if $(filter dev,$(ENV)),_dev,)

.PHONY: redis-shell
redis-shell: ## Open Redis shell
	@docker-compose -f docker-compose.$(ENV).yaml exec redis redis-cli

# ======================
# MONITORING COMMANDS
# ======================

.PHONY: health
health: ## Check application health
	@curl -f http://localhost:8080/health || echo "$(RED)Application is not healthy$(NC)"

.PHONY: stats
stats: ## Show container resource usage
	@docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.PIDs}}"

# ======================
# SSL/TLS COMMANDS
# ======================

.PHONY: ssl-setup
ssl-setup: ## Setup SSL/TLS with Let's Encrypt (requires domain and email)
	@echo "$(GREEN)Setting up SSL/TLS...$(NC)"
	@if [ -z "$(DOMAIN)" ] || [ -z "$(EMAIL)" ]; then \
		echo "$(RED)Usage: make ssl-setup DOMAIN=yourdomain.com EMAIL=your-email@example.com$(NC)"; \
		exit 1; \
	fi
	@chmod +x scripts/setup_ssl.sh
	@./scripts/setup_ssl.sh $(DOMAIN) $(EMAIL)
	@echo "$(GREEN)SSL/TLS setup complete!$(NC)"

.PHONY: ssl-renew
ssl-renew: ## Renew SSL certificates
	@echo "$(GREEN)Renewing SSL certificates...$(NC)"
	@chmod +x scripts/renew_ssl.sh
	@./scripts/renew_ssl.sh
	@echo "$(GREEN)SSL certificates renewed!$(NC)"

.PHONY: ssl-check
ssl-check: ## Check SSL certificate status
	@echo "$(GREEN)Checking SSL certificate...$(NC)"
	@chmod +x scripts/check_ssl.sh
	@./scripts/check_ssl.sh $(if $(DOMAIN),$(DOMAIN),yourdomain.com)

.PHONY: ssl-test
ssl-test: ## Test SSL configuration online
	@echo "$(GREEN)Testing SSL configuration...$(NC)"
	@echo "$(YELLOW)SSL Labs: https://www.ssllabs.com/ssltest/analyze.html?d=$(if $(DOMAIN),$(DOMAIN),yourdomain.com)$(NC)"
	@echo "$(YELLOW)Security Headers: https://securityheaders.com/?q=$(if $(DOMAIN),$(DOMAIN),yourdomain.com)$(NC)"

.PHONY: ssl-start
ssl-start: ## Start production services with SSL
	@echo "$(GREEN)Starting production services with SSL...$(NC)"
	@docker-compose -f docker-compose.ssl.yaml up -d
	@echo "$(GREEN)Production services with SSL started!$(NC)"

.PHONY: ssl-stop
ssl-stop: ## Stop production services with SSL
	@echo "$(YELLOW)Stopping production services with SSL...$(NC)"
	@docker-compose -f docker-compose.ssl.yaml down

.PHONY: ssl-logs
ssl-logs: ## View SSL-enabled production logs
	@docker-compose -f docker-compose.ssl.yaml logs -f

# Default target
.DEFAULT_GOAL := help 