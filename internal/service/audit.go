package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gonotes/internal/model"
)

// AuditService handles audit logging operations
type AuditService struct {
	logFile *os.File
}

// NewAuditService creates a new audit service
func NewAuditService() *AuditService {
	// Create audit log file (in production, this would be a proper logging system)
	logFile, err := os.OpenFile("audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open audit log file: %v", err)
		logFile = nil
	}

	return &AuditService{
		logFile: logFile,
	}
}

// LogEvent logs an audit event
func (s *AuditService) LogEvent(event *model.AuditEvent) {
	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal audit event: %v", err)
		return
	}

	// Create log entry
	timestamp := event.CreatedAt.Format(time.RFC3339)
	logEntry := fmt.Sprintf("[AUDIT] %s - %s\n", timestamp, string(eventJSON))

	// Write to console
	fmt.Print(logEntry)

	// Write to file if available
	if s.logFile != nil {
		s.logFile.WriteString(logEntry)
		s.logFile.Sync()
	}
}

// LogAuthEvent logs an authentication-related event
func (s *AuditService) LogAuthEvent(action, resource, ipAddress string, userAgent *string, userID *model.User, success bool, errorMsg *string) {
	event := model.CreateAuditEvent(model.EventTypeAuth, action, resource)
	event.SetClientInfo(ipAddress, userAgent)

	if userID != nil {
		event.SetUser(userID.ID)
	}

	if !success && errorMsg != nil {
		event.SetError(*errorMsg)
	}

	s.LogEvent(event)
}

// LogSessionEvent logs a session-related event
func (s *AuditService) LogSessionEvent(action string, userID model.User, sessionID string, ipAddress string, userAgent *string, success bool) {
	event := model.CreateAuditEvent(model.EventTypeSession, action, "session")
	event.SetUser(userID.ID)
	event.SetResourceID(sessionID)
	event.SetClientInfo(ipAddress, userAgent)

	if !success {
		event.SetError("Session operation failed")
	}

	s.LogEvent(event)
}

// LogUserEvent logs a user-related event
func (s *AuditService) LogUserEvent(action string, userID model.User, ipAddress string, userAgent *string, details *string, success bool) {
	event := model.CreateAuditEvent(model.EventTypeUser, action, "user")
	event.SetUser(userID.ID)
	event.SetResourceID(userID.ID.String())
	event.SetClientInfo(ipAddress, userAgent)

	if details != nil {
		event.SetDetails(*details)
	}

	if !success {
		event.SetError("User operation failed")
	}

	s.LogEvent(event)
}

// LogNoteEvent logs a note-related event
func (s *AuditService) LogNoteEvent(action string, userID model.User, noteID string, ipAddress string, userAgent *string, success bool) {
	event := model.CreateAuditEvent(model.EventTypeNote, action, "note")
	event.SetUser(userID.ID)
	event.SetResourceID(noteID)
	event.SetClientInfo(ipAddress, userAgent)

	if !success {
		event.SetError("Note operation failed")
	}

	s.LogEvent(event)
}

// LogSecurityEvent logs a security-related event
func (s *AuditService) LogSecurityEvent(action, details, ipAddress string, userAgent *string, userID *model.User) {
	event := model.CreateAuditEvent(model.EventTypeSecurity, action, "security")
	event.SetClientInfo(ipAddress, userAgent)
	event.SetDetails(details)

	if userID != nil {
		event.SetUser(userID.ID)
	}

	s.LogEvent(event)
}

// Close closes the audit service and its resources
func (s *AuditService) Close() {
	if s.logFile != nil {
		s.logFile.Close()
	}
}

// AuditMiddleware creates a middleware that logs audit events
func (s *AuditService) AuditMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This would integrate with the existing audit middleware
			// For now, we'll keep it simple
			next.ServeHTTP(w, r)
		})
	}
}
