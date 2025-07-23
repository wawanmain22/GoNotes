-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_is_valid;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_user_id;

-- Drop sessions table
DROP TABLE IF EXISTS sessions; 