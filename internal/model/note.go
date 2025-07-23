package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// NoteStatus represents the status of a note
type NoteStatus string

const (
	NoteStatusActive  NoteStatus = "active"
	NoteStatusDeleted NoteStatus = "deleted"
	NoteStatusDraft   NoteStatus = "draft"
)

// Note represents a note in the system
type Note struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Title     string     `json:"title" db:"title"`
	Content   *string    `json:"content" db:"content"`
	Status    NoteStatus `json:"status" db:"status"`
	Tags      *string    `json:"tags" db:"tags"` // JSON array stored as string
	IsPublic  bool       `json:"is_public" db:"is_public"`
	ViewCount int64      `json:"view_count" db:"view_count"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ToResponse converts Note to response format (without sensitive data)
func (n *Note) ToResponse() *NoteResponse {
	return &NoteResponse{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Status:    n.Status,
		Tags:      n.GetTagsArray(),
		IsPublic:  n.IsPublic,
		ViewCount: n.ViewCount,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// ToListItem converts Note to list item format (minimal data)
func (n *Note) ToListItem() *NoteListItem {
	preview := ""
	if n.Content != nil {
		content := *n.Content
		if len(content) > 150 {
			preview = content[:150] + "..."
		} else {
			preview = content
		}
	}

	return &NoteListItem{
		ID:        n.ID,
		Title:     n.Title,
		Preview:   preview,
		Status:    n.Status,
		Tags:      n.GetTagsArray(),
		IsPublic:  n.IsPublic,
		ViewCount: n.ViewCount,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// GetTagsArray converts tags string to array
func (n *Note) GetTagsArray() []string {
	if n.Tags == nil || *n.Tags == "" {
		return []string{}
	}

	// Simple split by comma for now, can be enhanced to JSON array later
	tags := strings.Split(*n.Tags, ",")
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// IsActive checks if note is active
func (n *Note) IsActive() bool {
	return n.Status == NoteStatusActive
}

// IsDeleted checks if note is soft deleted
func (n *Note) IsDeleted() bool {
	return n.Status == NoteStatusDeleted || n.DeletedAt != nil
}

// CanEdit checks if note can be edited
func (n *Note) CanEdit() bool {
	return n.Status == NoteStatusActive || n.Status == NoteStatusDraft
}

// NoteResponse represents a note response (full data)
type NoteResponse struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Content   *string    `json:"content"`
	Status    NoteStatus `json:"status"`
	Tags      []string   `json:"tags"`
	IsPublic  bool       `json:"is_public"`
	ViewCount int64      `json:"view_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NoteListItem represents a note in list view (minimal data)
type NoteListItem struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Preview   string     `json:"preview"`
	Status    NoteStatus `json:"status"`
	Tags      []string   `json:"tags"`
	IsPublic  bool       `json:"is_public"`
	ViewCount int64      `json:"view_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateNoteRequest represents a request to create a note
type CreateNoteRequest struct {
	Title    string   `json:"title" validate:"required,min=1,max=255"`
	Content  *string  `json:"content" validate:"omitempty,max=10000"`
	Status   *string  `json:"status" validate:"omitempty,oneof=active draft"`
	Tags     []string `json:"tags" validate:"omitempty,max=10,dive,min=1,max=50"`
	IsPublic *bool    `json:"is_public"`
}

// ToNote converts CreateNoteRequest to Note model
func (req *CreateNoteRequest) ToNote(userID uuid.UUID) *Note {
	note := &Note{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     strings.TrimSpace(req.Title),
		Content:   req.Content,
		Status:    NoteStatusActive,
		IsPublic:  false,
		ViewCount: 0,
	}

	// Set status
	if req.Status != nil {
		note.Status = NoteStatus(*req.Status)
	}

	// Set is_public
	if req.IsPublic != nil {
		note.IsPublic = *req.IsPublic
	}

	// Set tags
	if len(req.Tags) > 0 {
		tagsStr := strings.Join(req.Tags, ",")
		note.Tags = &tagsStr
	}

	return note
}

// UpdateNoteRequest represents a request to update a note
type UpdateNoteRequest struct {
	Title    *string  `json:"title" validate:"omitempty,min=1,max=255"`
	Content  *string  `json:"content" validate:"omitempty,max=10000"`
	Status   *string  `json:"status" validate:"omitempty,oneof=active draft deleted"`
	Tags     []string `json:"tags" validate:"omitempty,max=10,dive,min=1,max=50"`
	IsPublic *bool    `json:"is_public"`
}

// ApplyToNote applies update request to existing note
func (req *UpdateNoteRequest) ApplyToNote(note *Note) {
	if req.Title != nil {
		note.Title = strings.TrimSpace(*req.Title)
	}

	if req.Content != nil {
		note.Content = req.Content
	}

	if req.Status != nil {
		note.Status = NoteStatus(*req.Status)
		// Set deleted_at when status changes to deleted
		if note.Status == NoteStatusDeleted && note.DeletedAt == nil {
			now := time.Now()
			note.DeletedAt = &now
		} else if note.Status != NoteStatusDeleted {
			note.DeletedAt = nil
		}
	}

	if req.IsPublic != nil {
		note.IsPublic = *req.IsPublic
	}

	// Handle tags update
	if req.Tags != nil {
		if len(req.Tags) > 0 {
			tagsStr := strings.Join(req.Tags, ",")
			note.Tags = &tagsStr
		} else {
			note.Tags = nil
		}
	}

	note.UpdatedAt = time.Now()
}

// GetNotesParams represents query parameters for getting notes
type GetNotesParams struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Search   string `json:"search" validate:"omitempty,max=255"`
	Status   string `json:"status" validate:"omitempty,oneof=active draft deleted all"`
	Tags     string `json:"tags" validate:"omitempty,max=500"`
	IsPublic *bool  `json:"is_public"`
	SortBy   string `json:"sort_by" validate:"omitempty,oneof=created_at updated_at title view_count"`
	SortDir  string `json:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

