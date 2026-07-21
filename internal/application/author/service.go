// Package author provides the author use case service.
package author

import (
	"context"
	"errors"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/author"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

type AuthorRepo = author.Repository

// Service is the author use case.
type Service struct {
	authors AuthorRepo
	log     *logger.Logger
}

// NewService wires dependencies.
func NewService(authors AuthorRepo, log *logger.Logger) *Service {
	return &Service{authors: authors, log: log}
}

// Create persists a new author. Returns ErrConflict if email already taken.
func (s *Service) Create(ctx context.Context, in CreateInput) (*AuthorDTO, error) {
	// Check email uniqueness
	existing, err := s.authors.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, author.ErrNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if existing != nil {
		return nil, errcode.ErrConflict.WithCause(author.ErrEmailTaken)
	}

	a, err := author.NewAuthor(in.Name, in.Email)
	if err != nil {
		return nil, errcode.ErrInvalidParam.WithCause(err)
	}
	saved, err := s.authors.Save(ctx, a)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toDTO(saved), nil
}

// GetByID fetches an author by id.
func (s *Service) GetByID(ctx context.Context, id int64) (*AuthorDTO, error) {
	a, err := s.authors.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toDTO(a), nil
}

// Update applies a partial update.
func (s *Service) Update(ctx context.Context, id int64, in UpdateInput) (*AuthorDTO, error) {
	a, err := s.authors.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if in.Name != nil {
		if err := a.Rename(*in.Name); err != nil {
			return nil, errcode.ErrInvalidParam.WithCause(err)
		}
	}
	if in.Email != nil {
		if err := a.ChangeEmail(*in.Email); err != nil {
			return nil, errcode.ErrInvalidParam.WithCause(err)
		}
	}
	if err := s.authors.Update(ctx, a); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toDTO(a), nil
}

// Delete removes an author.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.authors.Delete(ctx, id); err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

// List returns a page of authors.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*AuthorDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.authors.List(ctx, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	dtos := make([]*AuthorDTO, 0, len(items))
	for _, a := range items {
		dtos = append(dtos, toDTO(a))
	}
	return dtos, nil
}

func toDTO(a *author.Author) *AuthorDTO {
	return &AuthorDTO{
		ID: a.ID, Name: a.Name, Email: a.Email,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}
