-- +goose Up
CREATE UNLOGGED TABLE cache_items (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    expires_at TIMESTAMPTZ 
);

CREATE INDEX idx_cache_expiration ON cache_items(expires_at);

-- +goose Down
DROP TABLE IF EXISTS cache_items;

