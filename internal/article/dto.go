package article

import (
	"FlexGin/internal/store"
)

type CreateRequest struct {
	Title     string `json:"title"     binding:"required,min=3,max=200"`
	Slug      string `json:"slug"      binding:"required,min=3,max=200"`
	Body      string `json:"body"      binding:"required"`
	Published bool   `json:"published"`
}

// UpdateRequest is a partial update: nil fields are left unchanged.
type UpdateRequest struct {
	Title     *string `json:"title"     binding:"omitempty,min=3,max=200"`
	Slug      *string `json:"slug"      binding:"omitempty,min=3,max=200"`
	Body      *string `json:"body"      binding:"omitempty,min=1"`
	Published *bool   `json:"published"`
}

type Response struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Body      string `json:"body"`
	Published bool   `json:"published"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func ToResponse(a store.Article) Response {
	return Response{
		ID: a.ID, Title: a.Title, Slug: a.Slug, Body: a.Body,
		Published: a.Published, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func ToResponses(in []store.Article) []Response {
	out := make([]Response, len(in))
	for i, a := range in {
		out[i] = ToResponse(a)
	}
	return out
}
