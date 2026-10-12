-- +goose Up
-- +goose StatementBegin
-- one row per login; the access token's jti is sessions.id, the refresh token is stored hashed
CREATE SEQUENCE IF NOT EXISTS sessions_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE TABLE sessions (
    id TEXT PRIMARY KEY DEFAULT generate_short_id('sessions_id_seq'),
    user_id      BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_hash TEXT        NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    created_at  BIGINT,
    CONSTRAINT sessions_refresh_hash_key UNIQUE (refresh_hash)
);
CREATE INDEX idx_sessions_user_id ON sessions (user_id);

CREATE TRIGGER sessions_created_at
  BEFORE INSERT ON users
  FOR EACH ROW EXECUTE PROCEDURE created_at_column();

-- id doubles as the "otp resend key"
CREATE SEQUENCE IF NOT EXISTS otp_challenges_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE TABLE otp_challenges (
    id TEXT PRIMARY KEY DEFAULT generate_short_id('otp_challenges_id_seq'),
    user_id     TEXT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    code_hash   TEXT        NOT NULL,
    attempts    INT         NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  BIGINT
);

CREATE INDEX idx_otp_challenges_user_created ON otp_challenges (user_id, created_at DESC);

CREATE TRIGGER otp_challenges_created_at
  BEFORE INSERT ON otp_challenges
  FOR EACH ROW EXECUTE PROCEDURE created_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS otp_challenges;
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
