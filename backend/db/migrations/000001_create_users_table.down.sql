
-- Rollback for 000001_create_users_table.up.sql
DROP TABLE IF EXISTS users;

-- Drop enum type created for users.source (if present)
DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'users_source') THEN
		DROP TYPE users_source;
	END IF;
END$$;
