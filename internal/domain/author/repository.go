package author

import "context"

// Repository is the persistence contract for Author aggregates.
type Repository interface {
	Save(ctx context.Context, a *Author) (*Author, error)
	GetByID(ctx context.Context, id int64) (*Author, error)
	GetByEmail(ctx context.Context, email string) (*Author, error)
	Update(ctx context.Context, a *Author) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*Author, error)
}
