package service

import (
	"fmt"
	"time"

	"gonotes/internal/config"
	"gonotes/internal/model"
	"gonotes/internal/repository"
	"gonotes/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionService handles business logic for sessions
type SessionService struct {
	sessionRepo *repository.SessionRepository
	userRepo    *repository.UserRepository
	rdb         *redis.Client
	cfg         *config.Config
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo *repository.SessionRepository, userRepo *repository.UserRepository, rdb *redis.Client, cfg *config.Config) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		rdb:         rdb,
		cfg:         cfg,
	}
}

// CreateSession creates a new session after successful login
func (s *SessionService) CreateSession(user *model.User, userAgent, ipAddress string) (*model.AuthResponse, error) {
	// Generate access token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.FullName, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken(user.ID, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Parse refresh token to get token ID for Redis
	refreshClaims, err := utils.ValidateToken(refreshToken, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	// Store refresh token in Redis
	err = utils.SetRefreshToken(s.rdb, refreshClaims.ID, user.ID.String(), s.cfg.RefreshExpire)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token in redis: %w", err)
	}

	// Create session in database
	session := &model.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    &userAgent,
		IPAddress:    &ipAddress,
		IsValid:      true,
		CreatedAt:    time.Now(),
		ExpiresAt:    &refreshClaims.ExpiresAt.Time,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		// Cleanup Redis if database insert fails
		utils.InvalidateRefreshToken(s.rdb, refreshClaims.ID)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Return auth response
	return &model.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWTExpire.Seconds()),
	}, nil
}

// RefreshSession generates new tokens from valid refresh token
func (s *SessionService) RefreshSession(refreshToken string) (*model.AuthResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateToken(refreshToken, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Check if refresh token exists in Redis
	userIDStr, err := utils.GetUserIDFromRefreshToken(s.rdb, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	if userIDStr == "" {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// Verify refresh token in database
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found or invalid")
	}

	// Check if session is expired
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session expired")
	}

	// Get user details
	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.FullName, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Return new auth response (same refresh token)
	return &model.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken, // Keep the same refresh token
		ExpiresIn:    int64(s.cfg.JWTExpire.Seconds()),
	}, nil
}

// InvalidateSession invalidates a session (logout)
func (s *SessionService) InvalidateSession(refreshToken string) error {
	// Validate refresh token to get token ID
	claims, err := utils.ValidateToken(refreshToken, s.cfg)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	// Remove from Redis
	if err := utils.InvalidateRefreshToken(s.rdb, claims.ID); err != nil {
		return fmt.Errorf("failed to invalidate refresh token in redis: %w", err)
	}

	// Mark as invalid in database
	if err := s.sessionRepo.InvalidateByRefreshToken(refreshToken); err != nil {
		return fmt.Errorf("failed to invalidate session in database: %w", err)
	}

	return nil
}

// InvalidateAllSessions invalidates all sessions for a user
func (s *SessionService) InvalidateAllSessions(userID uuid.UUID) error {
	// Get all sessions for user
	sessions, err := s.sessionRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Invalidate each refresh token in Redis
	for _, session := range sessions {
		if claims, err := utils.ValidateToken(session.RefreshToken, s.cfg); err == nil {
			utils.InvalidateRefreshToken(s.rdb, claims.ID)
		}
	}

	// Invalidate all sessions in database
	if err := s.sessionRepo.InvalidateAllByUserID(userID); err != nil {
		return fmt.Errorf("failed to invalidate all sessions: %w", err)
	}

	return nil
}

// ValidateAccessToken validates access token and returns user claims
func (s *SessionService) ValidateAccessToken(accessToken string) (*utils.JWTClaims, error) {
	claims, err := utils.ValidateToken(accessToken, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	if claims.Type != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(userID uuid.UUID, currentRefreshToken *string) ([]*model.SessionResponse, error) {
	// Get all sessions from database
	sessions, err := s.sessionRepo.GetUserSessions(userID)
	if err != nil {
		return nil, err
	}

	// Determine current session ID if refresh token provided
	var currentSessionID *uuid.UUID
	if currentRefreshToken != nil {
		if currentSession, err := s.sessionRepo.GetByRefreshToken(*currentRefreshToken); err == nil && currentSession != nil {
			currentSessionID = &currentSession.ID
		}
	}

	// Convert to SessionResponse with device info
	result := make([]*model.SessionResponse, len(sessions))
	for i, session := range sessions {
		response := session.ToResponse(currentSessionID)

		// Add device info
		if session.UserAgent != nil {
			response.DeviceInfo = utils.ParseUserAgent(*session.UserAgent)
		}

		result[i] = response
	}

	return result, nil
}

// InvalidateSpecificSession invalidates a specific session for a user
func (s *SessionService) InvalidateSpecificSession(userID, sessionID uuid.UUID) error {
	// Get the session to be invalidated
	session, err := s.sessionRepo.GetSessionByIDAndUserID(sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found or not owned by user")
	}

	// Remove refresh token from Redis
	if claims, err := utils.ValidateToken(session.RefreshToken, s.cfg); err == nil {
		if err := utils.InvalidateRefreshToken(s.rdb, claims.ID); err != nil {
			// Log error but continue with database invalidation
			fmt.Printf("Failed to invalidate refresh token in Redis: %v\n", err)
		}
	}

	// Invalidate session in database
	if err := s.sessionRepo.InvalidateBySessionIDAndUserID(sessionID, userID); err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

// GetCurrentSessionFromToken determines current session from refresh token
func (s *SessionService) GetCurrentSessionFromToken(refreshToken string) (*uuid.UUID, error) {
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	return &session.ID, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions() error {
	return s.sessionRepo.CleanupExpiredSessions()
}
