CREATE TABLE IF NOT EXISTS oauth_sessions (
    id TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
)
