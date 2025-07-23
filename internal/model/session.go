package model

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user session with refresh token
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	RefreshToken string     `json:"refresh_token" db:"refresh_token"`
	UserAgent    *string    `json:"user_agent" db:"user_agent"`
	IPAddress    *string    `json:"ip_address" db:"ip_address"`
	IsValid      bool       `json:"is_valid" db:"is_valid"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at" db:"expires_at"`
}

// AuthResponse represents successful authentication response
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"` // seconds
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// SessionResponse represents a session in API responses
type SessionResponse struct {
	ID         uuid.UUID   `json:"id"`
	UserAgent  *string     `json:"user_agent"`
	IPAddress  *string     `json:"ip_address"`
	IsCurrent  bool        `json:"is_current"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  *time.Time  `json:"expires_at"`
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
}

// DeviceInfo represents parsed device information
type DeviceInfo struct {
	Browser  string `json:"browser"`
	OS       string `json:"os"`
	Device   string `json:"device"`
	IsMobile bool   `json:"is_mobile"`
}

// ToResponse converts Session to SessionResponse
func (s *Session) ToResponse(currentSessionID *uuid.UUID) *SessionResponse {
	response := &SessionResponse{
		ID:        s.ID,
		UserAgent: s.UserAgent,
		IPAddress: s.IPAddress,
		IsCurrent: currentSessionID != nil && *currentSessionID == s.ID,
		CreatedAt: s.CreatedAt,
		ExpiresAt: s.ExpiresAt,
	}

	// Parse device info from user agent
	if s.UserAgent != nil {
		// Device info will be set by the caller using utils.ParseUserAgent
		// This is to avoid circular import dependencies
	}

	return response
}

// InvalidateSessionRequest represents a request to invalidate a specific session
type InvalidateSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" validate:"required"`
}
