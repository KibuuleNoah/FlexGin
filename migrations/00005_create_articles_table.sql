-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS articles_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

CREATE TABLE articles (
	id TEXT PRIMARY KEY DEFAULT generate_short_id('articles_id_seq'),
	title VARCHAR(200) NOT NULL,
	slug VARCHAR(220) NOT NULL UNIQUE,
	body TEXT NOT NULL,
	author_id TEXT NOT NULL REFERENCES users(id),
	published BOOLEAN NOT NULL DEFAULT false,
  created_at BIGINT NOT NULL,
  updated_at BIGINT NOT NULL,
  CONSTRAINT articles_slug_key UNIQUE (slug)
);

CREATE INDEX idx_articles_author_id ON articles(author_id);
CREATE INDEX idx_articles_published_created_at ON articles(published, created_at DESC);

-- this automatically creates created_at and updated_at columns
CREATE TRIGGER create_articles_created_at
  BEFORE INSERT ON articles
  FOR EACH ROW EXECUTE PROCEDURE created_at_column();

-- this automatically creates updated_at column
CREATE TRIGGER update_articles_updated_at
  BEFORE UPDATE ON articles
  FOR EACH ROW EXECUTE PROCEDURE update_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE articles;
DROP SEQUENCE IF EXISTS articles_id_seq;
-- +goose StatementEnd
