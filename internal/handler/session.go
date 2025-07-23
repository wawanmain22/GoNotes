package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"gonotes/internal/middleware"
	"gonotes/internal/model"
	"gonotes/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SessionHandler handles session management requests
type SessionHandler struct {
	sessionService *service.SessionService
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// GetActiveSessions handles GET /api/v1/user/sessions/active
func (h *SessionHandler) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Get current session info from Authorization header (optional)
	var currentRefreshToken *string
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// This will be used to identify the current session
		// We need to get the refresh token somehow, but for now we'll skip this
	}

	// Get all sessions for user
	sessions, err := h.sessionService.GetUserSessions(userID, currentRefreshToken)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get user sessions", nil, err.Error())
		return
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Sessions retrieved successfully", sessions, nil)
}

// InvalidateSession handles DELETE /api/v1/user/sessions/{sessionId}
func (h *SessionHandler) InvalidateSession(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Get session ID from URL parameter
	sessionIDStr := chi.URLParam(r, "sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid session ID format", nil, err.Error())
		return
	}

	// Invalidate the specific session
	if err := h.sessionService.InvalidateSpecificSession(userID, sessionID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			sendResponse(w, http.StatusNotFound, "error", "Session not found", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to invalidate session", nil, err.Error())
		return
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Session invalidated successfully", nil, nil)
}

// InvalidateAllSessions handles DELETE /api/v1/user/sessions
func (h *SessionHandler) InvalidateAllSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Invalidate all sessions for user
	if err := h.sessionService.InvalidateAllSessions(userID); err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to invalidate all sessions", nil, err.Error())
		return
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "All sessions invalidated successfully", nil, nil)
}

// InvalidateSessionByRequest handles POST /api/v1/user/sessions/invalidate
func (h *SessionHandler) InvalidateSessionByRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Parse request body
	var req model.InvalidateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid JSON format", nil, err.Error())
		return
	}

	// Validate request
	if req.SessionID == uuid.Nil {
		sendResponse(w, http.StatusBadRequest, "error", "Session ID is required", nil, nil)
		return
	}

	// Invalidate the specific session
	if err := h.sessionService.InvalidateSpecificSession(userID, req.SessionID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			sendResponse(w, http.StatusNotFound, "error", "Session not found", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to invalidate session", nil, err.Error())
		return
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Session invalidated successfully", nil, nil)
}

// GetSessionsStats handles GET /api/v1/user/sessions/stats
func (h *SessionHandler) GetSessionsStats(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Get all sessions for user
	sessions, err := h.sessionService.GetUserSessions(userID, nil)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get user sessions", nil, err.Error())
		return
	}

	// Calculate statistics
	stats := calculateSessionStats(sessions)

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Session statistics retrieved successfully", stats, nil)
}

// calculateSessionStats calculates session statistics from session list
func calculateSessionStats(sessions []*model.SessionResponse) map[string]interface{} {
	stats := map[string]interface{}{
		"total_sessions":    len(sessions),
		"desktop_sessions":  0,
		"mobile_sessions":   0,
		"browser_breakdown": make(map[string]int),
		"os_breakdown":      make(map[string]int),
		"device_breakdown":  make(map[string]int),
		"current_session":   false,
	}

	for _, session := range sessions {
		// Count current session
		if session.IsCurrent {
			stats["current_session"] = true
		}

		// Count mobile/desktop
		if session.DeviceInfo != nil {
			if session.DeviceInfo.IsMobile {
				stats["mobile_sessions"] = stats["mobile_sessions"].(int) + 1
			} else {
				stats["desktop_sessions"] = stats["desktop_sessions"].(int) + 1
			}

			// Browser breakdown
			browserBreakdown := stats["browser_breakdown"].(map[string]int)
			browserBreakdown[session.DeviceInfo.Browser]++

			// OS breakdown
			osBreakdown := stats["os_breakdown"].(map[string]int)
			osBreakdown[session.DeviceInfo.OS]++

			// Device breakdown
			deviceBreakdown := stats["device_breakdown"].(map[string]int)
			deviceBreakdown[session.DeviceInfo.Device]++
		}
	}

	return stats
}
