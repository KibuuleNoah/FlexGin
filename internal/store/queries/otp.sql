-- name: CreateOTPChallenge :one
INSERT INTO otp_challenges (id, user_id, code_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestOTPChallenge :one
SELECT * FROM otp_challenges
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetActiveOTPChallenge :one
SELECT * FROM otp_challenges
WHERE user_id = $1 AND consumed_at IS NULL AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;

-- name: GetOTPChallenge :one
SELECT * FROM otp_challenges
WHERE id = $1 AND consumed_at IS NULL AND expires_at > now();

-- name: IncrementOTPAttempts :exec
UPDATE otp_challenges SET attempts = attempts + 1 WHERE id = $1;

-- name: ConsumeOTPChallenge :execrows
UPDATE otp_challenges SET consumed_at = now()
WHERE id = $1 AND consumed_at IS NULL AND expires_at > now();
