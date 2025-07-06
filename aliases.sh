#!/bin/bash

# GoNotes Development Aliases
# Source this file in your shell profile (.bashrc, .zshrc, etc.)
# Usage: source aliases.sh

# Colors for output
export GONOTES_GREEN='\033[0;32m'
export GONOTES_YELLOW='\033[1;33m'
export GONOTES_RED='\033[0;31m'
export GONOTES_NC='\033[0m' # No Color

# Development aliases
alias gn-dev-start='make dev-start'
alias gn-dev-stop='make dev-stop'
alias gn-dev-clean='make dev-clean'
alias gn-dev-db='make dev-db'
alias gn-dev-tools='make dev-tools'
alias gn-dev-migrate='make dev-migrate'
alias gn-dev-run='make dev-run'
alias gn-dev-build='make dev-build'
alias gn-dev-app='make dev-app'

# Testing aliases
alias gn-test='make test-unit'
alias gn-test-all='make test-all'
alias gn-test-api='make test-api'
alias gn-test-notes='make test-notes'
alias gn-test-profile='make test-profile'
alias gn-test-session='make test-session'
alias gn-test-integration='make test-integration'
alias gn-test-performance='make test-performance'
alias gn-test-perf='make test-perf-quick'
alias gn-test-stress='make test-perf-stress'

# Production aliases
alias gn-prod-build='make prod-build'
alias gn-prod-start='make prod-start'
alias gn-prod-stop='make prod-stop'
alias gn-prod-deploy='make prod-deploy'
alias gn-prod-enhanced='make prod-enhanced'
alias gn-prod-monitoring='make prod-monitoring'
alias gn-prod-backup='make prod-backup'
alias gn-prod-backup-manual='make prod-backup-manual'
alias gn-prod-health='make prod-health'
alias gn-prod-logs='make prod-logs'
alias gn-prod-logs-all='make prod-logs-all'
alias gn-prod-status='make prod-status'
alias gn-prod-scale='make prod-scale'

# SSL/TLS aliases
alias gn-ssl-setup='make ssl-setup'
alias gn-ssl-renew='make ssl-renew'
alias gn-ssl-check='make ssl-check'
alias gn-ssl-test='make ssl-test'
alias gn-ssl-start='make ssl-start'
alias gn-ssl-stop='make ssl-stop'
alias gn-ssl-logs='make ssl-logs'

# Utility aliases
alias gn-build='make build'
alias gn-clean='make clean'
alias gn-cleanup='make cleanup'
alias gn-optimize='make optimize'
alias gn-cleanup-temp='make cleanup-temp'
alias gn-optimize-security='make optimize-security'
alias gn-fmt='make fmt'
alias gn-check='make check'
alias gn-deps='make deps'
alias gn-health='make health'
alias gn-stats='make stats'

# Shell access aliases
alias gn-shell='make shell'
alias gn-db-shell='make db-shell'
alias gn-redis-shell='make redis-shell'

# Log aliases
alias gn-logs='make logs'
alias gn-logs-dev='make logs ENV=dev'
alias gn-logs-prod='make logs ENV=prod'

# Quick navigation
alias gn-scripts='cd scripts'
alias gn-docs='cd docs'
alias gn-migrations='cd migrations'

