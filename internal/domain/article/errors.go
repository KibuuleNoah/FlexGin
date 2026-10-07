package article

import "flexgin/internal/apperror" // replace "yourmodule" with the module path from go.mod

var (
	ErrNotFound  = apperror.NotFound("article_not_found", "article not found")
	ErrSlugTaken = apperror.Conflict("article_slug_taken", "an article with this slug already exists")
)
