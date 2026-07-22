package article

import (
	"context"
)

// Repository is the persistence contract for Article aggregates.
// Implementations live in internal/infrastructure/persistence/gorm.
//
// Methods take and return domain entities, not ORM models — keeping the
// domain layer unaware of the storage technology.
type Repository interface {
	// Save inserts a new article and returns the populated entity (with ID).
	Save(ctx context.Context, a *Article) (*Article, error)

	// GetByID fetches a single article. Returns ErrNotFound if missing.
	GetByID(ctx context.Context, id int64) (*Article, error)

	// List returns a page of articles ordered by ID DESC, plus a cursor for
	// the next page (empty cursor = no more pages).
	List(ctx context.Context, cursor string, limit int) ([]*Article, string, error)

	// Update persists changes to an existing article.
	Update(ctx context.Context, a *Article) error

	// Delete removes the article. Returns ErrNotFound if it doesn't exist.
	Delete(ctx context.Context, id int64) error

	// GetByAuthor returns all articles authored by the given author.
	GetByAuthor(ctx context.Context, authorID int64) ([]*Article, error)
}
