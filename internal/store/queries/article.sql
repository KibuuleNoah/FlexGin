-- name: CreateArticle :one
INSERT INTO articles (title, slug, body, published)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetArticle :one
SELECT * FROM articles WHERE id = $1;

-- name: GetArticleBySlug :one
SELECT * FROM articles WHERE slug = $1;

-- name: ListArticles :many
SELECT * FROM articles
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountArticles :one
SELECT count(*) FROM articles;

-- name: UpdateArticle :one
UPDATE articles
SET title      = COALESCE(sqlc.narg(title), title),
    slug       = COALESCE(sqlc.narg(slug), slug),
    body       = COALESCE(sqlc.narg(body), body),
    published  = COALESCE(sqlc.narg(published), published),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteArticle :execrows
DELETE FROM articles WHERE id = $1;
