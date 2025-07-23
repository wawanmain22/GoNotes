-- Create enum type for note status
CREATE TYPE note_status AS ENUM ('active', 'draft', 'deleted');

-- Add new columns to notes table
ALTER TABLE notes 
ADD COLUMN status note_status NOT NULL DEFAULT 'active',
ADD COLUMN tags TEXT,
ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN view_count BIGINT NOT NULL DEFAULT 0,
ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Create indexes for better performance on new fields
CREATE INDEX idx_notes_status ON notes(status) WHERE status != 'deleted';
CREATE INDEX idx_notes_is_public ON notes(is_public) WHERE is_public = TRUE;
CREATE INDEX idx_notes_view_count ON notes(view_count DESC);
CREATE INDEX idx_notes_deleted_at ON notes(deleted_at) WHERE deleted_at IS NOT NULL;

-- Create composite indexes for common query patterns
CREATE INDEX idx_notes_user_status ON notes(user_id, status) WHERE status != 'deleted';
CREATE INDEX idx_notes_user_updated ON notes(user_id, updated_at DESC);
CREATE INDEX idx_notes_public_active ON notes(is_public, status) WHERE is_public = TRUE AND status = 'active';

-- Create full-text search index for title and content
CREATE INDEX idx_notes_search_title ON notes USING gin(to_tsvector('english', title));
CREATE INDEX idx_notes_search_content ON notes USING gin(to_tsvector('english', coalesce(content, '')));

-- Create index for tags search (using GIN for better text search)
CREATE INDEX idx_notes_tags ON notes USING gin(to_tsvector('english', coalesce(tags, '')));

-- Add constraint to ensure deleted notes have deleted_at timestamp
ALTER TABLE notes 
ADD CONSTRAINT check_deleted_at_when_deleted 
CHECK (
  (status = 'deleted' AND deleted_at IS NOT NULL) OR 
  (status != 'deleted' AND deleted_at IS NULL)
);

-- Add constraint for view_count to be non-negative
ALTER TABLE notes 
ADD CONSTRAINT check_view_count_non_negative 
CHECK (view_count >= 0);

-- Create function to automatically set deleted_at when status changes to deleted
CREATE OR REPLACE FUNCTION handle_note_status_change()
RETURNS TRIGGER AS $$
BEGIN
  -- If status changed to deleted, set deleted_at
  IF NEW.status = 'deleted' AND OLD.status != 'deleted' THEN
    NEW.deleted_at = NOW();
  -- If status changed from deleted, clear deleted_at
  ELSIF NEW.status != 'deleted' AND OLD.status = 'deleted' THEN
    NEW.deleted_at = NULL;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic deleted_at handling
CREATE TRIGGER trigger_handle_note_status_change
  BEFORE UPDATE ON notes
  FOR EACH ROW
  EXECUTE FUNCTION handle_note_status_change();

-- Update existing notes to have proper status (all existing notes should be active)
UPDATE notes SET status = 'active' WHERE status IS NULL;

-- Add comment to the table for documentation
COMMENT ON TABLE notes IS 'Notes table with support for status management, tags, public sharing, and soft delete';
COMMENT ON COLUMN notes.status IS 'Note status: active, draft, or deleted (soft delete)';
COMMENT ON COLUMN notes.tags IS 'Comma-separated tags for categorization';
COMMENT ON COLUMN notes.is_public IS 'Whether the note is publicly viewable';
COMMENT ON COLUMN notes.view_count IS 'Number of times the note has been viewed';
COMMENT ON COLUMN notes.deleted_at IS 'Timestamp when the note was soft deleted'; 