package article

import "flexgin/internal/apperror"

var (
	ErrNotFound  = apperror.NotFound("article_not_found", "article not found")
	ErrSlugTaken = apperror.Conflict("article_slug_taken", "an article with this slug already exists")
)
