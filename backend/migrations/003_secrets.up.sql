CREATE TABLE secrets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment     TEXT NOT NULL,           -- 'dev' | 'staging' | 'prod'
    key             TEXT NOT NULL,
    encrypted_value BYTEA NOT NULL,          -- AES-256-GCM ciphertext of the secret value
    data_nonce      BYTEA NOT NULL,          -- 12-byte nonce used to encrypt the value
    encrypted_dek   BYTEA NOT NULL,          -- AES-256-GCM ciphertext of the DEK
    dek_nonce       BYTEA NOT NULL,          -- 12-byte nonce used to encrypt the DEK
    version         INTEGER NOT NULL DEFAULT 1,
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- A secret key is unique per project+environment combination
    UNIQUE(project_id, environment, key)
);

CREATE INDEX idx_secrets_project_env ON secrets(project_id, environment);
