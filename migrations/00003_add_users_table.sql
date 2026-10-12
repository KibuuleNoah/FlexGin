-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS users_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY DEFAULT generate_short_id('users_id_seq'),
    username      VARCHAR(255),
    password_hash TEXT,
    identifier    TEXT         NOT NULL UNIQUE,
    identifier_alt TEXT,
    role          TEXT         NOT NULL DEFAULT 'guest',
    avatar_url    TEXT,
    is_banned     BOOLEAN      NOT NULL DEFAULT FALSE,
    verified      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    BIGINT       NOT NULL,
    updated_at    BIGINT       NOT NULL,
    last_login_at BIGINT,
    CONSTRAINT users_role_check CHECK (role IN ('guest', 'user', 'admin'))
);

CREATE TRIGGER create_user_created_at
  BEFORE INSERT ON users
  FOR EACH ROW EXECUTE PROCEDURE created_at_column();

CREATE TRIGGER update_user_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE PROCEDURE update_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP SEQUENCE IF EXISTS users_id_seq;
-- +goose StatementEnd
