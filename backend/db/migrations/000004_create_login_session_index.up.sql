CREATE UNIQUE INDEX login_sessions_token_key ON login_sessions (token);
CREATE INDEX login_sessions_expires_at_idx ON login_sessions (expires_at);

CREATE EXTENSION IF NOT EXISTS pgcrypto;
ALTER TABLE login_sessions
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
