-- Drop trigger
DROP TRIGGER IF EXISTS update_notes_updated_at ON notes;

-- Drop indexes
DROP INDEX IF EXISTS idx_notes_title;
DROP INDEX IF EXISTS idx_notes_created_at;
DROP INDEX IF EXISTS idx_notes_user_id;

-- Drop notes table
DROP TABLE IF EXISTS notes; 