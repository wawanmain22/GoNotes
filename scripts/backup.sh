#!/bin/bash

# GoNotes Production Database Backup Script
# Automated backup with compression, retention, and monitoring

set -euo pipefail

# Configuration
POSTGRES_HOST="${POSTGRES_HOST:-db}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD}"
POSTGRES_DB="${POSTGRES_DB:-gonotes}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
COMPRESSION_LEVEL="${COMPRESSION_LEVEL:-6}"

# Timestamp for backup filename
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="gonotes_backup_${TIMESTAMP}.sql"
COMPRESSED_FILE="${BACKUP_FILE}.gz"

# Colors for logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging function
log() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} [$timestamp] $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} [$timestamp] $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}[WARNING]${NC} [$timestamp] $message"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} [$timestamp] $message"
            ;;
    esac
    
    # Also log to file if backup directory exists
    if [[ -d "$BACKUP_DIR" ]]; then
        echo "[$timestamp] [$level] $message" >> "$BACKUP_DIR/backup.log"
    fi
}

# Error handling
handle_error() {
    local exit_code=$?
    log "ERROR" "Backup failed with exit code $exit_code"
    
    # Cleanup incomplete backup file
    if [[ -f "$BACKUP_DIR/$BACKUP_FILE" ]]; then
        rm -f "$BACKUP_DIR/$BACKUP_FILE"
        log "INFO" "Cleaned up incomplete backup file"
    fi
    
    if [[ -f "$BACKUP_DIR/$COMPRESSED_FILE" ]]; then
        rm -f "$BACKUP_DIR/$COMPRESSED_FILE"
        log "INFO" "Cleaned up incomplete compressed backup file"
    fi
    
    exit $exit_code
}

trap handle_error ERR

# Check prerequisites
check_prerequisites() {
    log "INFO" "Checking prerequisites..."
    
    # Check if backup directory exists
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log "INFO" "Creating backup directory: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
    
    # Check if pg_dump is available
    if ! command -v pg_dump &> /dev/null; then
        log "ERROR" "pg_dump is not available"
        exit 1
    fi
    
    # Check if gzip is available
    if ! command -v gzip &> /dev/null; then
        log "ERROR" "gzip is not available"
        exit 1
    fi
    
    # Check database connectivity
    log "INFO" "Testing database connectivity..."
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    if ! pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" &> /dev/null; then
        log "ERROR" "Cannot connect to database"
        exit 1
    fi
    
    log "SUCCESS" "Prerequisites check passed"
}

# Get database size
get_database_size() {
    local size=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT pg_size_pretty(pg_database_size('$POSTGRES_DB'));" 2>/dev/null | xargs)
    echo "$size"
}

# Perform database backup
perform_backup() {
    local db_size=$(get_database_size)
    log "INFO" "Starting backup of database '$POSTGRES_DB' (Size: $db_size)"
    log "INFO" "Backup file: $COMPRESSED_FILE"
    
    local start_time=$(date +%s)
    
    # Create backup with compression pipeline
    pg_dump -h "$POSTGRES_HOST" \
            -p "$POSTGRES_PORT" \
            -U "$POSTGRES_USER" \
            -d "$POSTGRES_DB" \
            --verbose \
            --no-password \
            --format=plain \
            --no-owner \
            --no-privileges \
            --column-inserts \
            --disable-triggers | \
    gzip -"$COMPRESSION_LEVEL" > "$BACKUP_DIR/$COMPRESSED_FILE"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Get backup file size
    local backup_size=$(du -h "$BACKUP_DIR/$COMPRESSED_FILE" | cut -f1)
    
    log "SUCCESS" "Backup completed successfully"
    log "INFO" "Backup duration: ${duration}s"
    log "INFO" "Backup size: $backup_size"
    log "INFO" "Original database size: $db_size"
}

# Verify backup integrity
verify_backup() {
    log "INFO" "Verifying backup integrity..."
    
    # Check if backup file exists and is not empty
    if [[ ! -f "$BACKUP_DIR/$COMPRESSED_FILE" ]]; then
        log "ERROR" "Backup file does not exist"
        return 1
    fi
    
    if [[ ! -s "$BACKUP_DIR/$COMPRESSED_FILE" ]]; then
        log "ERROR" "Backup file is empty"
        return 1
    fi
    
    # Test gzip integrity
    if ! gzip -t "$BACKUP_DIR/$COMPRESSED_FILE" &> /dev/null; then
        log "ERROR" "Backup file is corrupted (gzip test failed)"
        return 1
    fi
    
    # Test SQL content (basic check)
    if ! zcat "$BACKUP_DIR/$COMPRESSED_FILE" | head -20 | grep -q "PostgreSQL database dump" &> /dev/null; then
        log "ERROR" "Backup file does not appear to contain valid PostgreSQL dump"
        return 1
    fi
    
    log "SUCCESS" "Backup integrity verification passed"
}

