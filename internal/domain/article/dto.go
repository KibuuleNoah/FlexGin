package article

type CreateArticleRequest struct {
	Title     string `json:"title"     binding:"required,min=3,max=200"`
	Slug      string `json:"slug"      binding:"required,min=3,max=200"`
	Body      string `json:"body"      binding:"required"`
	Published bool   `json:"published"`
}

// UpdateArticleRequest is a partial update: nil fields are left unchanged.
type UpdateArticleRequest struct {
	Title     *string `json:"title"     binding:"omitempty,min=3,max=200"`
	Slug      *string `json:"slug"      binding:"omitempty,min=3,max=200"`
	Body      *string `json:"body"      binding:"omitempty,min=1"`
	Published *bool   `json:"published"`
}
