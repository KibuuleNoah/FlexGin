-- name: CreateUser :one
INSERT INTO users (username, phone_no, email, password_hash, avatar_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- OTP sign-up/login: creates a guest on first contact, returns the existing row otherwise
-- name: UpsertUserByPhone :one
INSERT INTO users (username, phone_no)
VALUES ($1, $2)
ON CONFLICT (phone_no) DO UPDATE SET phone_no = EXCLUDED.phone_no
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_no = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: UpdateUser :one
UPDATE users
SET username   = COALESCE(sqlc.narg(username), username),
    email      = COALESCE(sqlc.narg(email), email),
    phone_no   = COALESCE(sqlc.narg(phone_no), phone_no),
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetUserRole :one
UPDATE users SET role = $2 WHERE id = $1 RETURNING *;

-- name: SetUserBanned :one
UPDATE users SET is_banned = $2 WHERE id = $1 RETURNING *;

-- name: MarkUserVerified :one
UPDATE users SET verified = TRUE, role = CASE WHEN role = 'guest' THEN 'user' ELSE role END
WHERE id = $1
RETURNING *;

-- name: TouchUserLastLogin :exec
UPDATE users SET last_login_at = (extract(epoch FROM now()))::bigint WHERE id = $1;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = $1;
