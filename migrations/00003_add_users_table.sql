-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS users_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY DEFAULT generate_short_id('users_id_seq'),
    username VARCHAR(255) NOT NULL,
    phone_no VARCHAR(16) NOT NULL UNIQUE,
    avatar_url TEXT,
    is_banned BOOLEAN DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT false,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    last_login_at BIGINT
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
DROP SEQUENCE IF EXISTS users_id_seq;
DROP TYPE IF EXISTS user_role;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
