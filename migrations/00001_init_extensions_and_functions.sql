-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;

CREATE OR REPLACE FUNCTION created_at_column() RETURNS trigger
  LANGUAGE plpgsql AS $$
BEGIN
  NEW.created_at := EXTRACT(EPOCH FROM NOW())::bigint;
  NEW.updated_at := EXTRACT(EPOCH FROM NOW())::bigint;
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION update_at_column() RETURNS trigger
  LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := EXTRACT(EPOCH FROM NOW())::bigint;
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION generate_short_id(seq_name TEXT)
RETURNS TEXT
LANGUAGE plpgsql
AS $$
DECLARE
  chars  TEXT    := '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
  base   INTEGER := 62;
  len    INTEGER := 6;
  result TEXT    := '';
  num    BIGINT;
  rem    INTEGER;
BEGIN
  num := nextval(seq_name);
  num := num + (random() * 1000000)::BIGINT * 1000000000;

  WHILE length(result) < len LOOP
    rem    := num % base;
    result := substr(chars, rem + 1, 1) || result;
    num    := num / base;
  END LOOP;

  RETURN result;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS generate_short_id(TEXT);
DROP FUNCTION IF EXISTS update_at_column();
DROP FUNCTION IF EXISTS created_at_column();
-- +goose StatementEnd

