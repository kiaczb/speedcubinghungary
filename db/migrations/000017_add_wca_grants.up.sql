CREATE TABLE wca_grants (
    user_id       INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    scope         TEXT NOT NULL,
    access_token  BYTEA NOT NULL,
    refresh_token BYTEA NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    granted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
