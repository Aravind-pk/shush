CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         TEXT NOT NULL UNIQUE,
    clerk_user_id TEXT UNIQUE,            -- Clerk identity; users are provisioned from session tokens
    role          TEXT NOT NULL DEFAULT 'member',  -- 'admin' | 'member'
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_clerk_user_id ON users(clerk_user_id);
