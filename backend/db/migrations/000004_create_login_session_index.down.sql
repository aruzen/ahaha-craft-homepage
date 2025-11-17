ALTER TABLE login_sessions
    ALTER COLUMN id DROP DEFAULT;

DROP UNIQUE INDEX IF EXISTS login_sessions_token_key;
DROP INDEX IF EXISTS login_sessions_expires_at_idx;