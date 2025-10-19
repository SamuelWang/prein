
-- Create extension and users table for Prein v0.1.0

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum for authentication source/provider
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'users_source') THEN
		CREATE TYPE users_source AS ENUM ('google');
	END IF;
END$$;

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	-- Unique identifier for the user from the external authentication provider (e.g., OAuth provider user ID). Expected format: string, unique per provider.
	provider_id TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	name TEXT,
	family_name TEXT,
	given_name TEXT,
	avatar_url TEXT,
	-- Indicates the authentication source for the user (e.g., 'google').
	-- Used in conjunction with provider_id to uniquely identify users from different providers.
	source users_source NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (provider_id, source)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
