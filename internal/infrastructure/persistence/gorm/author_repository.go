package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/author"
)

type authorEntity = author.Author

// AuthorRepository is the GORM implementation of author.Repository.
type AuthorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}

func (r *AuthorRepository) Save(ctx context.Context, a *author.Author) (*author.Author, error) {
	m := &AuthorModel{Name: a.Name, Email: a.Email}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("insert author: %w", err)
	}
	return m.toEntity(), nil
}

func (r *AuthorRepository) GetByID(ctx context.Context, id int64) (*author.Author, error) {
	var m AuthorModel
	err := r.db.WithContext(ctx).First(&m, uint64(id)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, author.ErrNotFound
		}
		return nil, fmt.Errorf("query author: %w", err)
	}
	return m.toEntity(), nil
}

func (r *AuthorRepository) GetByEmail(ctx context.Context, email string) (*author.Author, error) {
	var m AuthorModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, author.ErrNotFound
		}
		return nil, fmt.Errorf("query author by email: %w", err)
	}
	return m.toEntity(), nil
}

func (r *AuthorRepository) Update(ctx context.Context, a *author.Author) error {
	res := r.db.WithContext(ctx).
		Model(&AuthorModel{}).
		Where("id = ?", a.ID).
		Updates(map[string]interface{}{"name": a.Name, "email": a.Email})
	if res.Error != nil {
		return fmt.Errorf("update author: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return author.ErrNotFound
	}
	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&AuthorModel{}, uint64(id))
	if res.Error != nil {
		return fmt.Errorf("delete author: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return author.ErrNotFound
	}
	return nil
}

func (r *AuthorRepository) List(ctx context.Context, limit, offset int) ([]*author.Author, error) {
	var rows []AuthorModel
	err := r.db.WithContext(ctx).
		Order("id ASC").
		Limit(limit).Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list authors: %w", err)
	}
	out := make([]*author.Author, 0, len(rows))
	for i := range rows {
		out = append(out, rows[i].toEntity())
	}
	return out, nil
}
