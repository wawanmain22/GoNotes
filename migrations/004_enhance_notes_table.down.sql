-- Drop comments
COMMENT ON TABLE notes IS NULL;
COMMENT ON COLUMN notes.status IS NULL;
COMMENT ON COLUMN notes.tags IS NULL;
COMMENT ON COLUMN notes.is_public IS NULL;
COMMENT ON COLUMN notes.view_count IS NULL;
COMMENT ON COLUMN notes.deleted_at IS NULL;

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_handle_note_status_change ON notes;

-- Drop function
DROP FUNCTION IF EXISTS handle_note_status_change();

-- Drop constraints
ALTER TABLE notes DROP CONSTRAINT IF EXISTS check_deleted_at_when_deleted;
ALTER TABLE notes DROP CONSTRAINT IF EXISTS check_view_count_non_negative;

-- Drop indexes
DROP INDEX IF EXISTS idx_notes_status;
DROP INDEX IF EXISTS idx_notes_is_public;
DROP INDEX IF EXISTS idx_notes_view_count;
DROP INDEX IF EXISTS idx_notes_deleted_at;
DROP INDEX IF EXISTS idx_notes_user_status;
DROP INDEX IF EXISTS idx_notes_user_updated;
DROP INDEX IF EXISTS idx_notes_public_active;
DROP INDEX IF EXISTS idx_notes_search_title;
DROP INDEX IF EXISTS idx_notes_search_content;
DROP INDEX IF EXISTS idx_notes_tags;

-- Drop new columns
ALTER TABLE notes 
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS tags,
DROP COLUMN IF EXISTS is_public,
DROP COLUMN IF EXISTS view_count,
DROP COLUMN IF EXISTS deleted_at;

-- Drop enum type
DROP TYPE IF EXISTS note_status; 