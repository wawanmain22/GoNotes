#!/bin/bash

# GoNotes Project Cleanup Script
# Comprehensive cleanup for development and production environments

set -e

echo "🧹 GoNotes Project Cleanup"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to print colored output
print_status() {
    local level=$1
    local message=$2
    
    case $level in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message"
            ;;
    esac
}

# Function to cleanup Docker resources
cleanup_docker() {
    print_status "INFO" "Cleaning up Docker resources..."
    
    # Stop all containers
    docker-compose -f docker-compose.dev.yaml down 2>/dev/null || true
    docker-compose -f docker-compose.prod.yaml down 2>/dev/null || true
    
    # Remove development and production images
    docker rmi gonotes:dev 2>/dev/null || true
    docker rmi gonotes:latest 2>/dev/null || true
    docker rmi gonotes:optimized 2>/dev/null || true
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes
    docker volume prune -f
    
    # Remove unused networks
    docker network prune -f
    
    print_status "SUCCESS" "Docker cleanup completed"
}

# Function to cleanup Go build artifacts
cleanup_go() {
    print_status "INFO" "Cleaning up Go build artifacts..."
    
    # Clean Go module cache (optional)
    go clean -modcache -cache 2>/dev/null || true
    
    # Remove binary files
    rm -f ./gonotes
    rm -f ./main
    rm -rf ./bin/
    
    # Clean test cache
    go clean -testcache
    
    print_status "SUCCESS" "Go cleanup completed"
}

# Function to cleanup temporary files
cleanup_temp_files() {
    print_status "INFO" "Cleaning up temporary files..."
    
    # Remove common temporary files
    find . -name "*.log" -not -path "./logs/*" -not -path "./backups/*" -delete 2>/dev/null || true
    find . -name "*.tmp" -delete 2>/dev/null || true
    find . -name "*.cache" -delete 2>/dev/null || true
    find . -name ".DS_Store" -delete 2>/dev/null || true
    find . -name "Thumbs.db" -delete 2>/dev/null || true
    
    # Clean up tmp directory
    if [ -d "./tmp" ]; then
        rm -rf ./tmp/*
    fi
    
    # Clean up test results
    rm -rf ./performance_results/temp_results_*.txt 2>/dev/null || true
    
    print_status "SUCCESS" "Temporary files cleanup completed"
}

# Function to cleanup logs (keep recent ones)
cleanup_logs() {
    print_status "INFO" "Cleaning up old log files..."
    
    # Keep only last 5 log files in each directory
    if [ -d "./logs" ]; then
        find ./logs -name "*.log" -type f | sort -r | tail -n +6 | xargs rm -f 2>/dev/null || true
    fi
    
    if [ -d "./backups" ]; then
        # Keep only last 10 backup files
        find ./backups -name "gonotes_backup_*.sql.gz" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    fi
    
    print_status "SUCCESS" "Log cleanup completed"
}

# Function to optimize dependencies
optimize_dependencies() {
    print_status "INFO" "Optimizing Go dependencies..."
    
    # Tidy up go.mod
    go mod tidy
    
    # Verify dependencies
    go mod verify
    
    # Download missing dependencies
    go mod download
    
    print_status "SUCCESS" "Dependencies optimization completed"
}

# Function to security check
security_check() {
    print_status "INFO" "Running basic security checks..."
    
    # Check for sensitive files that shouldn't be committed
    local sensitive_files=(
        ".env"
        "*.pem"
        "*.key"
        "*.crt"
        "*password*"
        "*secret*"
    )
    
    for pattern in "${sensitive_files[@]}"; do
        if find . -name "$pattern" -not -path "./docs/*" -not -path "./.env.example" -not -path "./.env.dev" -not -path "./.env.prod" -not -path "./.env.prod.enhanced" | grep -q .; then
            print_status "WARNING" "Found potentially sensitive files matching: $pattern"
        fi
    done
    
    # Check for default passwords in config files
    if grep -r "password123\|admin123\|secret123" . --exclude-dir=.git --exclude-dir=docs 2>/dev/null | grep -v "CHANGE_THIS" | grep -v "example" | grep -v "#"; then
        print_status "WARNING" "Found default passwords in configuration files"
    fi
    
    print_status "SUCCESS" "Security check completed"
}

# Function to generate cleanup report
generate_report() {
    print_status "INFO" "Generating cleanup report..."
    
    local report_file="cleanup_report_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$report_file" << EOF
GoNotes Cleanup Report
=====================
Date: $(date)
User: $(whoami)
Directory: $(pwd)

File Count Summary:
- Go files: $(find . -name "*.go" | wc -l)
- Docker files: $(find . -name "docker-compose*.yaml" -o -name "Dockerfile*" | wc -l)
- Scripts: $(find ./scripts -name "*.sh" | wc -l)
- Documentation: $(find ./docs -name "*.md" | wc -l)
- Environment files: $(find . -name ".env*" | wc -l)

Docker Images:
$(docker images | grep gonotes || echo "No GoNotes images found")

Docker Volumes:
$(docker volume ls | grep gonotes || echo "No GoNotes volumes found")

Go Module Status:
$(go list -m all | head -5)

Project Size:
$(du -sh . 2>/dev/null || echo "Size calculation failed")

Cleanup Actions Performed:
- Docker resource cleanup
- Go build artifacts cleanup
- Temporary files cleanup
- Log files cleanup (kept recent)
- Dependencies optimization
- Security checks

EOF
    
    print_status "SUCCESS" "Cleanup report saved to: $report_file"
}

# Main cleanup function
main() {
    print_status "INFO" "Starting comprehensive cleanup..."
    
    # Parse command line arguments
    local cleanup_type="${1:-full}"
    
    case $cleanup_type in
        "docker")
            cleanup_docker
            ;;
        "go")
            cleanup_go
            ;;
        "temp")
            cleanup_temp_files
            ;;
        "logs")
            cleanup_logs
            ;;
        "deps")
            optimize_dependencies
            ;;
        "security")
            security_check
            ;;
        "full")
            cleanup_docker
            cleanup_go
            cleanup_temp_files
            cleanup_logs
            optimize_dependencies
            security_check
            ;;
        *)
            print_status "ERROR" "Unknown cleanup type: $cleanup_type"
            echo "Usage: $0 [docker|go|temp|logs|deps|security|full]"
            exit 1
            ;;
    esac
    
    # Always generate report for full cleanup
    if [ "$cleanup_type" = "full" ]; then
        generate_report
    fi
    
    print_status "SUCCESS" "Cleanup completed successfully!"
    
    # Show disk space saved (approximate)
    echo ""
    print_status "INFO" "Current project size: $(du -sh . 2>/dev/null | cut -f1)"
}

# Help function
show_help() {
    echo "GoNotes Cleanup Script"
    echo ""
    echo "Usage: $0 [option]"
    echo ""
    echo "Options:"
    echo "  docker     Clean Docker resources (containers, images, volumes)"
    echo "  go         Clean Go build artifacts and cache"
    echo "  temp       Clean temporary files and caches"
    echo "  logs       Clean old log files (keep recent)"
    echo "  deps       Optimize Go dependencies"
    echo "  security   Run basic security checks"
    echo "  full       Run all cleanup operations (default)"
    echo "  help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                # Full cleanup"
    echo "  $0 docker         # Clean only Docker resources"
    echo "  $0 temp           # Clean only temporary files"
}

# Handle command line arguments
case "${1:-}" in
    "--help"|"-h"|"help")
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac 