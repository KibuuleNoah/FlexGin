-- name: CreateSession :one
INSERT INTO sessions (id, user_id, refresh_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: SessionIsActive :one
SELECT EXISTS (
    SELECT 1 FROM sessions
    WHERE id = $1 AND revoked_at IS NULL AND expires_at > now()
);

-- atomic claim: only one concurrent refresh can win
-- name: ClaimRefreshSession :one
UPDATE sessions SET revoked_at = now()
WHERE refresh_hash = $1 AND revoked_at IS NULL AND expires_at > now()
RETURNING *;

-- name: RevokeSession :execrows
UPDATE sessions SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL;
