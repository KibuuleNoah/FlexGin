package article

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"flex/internal/apperror"
	"flex/internal/store"
)

// fakeQ is an in-memory store.Querier. Embedding the interface means any
// method not overridden panics, so tests fail loudly if the service calls something unexpected.
type fakeQ struct {
	store.Querier
	items  map[int64]store.Article
	nextID int64
}

func newFakeQ() *fakeQ {
	return &fakeQ{items: map[int64]store.Article{}, nextID: 1}
}

func (f *fakeQ) CreateArticle(_ context.Context, p store.CreateArticleParams) (store.Article, error) {
	for _, it := range f.items {
		if it.Slug == p.Slug {
			return store.Article{}, &pgconn.PgError{Code: "23505"}
		}
	}
	now := time.Now()
	a := store.Article{
		ID: f.nextID, Title: p.Title, Slug: p.Slug, Body: p.Body,
		Published: p.Published, CreatedAt: now, UpdatedAt: now,
	}
	f.items[a.ID] = a
	f.nextID++
	return a, nil
}

func (f *fakeQ) GetArticle(_ context.Context, id int64) (store.Article, error) {
	a, ok := f.items[id]
	if !ok {
		return store.Article{}, sql.ErrNoRows
	}
	return a, nil
}

func (f *fakeQ) ListArticles(context.Context, store.ListArticlesParams) ([]store.Article, error) {
	out := make([]store.Article, 0, len(f.items))
	for _, a := range f.items {
		out = append(out, a)
	}
	return out, nil
}

func (f *fakeQ) CountArticles(context.Context) (int64, error) {
	return int64(len(f.items)), nil
}

func (f *fakeQ) UpdateArticle(_ context.Context, p store.UpdateArticleParams) (store.Article, error) {
	a, ok := f.items[p.ID]
	if !ok {
		return store.Article{}, sql.ErrNoRows
	}
	if p.Title != nil {
		a.Title = *p.Title
	}
	if p.Slug != nil {
		a.Slug = *p.Slug
	}
	if p.Body != nil {
		a.Body = *p.Body
	}
	if p.Published != nil {
		a.Published = *p.Published
	}
	a.UpdatedAt = time.Now()
	f.items[p.ID] = a
	return a, nil
}

func (f *fakeQ) DeleteArticle(_ context.Context, id int64) (int64, error) {
	if _, ok := f.items[id]; !ok {
		return 0, nil
	}
	delete(f.items, id)
	return 1, nil
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name           string
		req            CreateRequest
		seed           bool
		wantErr        error
		wantValidation bool
	}{
		{name: "ok", req: CreateRequest{Title: "Hello", Slug: "hello-world", Body: "x"}},
		{name: "bad slug", req: CreateRequest{Title: "Hello", Slug: "Hello World", Body: "x"}, wantValidation: true},
		{name: "duplicate slug", req: CreateRequest{Title: "Hello", Slug: "hello-world", Body: "x"}, seed: true, wantErr: ErrSlugTaken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(newFakeQ())
			if tt.seed {
				if _, err := svc.Create(context.Background(), tt.req); err != nil {
					t.Fatal(err)
				}
			}
			_, err := svc.Create(context.Background(), tt.req)
			switch {
			case tt.wantValidation:
				var ae *apperror.Error
				if !errors.As(err, &ae) || ae.Kind != apperror.KindValidation {
					t.Fatalf("want validation error, got %v", err)
				}
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
			case err != nil:
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	svc := NewService(newFakeQ())
	if _, err := svc.Get(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestService_Update(t *testing.T) {
	svc := NewService(newFakeQ())
	ctx := context.Background()

	a, err := svc.Create(ctx, CreateRequest{Title: "Hello", Slug: "hello", Body: "x"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Update(ctx, a.ID, UpdateRequest{}); err == nil {
		t.Fatal("empty update must fail")
	}

	bad := "Bad Slug"
	if _, err := svc.Update(ctx, a.ID, UpdateRequest{Slug: &bad}); err == nil {
		t.Fatal("invalid slug must fail")
	}

	title := "Changed"
	got, err := svc.Update(ctx, a.ID, UpdateRequest{Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Changed" || got.Slug != "hello" {
		t.Fatalf("partial update wrong: %+v", got)
	}

	if _, err := svc.Update(ctx, 999, UpdateRequest{Title: &title}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	svc := NewService(newFakeQ())
	ctx := context.Background()
	for _, slug := range []string{"one", "two"} {
		if _, err := svc.Create(ctx, CreateRequest{Title: "Title", Slug: slug, Body: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	items, total, err := svc.List(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || total != 2 {
		t.Fatalf("got %d items, total %d", len(items), total)
	}
}

func TestService_Delete(t *testing.T) {
	svc := NewService(newFakeQ())
	if err := svc.Delete(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