# Cleanup old backups
cleanup_old_backups() {
    log "INFO" "Cleaning up backups older than $RETENTION_DAYS days..."
    
    local deleted_count=0
    
    # Find and delete old backup files
    while IFS= read -r -d '' file; do
        local filename=$(basename "$file")
        log "INFO" "Deleting old backup: $filename"
        rm -f "$file"
        ((deleted_count++))
    done < <(find "$BACKUP_DIR" -name "gonotes_backup_*.sql.gz" -type f -mtime +$RETENTION_DAYS -print0)
    
    # Cleanup old log entries (keep only last 1000 lines)
    if [[ -f "$BACKUP_DIR/backup.log" ]]; then
        tail -1000 "$BACKUP_DIR/backup.log" > "$BACKUP_DIR/backup.log.tmp"
        mv "$BACKUP_DIR/backup.log.tmp" "$BACKUP_DIR/backup.log"
    fi
    
    if [[ $deleted_count -gt 0 ]]; then
        log "SUCCESS" "Deleted $deleted_count old backup(s)"
    else
        log "INFO" "No old backups to delete"
    fi
}

# Generate backup report
generate_report() {
    log "INFO" "Generating backup report..."
    
    local backup_count=$(find "$BACKUP_DIR" -name "gonotes_backup_*.sql.gz" -type f | wc -l)
    local total_size=$(du -sh "$BACKUP_DIR" | cut -f1)
    local latest_backup=$(ls -t "$BACKUP_DIR"/gonotes_backup_*.sql.gz 2>/dev/null | head -1 | xargs basename 2>/dev/null || echo "None")
    
    cat > "$BACKUP_DIR/backup_report.txt" << EOF
GoNotes Backup Report
====================
Date: $(date)
Status: Success
Latest Backup: $latest_backup
Total Backups: $backup_count
Total Size: $total_size
Retention: $RETENTION_DAYS days

Database Info:
- Host: $POSTGRES_HOST:$POSTGRES_PORT
- Database: $POSTGRES_DB
- User: $POSTGRES_USER

Backup Settings:
- Compression Level: $COMPRESSION_LEVEL
- Backup Directory: $BACKUP_DIR
EOF
    
    log "SUCCESS" "Backup report generated: $BACKUP_DIR/backup_report.txt"
}

# Send notification (if configured)
send_notification() {
    local status=$1
    local message=$2
    
    # Email notification (if configured)
    if [[ -n "${BACKUP_EMAIL:-}" ]] && command -v mail &> /dev/null; then
        echo "$message" | mail -s "GoNotes Backup: $status" "$BACKUP_EMAIL"
        log "INFO" "Email notification sent to $BACKUP_EMAIL"
    fi
    
    # Webhook notification (if configured)
    if [[ -n "${BACKUP_WEBHOOK:-}" ]] && command -v curl &> /dev/null; then
        curl -s -X POST "$BACKUP_WEBHOOK" \
             -H "Content-Type: application/json" \
             -d "{\"status\":\"$status\",\"message\":\"$message\",\"timestamp\":\"$(date -Iseconds)\"}" \
             > /dev/null
        log "INFO" "Webhook notification sent"
    fi
    
    # Slack notification (if configured)
    if [[ -n "${SLACK_WEBHOOK:-}" ]] && command -v curl &> /dev/null; then
        local color="good"
        if [[ "$status" != "Success" ]]; then
            color="danger"
        fi
        
        curl -s -X POST "$SLACK_WEBHOOK" \
             -H "Content-Type: application/json" \
             -d "{\"attachments\":[{\"color\":\"$color\",\"title\":\"GoNotes Backup: $status\",\"text\":\"$message\",\"ts\":$(date +%s)}]}" \
             > /dev/null
        log "INFO" "Slack notification sent"
    fi
}

# Main backup function
main() {
    log "INFO" "=== GoNotes Database Backup Started ==="
    
    local start_time=$(date +%s)
    
    # Run backup process
    check_prerequisites
    perform_backup
    verify_backup
    cleanup_old_backups
    generate_report
    
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    local success_message="Backup completed successfully in ${total_duration}s. File: $COMPRESSED_FILE"
    log "SUCCESS" "$success_message"
    log "INFO" "=== GoNotes Database Backup Completed ==="
    
    # Send success notification
    send_notification "Success" "$success_message"
}

# Error handling for main function
if ! main "$@"; then
    local error_message="Backup failed. Check logs for details."
    log "ERROR" "$error_message"
    send_notification "Failed" "$error_message"
    exit 1
fi 