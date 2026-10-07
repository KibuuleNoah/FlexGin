// ===== internal/article/service.go =====
package article

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"FlexGin/internal/apperror"
	"FlexGin/internal/database"
	"FlexGin/internal/store"
	"FlexGin/internal/utils"
)

type Service struct {
	q store.Querier
}

// q is store.New(db) (db is *database.DB, which embeds *sql.DB).
func NewService(q store.Querier) *Service { return &Service{q: q} }

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateSlug(s string) error {
	if !slugRe.MatchString(s) {
		return apperror.Validation("validation failed", map[string]string{
			"slug": "must be lowercase letters, digits and single hyphens",
		})
	}
	return nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (store.Article, error) {
	slug := strings.TrimSpace(req.Slug)
	if err := validateSlug(slug); err != nil {
		return store.Article{}, err
	}
	a, err := s.q.CreateArticle(ctx, store.CreateArticleParams{
		Title: strings.TrimSpace(req.Title), Slug: slug, Body: req.Body, Published: req.Published,
	})
	return a, mapErr(err, "create article")
}

func (s *Service) Get(ctx context.Context, id string) (store.Article, error) {
	a, err := s.q.GetArticle(ctx, id)
	return a, mapErr(err, "get article")
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]store.Article, int, error) {
	rows, err := s.q.ListArticles(ctx, store.ListArticlesParams{PageLimit: int32(limit), PageOffset: int32(offset)})
	if err != nil {
		return nil, 0, mapErr(err, "list articles")
	}
	total, err := s.q.CountArticles(ctx)
	if err != nil {
		return nil, 0, mapErr(err, "count articles")
	}
	return rows, int(total), nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (store.Article, error) {
	if req.Title == nil && req.Slug == nil && req.Body == nil && req.Published == nil {
		return store.Article{}, apperror.Validation("no fields to update", nil)
	}

	if req.Slug != nil {
		if err := validateSlug(*req.Slug); err != nil {
			return store.Article{}, err
		}
	}
	a, err := s.q.UpdateArticle(
		ctx,
		store.UpdateArticleParams{
			ID:        id,
			Title:     utils.ToNullString(req.Title),
			Slug:      utils.ToNullString(req.Slug),
			Body:      req.Body,
			Published: utils.ToNullBool(req.Published),
		},
	)
	return a, mapErr(err, "update article")
}

func (s *Service) Delete(ctx context.Context, id string) error {
	n, err := s.q.DeleteArticle(ctx, id)
	if err != nil {
		return mapErr(err, "delete article")
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func mapErr(err error, op string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	case database.IsUniqueViolation(err):
		return ErrSlugTaken
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
