package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"gonotes/internal/config"
	"gonotes/internal/service"
	"gonotes/internal/utils"

	"github.com/google/uuid"
)

// ContextKey represents keys used in context
type ContextKey string

const (
	// UserClaimsKey is the context key for user claims
	UserClaimsKey ContextKey = "user_claims"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

// AuthMiddleware handles JWT authentication for protected routes
type AuthMiddleware struct {
	sessionService *service.SessionService
	cfg            *config.Config
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(sessionService *service.SessionService, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		sessionService: sessionService,
		cfg:            cfg,
	}
}

// APIErrorResponse represents error response format
type APIErrorResponse struct {
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

// sendErrorResponse sends a JSON error response
func sendErrorResponse(w http.ResponseWriter, code int, message string, err interface{}) {
	response := APIErrorResponse{
		Status:  "error",
		Code:    code,
		Message: message,
		Error:   err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RequireAuth middleware that requires valid JWT authentication
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendErrorResponse(w, http.StatusUnauthorized, "Authorization header required", nil)
			return
		}

		// Extract Bearer token
		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			sendErrorResponse(w, http.StatusUnauthorized, "Bearer token required", nil)
			return
		}

		// Validate access token
		claims, err := am.sessionService.ValidateAccessToken(token)
		if err != nil {
			sendErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)

		// Continue to next handler with user context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware that allows but doesn't require authentication
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Extract Bearer token
		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			// Invalid token format, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate access token
		claims, err := am.sessionService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Valid token, add claims to context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims extracts user claims from request context
func GetUserClaims(r *http.Request) (*utils.JWTClaims, bool) {
	claims, ok := r.Context().Value(UserClaimsKey).(*utils.JWTClaims)
	return claims, ok
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetCurrentUser helper function to get user info from context
func GetCurrentUser(r *http.Request) (userID uuid.UUID, email string, fullName string, ok bool) {
	claims, exists := GetUserClaims(r)
	if !exists {
		return uuid.Nil, "", "", false
	}

	return claims.UserID, claims.Email, claims.FullName, true
}

// MustGetUserID extracts user ID from context or panics (for internal use)
func MustGetUserID(r *http.Request) uuid.UUID {
	userID, ok := GetUserID(r)
	if !ok {
		panic("user ID not found in context - middleware not properly applied")
	}
	return userID
}

// AdminOnly middleware that requires admin role (for future use)
// func (am *AuthMiddleware) AdminOnly(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// First ensure user is authenticated
// 		am.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			claims, ok := GetUserClaims(r)
// 			if !ok {
// 				sendErrorResponse(w, http.StatusUnauthorized, "Authentication required", nil)
// 				return
// 			}

// 			// Check if user has admin privileges
// 			// Admin functionality can be extended by:
// 			// 1. Adding role field to JWT claims during login
// 			// 2. Checking user role from database
// 			// 3. Using separate admin table/permissions

// 			// For production use, implement one of the following:
// 			// if claims.Role != "admin" {
// 			//     sendErrorResponse(w, http.StatusForbidden, "Admin access required", nil)
// 			//     return
// 			// }

// 			// Current implementation: Allow all authenticated users
// 			// Modify this section when admin roles are implemented
// 			next.ServeHTTP(w, r)
// 		})).ServeHTTP(w, r)
// 	})
// }

// CORS middleware for handling preflight requests
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Link")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
