package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gonotes/internal/model"

	"github.com/google/uuid"
)

// NoteRepository handles database operations for notes
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository creates a new note repository
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{
		db: db,
	}
}

// Create creates a new note
func (r *NoteRepository) Create(note *model.Note) error {
	query := `
		INSERT INTO notes (id, user_id, title, content, status, tags, is_public, view_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now

	_, err := r.db.Exec(query,
		note.ID,
		note.UserID,
		note.Title,
		note.Content,
		note.Status,
		note.Tags,
		note.IsPublic,
		note.ViewCount,
		note.CreatedAt,
		note.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

// GetByID retrieves a note by ID
func (r *NoteRepository) GetByID(id uuid.UUID) (*model.Note, error) {
	query := `
		SELECT id, user_id, title, content, status, tags, is_public, view_count, 
			   created_at, updated_at, deleted_at
		FROM notes 
		WHERE id = $1
	`

	var note model.Note
	err := r.db.QueryRow(query, id).Scan(
		&note.ID,
		&note.UserID,
		&note.Title,
		&note.Content,
		&note.Status,
		&note.Tags,
		&note.IsPublic,
		&note.ViewCount,
		&note.CreatedAt,
		&note.UpdatedAt,
		&note.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note by ID: %w", err)
	}

	return &note, nil
}

// GetByIDAndUserID retrieves a note by ID and user ID (for security)
func (r *NoteRepository) GetByIDAndUserID(id, userID uuid.UUID) (*model.Note, error) {
	query := `
		SELECT id, user_id, title, content, status, tags, is_public, view_count, 
			   created_at, updated_at, deleted_at
		FROM notes 
		WHERE id = $1 AND user_id = $2
	`

	var note model.Note
	err := r.db.QueryRow(query, id, userID).Scan(
		&note.ID,
		&note.UserID,
		&note.Title,
		&note.Content,
		&note.Status,
		&note.Tags,
		&note.IsPublic,
		&note.ViewCount,
		&note.CreatedAt,
		&note.UpdatedAt,
		&note.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note by ID and user ID: %w", err)
	}

	return &note, nil
}

// Update updates an existing note
func (r *NoteRepository) Update(note *model.Note) error {
	query := `
		UPDATE notes 
		SET title = $2, content = $3, status = $4, tags = $5, is_public = $6, 
			updated_at = $7, deleted_at = $8
		WHERE id = $1 AND user_id = $9
	`

	note.UpdatedAt = time.Now()

	result, err := r.db.Exec(query,
		note.ID,
		note.Title,
		note.Content,
		note.Status,
		note.Tags,
		note.IsPublic,
		note.UpdatedAt,
		note.DeletedAt,
		note.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note not found or no permission to update")
	}

	return nil
}

// Delete soft deletes a note (sets status to deleted and deleted_at timestamp)
func (r *NoteRepository) Delete(id, userID uuid.UUID) error {
	query := `
		UPDATE notes 
		SET status = 'deleted', deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status != 'deleted'
	`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note not found or already deleted")
	}

	return nil
}

// Restore restores a soft-deleted note
func (r *NoteRepository) Restore(id, userID uuid.UUID) error {
	query := `
		UPDATE notes 
		SET status = 'active', deleted_at = NULL, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'deleted'
	`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to restore note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note not found or not deleted")
	}

	return nil
}

// HardDelete permanently deletes a note from database
func (r *NoteRepository) HardDelete(id, userID uuid.UUID) error {
	query := `DELETE FROM notes WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to hard delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note not found")
	}

	return nil
}

