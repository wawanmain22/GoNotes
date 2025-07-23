package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gonotes/internal/model"

	"github.com/google/uuid"
)

// SessionRepository handles database operations for sessions
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session in the database
func (r *SessionRepository) Create(session *model.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip_address, is_valid, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(
		query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.IsValid,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *SessionRepository) GetByRefreshToken(refreshToken string) (*model.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, is_valid, created_at, expires_at
		FROM sessions
		WHERE refresh_token = $1 AND is_valid = true
	`

	session := &model.Session{}

	err := r.db.QueryRow(query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.IsValid,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}

	return session, nil
}

// GetByUserID retrieves all valid sessions for a user
func (r *SessionRepository) GetByUserID(userID uuid.UUID) ([]model.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, is_valid, created_at, expires_at
		FROM sessions
		WHERE user_id = $1 AND is_valid = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user ID: %w", err)
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var session model.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.UserAgent,
			&session.IPAddress,
			&session.IsValid,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// InvalidateByRefreshToken marks a session as invalid by refresh token
func (r *SessionRepository) InvalidateByRefreshToken(refreshToken string) error {
	query := `
		UPDATE sessions
		SET is_valid = false
		WHERE refresh_token = $1
	`

	result, err := r.db.Exec(query, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// InvalidateBySessionID marks a session as invalid by session ID
func (r *SessionRepository) InvalidateBySessionID(sessionID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_valid = false
		WHERE id = $1
	`

	result, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// InvalidateAllByUserID marks all sessions as invalid for a user
func (r *SessionRepository) InvalidateAllByUserID(userID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_valid = false
		WHERE user_id = $1 AND is_valid = true
	`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate all sessions: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from database
func (r *SessionRepository) CleanupExpiredSessions() error {
	query := `
		DELETE FROM sessions
		WHERE expires_at < $1 OR (expires_at IS NULL AND created_at < $2)
	`

	// Remove sessions expired or older than 30 days if no expires_at
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	_, err := r.db.Exec(query, time.Now(), thirtyDaysAgo)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *SessionRepository) GetUserSessions(userID uuid.UUID) ([]model.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, 
			   is_valid, created_at, expires_at
		FROM sessions 
		WHERE user_id = $1 AND is_valid = true 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var session model.Session
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.UserAgent,
			&session.IPAddress,
			&session.IsValid,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return sessions, nil
}

// GetSessionByIDAndUserID retrieves a specific session by ID and user ID
func (r *SessionRepository) GetSessionByIDAndUserID(sessionID, userID uuid.UUID) (*model.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, 
			   is_valid, created_at, expires_at
		FROM sessions 
		WHERE id = $1 AND user_id = $2 AND is_valid = true
	`

	session := &model.Session{}
	err := r.db.QueryRow(query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.IsValid,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to get session by ID and user ID: %w", err)
	}

	return session, nil
}

// InvalidateBySessionIDAndUserID invalidates a specific session for a user
func (r *SessionRepository) InvalidateBySessionIDAndUserID(sessionID, userID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_valid = false
		WHERE id = $1 AND user_id = $2 AND is_valid = true
	`

	result, err := r.db.Exec(query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or not owned by user")
	}

	return nil
}