// SetDefaults sets default values for GetNotesParams
func (p *GetNotesParams) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.SortBy == "" {
		p.SortBy = "updated_at"
	}
	if p.SortDir == "" {
		p.SortDir = "desc"
	}
	if p.Status == "" {
		p.Status = "active"
	}
}

// GetTagsArray converts tags string to array for filtering
func (p *GetNotesParams) GetTagsArray() []string {
	if p.Tags == "" {
		return []string{}
	}

	tags := strings.Split(p.Tags, ",")
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// NotesListResponse represents paginated notes response
type NotesListResponse struct {
	Notes      []NoteListItem `json:"notes"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	HasNext    bool           `json:"has_next"`
	HasPrev    bool           `json:"has_prev"`
}

// NewNotesListResponse creates a new paginated response
func NewNotesListResponse(notes []Note, total int64, params *GetNotesParams) *NotesListResponse {
	// Convert notes to list items
	items := make([]NoteListItem, len(notes))
	for i, note := range notes {
		items[i] = *note.ToListItem()
	}

	// Calculate pagination
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))
	hasNext := params.Page < totalPages
	hasPrev := params.Page > 1

	return &NotesListResponse{
		Notes:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// NoteSearchRequest represents advanced search request
type NoteSearchRequest struct {
	Query          string   `json:"query" validate:"omitempty,min=1,max=255"`
	Tags           []string `json:"tags" validate:"omitempty,max=10,dive,min=1,max=50"`
	Status         string   `json:"status" validate:"omitempty,oneof=active draft deleted all"`
	IsPublic       *bool    `json:"is_public"`
	DateFrom       *string  `json:"date_from" validate:"omitempty,datetime=2006-01-02"`
	DateTo         *string  `json:"date_to" validate:"omitempty,datetime=2006-01-02"`
	IncludeContent bool     `json:"include_content"`
	Page           int      `json:"page" validate:"min=1"`
	PageSize       int      `json:"page_size" validate:"min=1,max=100"`
}

// SetDefaults sets default values for NoteSearchRequest
func (req *NoteSearchRequest) SetDefaults() {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.Status == "" {
		req.Status = "active"
	}
}

// BulkOperationRequest represents bulk operations on notes
type BulkOperationRequest struct {
	NoteIDs   []uuid.UUID `json:"note_ids" validate:"required,min=1,max=50,dive,required"`
	Operation string      `json:"operation" validate:"required,oneof=delete restore update_status add_tags remove_tags"`
	Data      interface{} `json:"data,omitempty"`
}

// BulkUpdateData represents data for bulk update operations
type BulkUpdateData struct {
	Status   *string  `json:"status" validate:"omitempty,oneof=active draft deleted"`
	Tags     []string `json:"tags" validate:"omitempty,max=10,dive,min=1,max=50"`
	IsPublic *bool    `json:"is_public"`
}