# Help functions
gn-help() {
    echo -e "${GONOTES_GREEN}GoNotes Development Aliases${GONOTES_NC}"
    echo "================================="
    echo ""
    echo -e "${GONOTES_YELLOW}Development:${GONOTES_NC}"
    echo "  gn-dev-start      # Start development environment"
    echo "  gn-dev-stop       # Stop development services"
    echo "  gn-dev-clean      # Clean development environment"
    echo "  gn-dev-db         # Start database and Redis"
    echo "  gn-dev-tools      # Start development tools"
    echo "  gn-dev-migrate    # Run database migrations"
    echo "  gn-dev-run        # Run app with hot reload"
    echo "  gn-dev-build      # Build development Docker image"
    echo "  gn-dev-app        # Start development app in container"
    echo ""
    echo -e "${GONOTES_YELLOW}Testing:${GONOTES_NC}"
    echo "  gn-test           # Run unit tests"
    echo "  gn-test-all       # Run all tests"
    echo "  gn-test-api       # Run API tests"
    echo "  gn-test-notes     # Run notes API tests"
    echo "  gn-test-profile   # Run profile API tests"
    echo "  gn-test-session   # Run session security tests"
    echo "  gn-test-integration # Run integration tests"
    echo "  gn-test-performance # Run comprehensive performance tests"
    echo "  gn-test-perf      # Run quick performance test"
    echo "  gn-test-stress    # Run stress performance test"
    echo ""
    echo -e "${GONOTES_YELLOW}Production:${GONOTES_NC}"
    echo "  gn-prod-build     # Build production image"
    echo "  gn-prod-start     # Start production services"
    echo "  gn-prod-stop      # Stop production services"
    echo "  gn-prod-deploy    # Deploy to production"
    echo "  gn-prod-enhanced  # Deploy enhanced production with monitoring"
    echo "  gn-prod-monitoring # Start monitoring services only"
    echo "  gn-prod-backup    # Backup database"
    echo "  gn-prod-backup-manual # Run manual backup"
    echo "  gn-prod-health    # Check all services health"
    echo "  gn-prod-logs      # View production logs"
    echo "  gn-prod-logs-all  # View all services logs"
    echo "  gn-prod-status    # Check services status"
    echo "  gn-prod-scale     # Scale production services"
    echo ""
    echo -e "${GONOTES_YELLOW}SSL/TLS:${GONOTES_NC}"
    echo "  gn-ssl-setup      # Setup SSL with Let's Encrypt"
    echo "  gn-ssl-renew      # Renew SSL certificates"
    echo "  gn-ssl-check      # Check SSL certificate status"
    echo "  gn-ssl-test       # Test SSL configuration online"
    echo "  gn-ssl-start      # Start production with SSL"
    echo "  gn-ssl-stop       # Stop SSL production services"
    echo "  gn-ssl-logs       # View SSL production logs"
    echo ""
    echo -e "${GONOTES_YELLOW}Utilities:${GONOTES_NC}"
    echo "  gn-build          # Build application"
    echo "  gn-clean          # Clean build artifacts"
    echo "  gn-cleanup        # Comprehensive project cleanup"
    echo "  gn-optimize       # Optimize project (build, docker, security)"
    echo "  gn-cleanup-temp   # Clean temporary files only"
    echo "  gn-optimize-security # Security optimization"
    echo "  gn-fmt            # Format code"
    echo "  gn-check          # Run code checks"
    echo "  gn-deps           # Download dependencies"
    echo "  gn-health         # Check application health"
    echo "  gn-stats          # Show container stats"
    echo ""
    echo -e "${GONOTES_YELLOW}Shell Access:${GONOTES_NC}"
    echo "  gn-shell          # Open app container shell"
    echo "  gn-db-shell       # Open database shell"
    echo "  gn-redis-shell    # Open Redis shell"
    echo ""
    echo -e "${GONOTES_YELLOW}Logs:${GONOTES_NC}"
    echo "  gn-logs           # View application logs"
    echo "  gn-logs-dev       # View development logs"
    echo "  gn-logs-prod      # View production logs"
    echo ""
    echo -e "${GONOTES_YELLOW}Navigation:${GONOTES_NC}"
    echo "  gn-scripts        # Go to scripts directory"
    echo "  gn-docs           # Go to docs directory"
    echo "  gn-migrations     # Go to migrations directory"
    echo ""
    echo -e "${GONOTES_YELLOW}Workflow Functions:${GONOTES_NC}"
    echo "  gn-quick-start    # Quick development setup"
    echo "  gn-quick-test     # Quick test suite"
    echo "  gn-deploy         # Production deployment"
    echo "  gn-final-cleanup  # Final cleanup and optimization"
    echo "  gn-production-ready # Complete production preparation"
    echo ""
    echo -e "${GONOTES_YELLOW}Help:${GONOTES_NC}"
    echo "  gn-help           # Show this help message"
    echo "  make help         # Show detailed make help"
}

# Development workflow functions
gn-quick-start() {
    echo -e "${GONOTES_GREEN}Starting GoNotes development environment...${GONOTES_NC}"
    make dev-setup
    make dev-db
    make dev-migrate
    echo -e "${GONOTES_GREEN}Environment ready! Run 'gn-dev-run' to start the application.${GONOTES_NC}"
}

gn-quick-test() {
    echo -e "${GONOTES_GREEN}Running quick test suite...${GONOTES_NC}"
    make test-unit
    make test-api
    make test-perf-quick
    echo -e "${GONOTES_GREEN}Tests completed!${GONOTES_NC}"
}

gn-deploy() {
    echo -e "${GONOTES_GREEN}Deploying to production...${GONOTES_NC}"
    make prod-setup
    echo -e "${GONOTES_YELLOW}Please edit .env.prod with your production values${GONOTES_NC}"
    read -p "Press Enter to continue after editing .env.prod..."
    make prod-deploy
    echo -e "${GONOTES_GREEN}Deployment completed!${GONOTES_NC}"
}

gn-final-cleanup() {
    echo -e "${GONOTES_GREEN}Running final cleanup and optimization...${GONOTES_NC}"
    make cleanup
    make optimize
    echo -e "${GONOTES_GREEN}Final cleanup and optimization completed!${GONOTES_NC}"
}

gn-production-ready() {
    echo -e "${GONOTES_GREEN}Making project production ready...${GONOTES_NC}"
    make cleanup
    make optimize
    make check
    make test-all
    echo -e "${GONOTES_GREEN}Project is production ready!${GONOTES_NC}"
}

# Show welcome message
echo -e "${GONOTES_GREEN}GoNotes aliases loaded! Type 'gn-help' for available commands.${GONOTES_NC}" 