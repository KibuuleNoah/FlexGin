package article

import "FlexGin/internal/apperror" // replace "yourmodule" with the module path from go.mod

// Sentinels are shared, read-only. Compare with errors.Is.
var (
	ErrNotFound  = apperror.NotFound("article_not_found", "article not found")
	ErrSlugTaken = apperror.Conflict("article_slug_taken", "an article with this slug already exists")
)