// GetByUserID retrieves notes by user ID with pagination and filtering
func (r *NoteRepository) GetByUserID(userID uuid.UUID, params *model.GetNotesParams) ([]model.Note, int64, error) {
	// Set defaults
	params.SetDefaults()

	// Build WHERE clause
	whereConditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIndex := 2

	// Status filter
	if params.Status != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	// Public filter
	if params.IsPublic != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *params.IsPublic)
		argIndex++
	}

	// Search in title and content
	if params.Search != "" {
		searchQuery := fmt.Sprintf(`(
			to_tsvector('english', title) @@ plainto_tsquery('english', $%d) OR
			to_tsvector('english', coalesce(content, '')) @@ plainto_tsquery('english', $%d) OR
			title ILIKE $%d OR
			content ILIKE $%d
		)`, argIndex, argIndex+1, argIndex+2, argIndex+3)

		whereConditions = append(whereConditions, searchQuery)
		searchPattern := "%" + params.Search + "%"
		args = append(args, params.Search, params.Search, searchPattern, searchPattern)
		argIndex += 4
	}

	// Tags filter
	if params.Tags != "" {
		tagsArray := params.GetTagsArray()
		if len(tagsArray) > 0 {
			tagConditions := make([]string, len(tagsArray))
			for i, tag := range tagsArray {
				tagConditions[i] = fmt.Sprintf("tags ILIKE $%d", argIndex)
				args = append(args, "%"+tag+"%")
				argIndex++
			}
			whereConditions = append(whereConditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notes WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notes: %w", err)
	}

	// Build main query with pagination
	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf(`
		SELECT id, user_id, title, content, status, tags, is_public, view_count, 
			   created_at, updated_at, deleted_at
		FROM notes 
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, params.SortBy, params.SortDir, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var note model.Note
		err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.Title,
			&note.Content,
			&note.Status,
			&note.Tags,
			&note.IsPublic,
			&note.ViewCount,
			&note.CreatedAt,
			&note.UpdatedAt,
			&note.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan note row: %w", err)
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating note rows: %w", err)
	}

	return notes, total, nil
}

// Search performs advanced search across notes
func (r *NoteRepository) Search(userID uuid.UUID, req *model.NoteSearchRequest) ([]model.Note, int64, error) {
	// Set defaults
	req.SetDefaults()

	// Build WHERE clause
	whereConditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIndex := 2

	// Status filter
	if req.Status != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	// Public filter
	if req.IsPublic != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	// Full-text search query (only if query is provided)
	if req.Query != "" {
		searchQuery := fmt.Sprintf(`(
			to_tsvector('english', title) @@ plainto_tsquery('english', $%d) OR
			to_tsvector('english', coalesce(content, '')) @@ plainto_tsquery('english', $%d)
		)`, argIndex, argIndex+1)

		whereConditions = append(whereConditions, searchQuery)
		args = append(args, req.Query, req.Query)
		argIndex += 2
	}

	// Tags filter
	if len(req.Tags) > 0 {
		tagConditions := make([]string, len(req.Tags))
		for i, tag := range req.Tags {
			tagConditions[i] = fmt.Sprintf("tags ILIKE $%d", argIndex)
			args = append(args, "%"+tag+"%")
			argIndex++
		}
		whereConditions = append(whereConditions, "("+strings.Join(tagConditions, " OR ")+")")
	}

	// Date range filter
	if req.DateFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *req.DateFrom)
		argIndex++
	}

	if req.DateTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *req.DateTo)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notes WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build main query with ranking
	offset := (req.Page - 1) * req.PageSize

	selectFields := "id, user_id, title, status, tags, is_public, view_count, created_at, updated_at, deleted_at"
	if req.IncludeContent {
		selectFields = "id, user_id, title, content, status, tags, is_public, view_count, created_at, updated_at, deleted_at"
	} else {
		selectFields = "id, user_id, title, NULL as content, status, tags, is_public, view_count, created_at, updated_at, deleted_at"
	}

	var query string
	if req.Query != "" {
		// With text search ranking
		query = fmt.Sprintf(`
			SELECT %s,
				   ts_rank_cd(to_tsvector('english', title), plainto_tsquery('english', $%d)) +
				   ts_rank_cd(to_tsvector('english', coalesce(content, '')), plainto_tsquery('english', $%d)) as rank
			FROM notes 
			WHERE %s
			ORDER BY rank DESC, updated_at DESC
			LIMIT $%d OFFSET $%d
		`, selectFields, argIndex, argIndex+1, whereClause, argIndex+2, argIndex+3)
		args = append(args, req.Query, req.Query, req.PageSize, offset)
	} else {
		// Without text search ranking
		query = fmt.Sprintf(`
			SELECT %s, 0 as rank
			FROM notes 
			WHERE %s
			ORDER BY updated_at DESC
			LIMIT $%d OFFSET $%d
		`, selectFields, whereClause, argIndex, argIndex+1)
		args = append(args, req.PageSize, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var note model.Note
		var rank float64
		err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.Title,
			&note.Content,
			&note.Status,
			&note.Tags,
			&note.IsPublic,
			&note.ViewCount,
			&note.CreatedAt,
			&note.UpdatedAt,
			&note.DeletedAt,
			&rank,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating search results: %w", err)
	}

	return notes, total, nil
}

// IncrementViewCount increments the view count for a note
func (r *NoteRepository) IncrementViewCount(id uuid.UUID) error {
	query := `UPDATE notes SET view_count = view_count + 1, updated_at = NOW() WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// GetPublicNotes retrieves public notes with pagination
func (r *NoteRepository) GetPublicNotes(params *model.GetNotesParams) ([]model.Note, int64, error) {
	// Set defaults
	params.SetDefaults()

	// Build WHERE clause for public notes
	whereConditions := []string{"is_public = true", "status = 'active'"}
	args := []interface{}{}
	argIndex := 1

	// Search in title and content
	if params.Search != "" {
		searchQuery := fmt.Sprintf(`(
			to_tsvector('english', title) @@ plainto_tsquery('english', $%d) OR
			to_tsvector('english', coalesce(content, '')) @@ plainto_tsquery('english', $%d) OR
			title ILIKE $%d OR
			content ILIKE $%d
		)`, argIndex, argIndex+1, argIndex+2, argIndex+3)

		whereConditions = append(whereConditions, searchQuery)
		searchPattern := "%" + params.Search + "%"
		args = append(args, params.Search, params.Search, searchPattern, searchPattern)
		argIndex += 4
	}

	// Tags filter
	if params.Tags != "" {
		tagsArray := params.GetTagsArray()
		if len(tagsArray) > 0 {
			tagConditions := make([]string, len(tagsArray))
			for i, tag := range tagsArray {
				tagConditions[i] = fmt.Sprintf("tags ILIKE $%d", argIndex)
				args = append(args, "%"+tag+"%")
				argIndex++
			}
			whereConditions = append(whereConditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notes WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count public notes: %w", err)
	}

	// Build main query with pagination
	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf(`
		SELECT id, user_id, title, content, status, tags, is_public, view_count, 
			   created_at, updated_at, deleted_at
		FROM notes 
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, params.SortBy, params.SortDir, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query public notes: %w", err)
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var note model.Note
		err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.Title,
			&note.Content,
			&note.Status,
			&note.Tags,
			&note.IsPublic,
			&note.ViewCount,
			&note.CreatedAt,
			&note.UpdatedAt,
			&note.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan public note row: %w", err)
		}
		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating public note rows: %w", err)
	}

	return notes, total, nil
}

// BulkUpdateStatus updates status for multiple notes
func (r *NoteRepository) BulkUpdateStatus(userID uuid.UUID, noteIDs []uuid.UUID, status model.NoteStatus) error {
	if len(noteIDs) == 0 {
		return fmt.Errorf("no note IDs provided")
	}

	// Create placeholders for note IDs
	placeholders := make([]string, len(noteIDs))
	args := []interface{}{userID, status}

	for i, noteID := range noteIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+3) // Start from $3 since $1=userID, $2=status
		args = append(args, noteID)
	}

	query := fmt.Sprintf(`
		UPDATE notes 
		SET status = $2, updated_at = NOW()
		WHERE user_id = $1 AND id IN (%s)
	`, strings.Join(placeholders, ","))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no notes updated")
	}

	return nil
}

// GetNoteStats returns statistics for user's notes
func (r *NoteRepository) GetNoteStats(userID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'draft' THEN 1 END) as drafts,
			COUNT(CASE WHEN status = 'deleted' THEN 1 END) as deleted,
			COUNT(CASE WHEN is_public = true AND status = 'active' THEN 1 END) as public,
			COALESCE(SUM(view_count), 0) as total_views
		FROM notes 
		WHERE user_id = $1
	`

	var total, active, drafts, deleted, public, totalViews int64
	err := r.db.QueryRow(query, userID).Scan(&total, &active, &drafts, &deleted, &public, &totalViews)
	if err != nil {
		return nil, fmt.Errorf("failed to get note stats: %w", err)
	}

	stats := map[string]interface{}{
		"total":       total,
		"active":      active,
		"drafts":      drafts,
		"deleted":     deleted,
		"public":      public,
		"total_views": totalViews,
	}

	return stats, nil
}
