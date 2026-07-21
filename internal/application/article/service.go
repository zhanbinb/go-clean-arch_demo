// Package article provides the article use case service.
//
// It orchestrates the article and author repositories, handles business rules
// (e.g. verifying the author exists before creating), and translates domain
// entities to/from DTOs.
package article

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/article"
	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/author"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

// Repository aliases let us swap implementations (e.g. for testing) without
// importing the concrete packages from the application layer.
type (
	ArticleRepo = article.Repository
	AuthorRepo  = author.Repository
)

// Service is the article use case.
type Service struct {
	articles ArticleRepo
	authors  AuthorRepo
	log      *logger.Logger
}

// NewService wires dependencies.
func NewService(articles ArticleRepo, authors AuthorRepo, log *logger.Logger) *Service {
	return &Service{articles: articles, authors: authors, log: log}
}

// Create validates the input, ensures the author exists, and persists the article.
func (s *Service) Create(ctx context.Context, in CreateInput) (*ArticleDTO, error) {
	// 1. Verify author exists
	authorEntity, err := s.authors.GetByID(ctx, in.AuthorID)
	if err != nil {
		if errors.Is(err, author.ErrNotFound) {
			return nil, errcode.New(40003, "author does not exist")
		}
		s.log.Error("lookup author", logger.FieldsFromContext(ctx)...)
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// 2. Build domain entity (re-validates invariants)
	a, err := article.NewArticle(in.Title, in.Content, in.AuthorID, authorEntity.Name)
	if err != nil {
		return nil, errcode.ErrInvalidParam.WithCause(err)
	}

	// 3. Persist
	saved, err := s.articles.Save(ctx, a)
	if err != nil {
		s.log.Error("save article", append(logger.FieldsFromContext(ctx), zap.Error(err))...)
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return toDTO(saved), nil
}

// GetByID returns the article by id.
func (s *Service) GetByID(ctx context.Context, id int64) (*ArticleDTO, error) {
	a, err := s.articles.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, article.ErrNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toDTO(a), nil
}

// List returns a page of articles.
func (s *Service) List(ctx context.Context, cursor string, limit int) (*ListResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	items, next, err := s.articles.List(ctx, cursor, limit)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	dtos := make([]*ArticleDTO, 0, len(items))
	for _, a := range items {
		dtos = append(dtos, toDTO(a))
	}
	return &ListResult{Items: dtos, NextCursor: next}, nil
}

// Update applies the partial update.
func (s *Service) Update(ctx context.Context, id int64, in UpdateInput) (*ArticleDTO, error) {
	a, err := s.articles.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, article.ErrNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := a.Update(in.Title, in.Content); err != nil {
		return nil, errcode.ErrInvalidParam.WithCause(err)
	}
	if err := s.articles.Update(ctx, a); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toDTO(a), nil
}

// Delete removes an article.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.articles.Delete(ctx, id); err != nil {
		if errors.Is(err, article.ErrNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

// ListByAuthor returns all articles by a given author.
func (s *Service) ListByAuthor(ctx context.Context, authorID int64) ([]*ArticleDTO, error) {
	items, err := s.articles.GetByAuthor(ctx, authorID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	dtos := make([]*ArticleDTO, 0, len(items))
	for _, a := range items {
		dtos = append(dtos, toDTO(a))
	}
	return dtos, nil
}

// toDTO converts a domain entity to an application DTO.
func toDTO(a *article.Article) *ArticleDTO {
	return &ArticleDTO{
		ID:         a.ID,
		Title:      a.Title,
		Content:    a.Content,
		AuthorID:   a.AuthorID,
		AuthorName: a.AuthorName,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

// Sentinel to prevent an unused import warning if formatting helpers change.
var _ = fmt.Sprintf
