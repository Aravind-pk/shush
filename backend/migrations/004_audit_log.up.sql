CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id),
    action      TEXT NOT NULL,       -- 'read' | 'write' | 'delete'
    secret_key  TEXT NOT NULL,
    project_id  UUID NOT NULL REFERENCES projects(id),
    environment TEXT NOT NULL,
    ip_address  TEXT,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_project ON audit_log(project_id, timestamp DESC);
