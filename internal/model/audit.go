package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an auditable event in the system
type AuditEvent struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      *uuid.UUID `json:"user_id" db:"user_id"`
	EventType   string     `json:"event_type" db:"event_type"`
	EventAction string     `json:"event_action" db:"event_action"`
	Resource    string     `json:"resource" db:"resource"`
	ResourceID  *string    `json:"resource_id" db:"resource_id"`
	IPAddress   string     `json:"ip_address" db:"ip_address"`
	UserAgent   *string    `json:"user_agent" db:"user_agent"`
	Details     *string    `json:"details" db:"details"`
	Success     bool       `json:"success" db:"success"`
	ErrorMsg    *string    `json:"error_message" db:"error_message"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// AuditEventType constants for different types of events
const (
	// Authentication events
	EventTypeAuth     = "authentication"
	EventTypeSession  = "session"
	EventTypeUser     = "user"
	EventTypeProfile  = "profile"
	EventTypeNote     = "note"
	EventTypeSecurity = "security"
)

// AuditEventAction constants for different actions
const (
	// Authentication actions
	ActionLogin        = "login"
	ActionLogout       = "logout"
	ActionLoginFailed  = "login_failed"
	ActionRegister     = "register"
	ActionRefreshToken = "refresh_token"

	// Session actions
	ActionSessionCreate     = "session_create"
	ActionSessionInvalidate = "session_invalidate"
	ActionSessionExpire     = "session_expire"

	// User actions
	ActionUserCreate    = "user_create"
	ActionUserUpdate    = "user_update"
	ActionUserDelete    = "user_delete"
	ActionProfileView   = "profile_view"
	ActionProfileUpdate = "profile_update"

	// Note actions
	ActionNoteCreate = "note_create"
	ActionNoteView   = "note_view"
	ActionNoteUpdate = "note_update"
	ActionNoteDelete = "note_delete"

	// Security actions
	ActionRateLimitExceeded  = "rate_limit_exceeded"
	ActionDDoSDetected       = "ddos_detected"
	ActionUnauthorizedAccess = "unauthorized_access"
	ActionSuspiciousActivity = "suspicious_activity"
)

// CreateAuditEvent creates a new audit event
func CreateAuditEvent(eventType, action, resource string) *AuditEvent {
	return &AuditEvent{
		ID:          uuid.New(),
		EventType:   eventType,
		EventAction: action,
		Resource:    resource,
		Success:     true,
		CreatedAt:   time.Now(),
	}
}

// SetUser sets the user ID for the audit event
func (a *AuditEvent) SetUser(userID uuid.UUID) *AuditEvent {
	a.UserID = &userID
	return a
}

// SetResourceID sets the resource ID for the audit event
func (a *AuditEvent) SetResourceID(resourceID string) *AuditEvent {
	a.ResourceID = &resourceID
	return a
}

// SetClientInfo sets the client information
func (a *AuditEvent) SetClientInfo(ipAddress string, userAgent *string) *AuditEvent {
	a.IPAddress = ipAddress
	a.UserAgent = userAgent
	return a
}

// SetDetails sets additional details for the event
func (a *AuditEvent) SetDetails(details string) *AuditEvent {
	a.Details = &details
	return a
}

// SetError marks the event as failed with an error message
func (a *AuditEvent) SetError(errorMsg string) *AuditEvent {
	a.Success = false
	a.ErrorMsg = &errorMsg
	return a
}

// AuditLogQuery represents query parameters for audit log retrieval
type AuditLogQuery struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	EventType   *string    `json:"event_type,omitempty"`
	EventAction *string    `json:"event_action,omitempty"`
	Resource    *string    `json:"resource,omitempty"`
	IPAddress   *string    `json:"ip_address,omitempty"`
	Success     *bool      `json:"success,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
}

// SetDefaults sets default values for audit log query
func (q *AuditLogQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
}

// AuditLogResponse represents audit log API response
type AuditLogResponse struct {
	Events     []*AuditEvent `json:"events"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
	HasNext    bool          `json:"has_next"`
	HasPrev    bool          `json:"has_prev"`
}

// NewAuditLogResponse creates a new audit log response
func NewAuditLogResponse(events []*AuditEvent, total int, query *AuditLogQuery) *AuditLogResponse {
	totalPages := (total + query.PageSize - 1) / query.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &AuditLogResponse{
		Events:     events,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}
}
