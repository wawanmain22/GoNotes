-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop index
DROP INDEX IF EXISTS idx_users_email;

-- Drop users table
DROP TABLE IF EXISTS users;

-- Drop function (only if no other triggers depend on it)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.triggers 
        WHERE trigger_name LIKE '%updated_at%'
    ) THEN
        DROP FUNCTION IF EXISTS update_updated_at_column();
    END IF;
END $$; 