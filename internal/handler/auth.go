package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"gonotes/internal/middleware"
	"gonotes/internal/model"
	"gonotes/internal/service"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *service.UserService, sessionService *service.SessionService) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		sessionService: sessionService,
	}
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// sendResponse sends a JSON response
func sendResponse(w http.ResponseWriter, code int, status string, message string, data interface{}, err interface{}) {
	response := APIResponse{
		Status:  status,
		Code:    code,
		Message: message,
		Data:    data,
		Error:   err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// extractClientInfo extracts client information from request
func extractClientInfo(r *http.Request) (userAgent, ipAddress string) {
	userAgent = r.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Get real IP address (considering proxies)
	ipAddress = r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	// Extract IP from address:port format
	if idx := strings.LastIndex(ipAddress, ":"); idx != -1 {
		ipAddress = ipAddress[:idx]
	}

	return userAgent, ipAddress
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest

	// Parse JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid JSON format", nil, err.Error())
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	// Register user
	user, err := h.userService.Register(&req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") ||
			strings.Contains(err.Error(), "email already exists") {
			sendResponse(w, http.StatusBadRequest, "error", err.Error(), nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to register user", nil, err.Error())
		return
	}

	// Return user response (without password)
	sendResponse(w, http.StatusCreated, "success", "User registered successfully", user.ToResponse(), nil)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest

	// Parse JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid JSON format", nil, err.Error())
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Authenticate user
	user, err := h.userService.Login(&req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") ||
			strings.Contains(err.Error(), "invalid email or password") {
			sendResponse(w, http.StatusUnauthorized, "error", "Invalid email or password", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Login failed", nil, err.Error())
		return
	}

	// Extract client information
	userAgent, ipAddress := extractClientInfo(r)

	// Create session with JWT tokens
	authResponse, err := h.sessionService.CreateSession(user, userAgent, ipAddress)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to create session", nil, err.Error())
		return
	}

	// Return authentication response with tokens
	sendResponse(w, http.StatusOK, "success", "Login successful", authResponse, nil)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest

	// Parse JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid JSON format", nil, err.Error())
		return
	}

	// Refresh session
	authResponse, err := h.sessionService.RefreshSession(req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "expired") ||
			strings.Contains(err.Error(), "not found") {
			sendResponse(w, http.StatusUnauthorized, "error", "Invalid or expired refresh token", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to refresh token", nil, err.Error())
		return
	}

	// Return new authentication response
	sendResponse(w, http.StatusOK, "success", "Token refreshed successfully", authResponse, nil)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest

	// Parse JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid JSON format", nil, err.Error())
		return
	}

	// Invalidate session
	err := h.sessionService.InvalidateSession(req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			sendResponse(w, http.StatusUnauthorized, "error", "Invalid refresh token", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to logout", nil, err.Error())
		return
	}

	// Return success response
	sendResponse(w, http.StatusOK, "success", "Logout successful", nil, nil)
}

// GetProfile handles GET /api/v1/user/profile request
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Get profile with caching
	profile, err := h.userService.GetProfileWithCache(userID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			sendResponse(w, http.StatusNotFound, "error", "User not found", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to get profile", nil, err.Error())
		return
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Profile retrieved successfully", profile, nil)
}

// UpdateProfile handles PUT /api/v1/user/profile request
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Parse request body
	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, "error", "Invalid JSON format", nil, err.Error())
		return
	}

	// Update profile
	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			sendResponse(w, http.StatusBadRequest, "error", err.Error(), nil, nil)
			return
		}
		if strings.Contains(err.Error(), "email already exists") {
			sendResponse(w, http.StatusConflict, "error", "Email already exists", nil, nil)
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			sendResponse(w, http.StatusNotFound, "error", "User not found", nil, nil)
			return
		}
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to update profile", nil, err.Error())
		return
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Profile updated successfully", user.ToResponse(), nil)
}

// GetSessions handles GET /api/v1/user/sessions request (legacy endpoint)
func (h *AuthHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		sendResponse(w, http.StatusUnauthorized, "error", "Authentication required", nil, nil)
		return
	}

	// Get all user sessions
	sessions, err := h.sessionService.GetUserSessions(userID, nil)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "error", "Failed to retrieve sessions", nil, err.Error())
		return
	}

	// Convert to legacy format for backward compatibility
	legacySessions := make([]map[string]interface{}, len(sessions))
	for i, session := range sessions {
		legacySessions[i] = map[string]interface{}{
			"id":         session.ID,
			"user_agent": session.UserAgent,
			"ip_address": session.IPAddress,
			"created_at": session.CreatedAt,
			"expires_at": session.ExpiresAt,
			"is_current": session.IsCurrent,
		}
	}

	// Send success response
	sendResponse(w, http.StatusOK, "success", "Sessions retrieved successfully", legacySessions, nil)
}
