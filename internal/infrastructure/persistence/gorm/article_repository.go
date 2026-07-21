package gorm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/article"
)

// articleEntity is a local alias of the domain Article so we don't have to
// import domain types into this file (keeps infrastructure independent).
type articleEntity = article.Article

// ArticleRepository is the GORM implementation of article.Repository.
type ArticleRepository struct {
	db *gorm.DB
}

// NewArticleRepository returns a new instance.
func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Save(ctx context.Context, a *article.Article) (*article.Article, error) {
	m := &ArticleModel{
		Title: a.Title, Content: a.Content,
		AuthorID: a.AuthorID, AuthorName: a.AuthorName,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("insert article: %w", err)
	}
	return m.toEntity(), nil
}

func (r *ArticleRepository) GetByID(ctx context.Context, id int64) (*article.Article, error) {
	var m ArticleModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, article.ErrNotFound
		}
		return nil, fmt.Errorf("query article: %w", err)
	}
	return m.toEntity(), nil
}

func (r *ArticleRepository) List(ctx context.Context, cursor string, limit int) ([]*article.Article, string, error) {
	if limit <= 0 {
		limit = 10
	}
	q := r.db.WithContext(ctx).Order("id DESC").Limit(limit + 1)
	if cursor != "" {
		if cid, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			q = q.Where("id < ?", cid)
		}
	}
	var rows []ArticleModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, "", fmt.Errorf("list articles: %w", err)
	}
	items := make([]*article.Article, 0, len(rows))
	for i := range rows {
		items = append(items, rows[i].toEntity())
	}
	next := ""
	if len(items) > limit {
		next = strconv.FormatInt(items[limit-1].ID, 10)
		items = items[:limit]
	}
	return items, next, nil
}

func (r *ArticleRepository) Update(ctx context.Context, a *article.Article) error {
	res := r.db.WithContext(ctx).
		Model(&ArticleModel{}).
		Where("id = ?", a.ID).
		Updates(map[string]interface{}{
			"title":       a.Title,
			"content":     a.Content,
			"author_id":   a.AuthorID,
			"author_name": a.AuthorName,
		})
	if res.Error != nil {
		return fmt.Errorf("update article: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return article.ErrNotFound
	}
	return nil
}

func (r *ArticleRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&ArticleModel{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete article: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return article.ErrNotFound
	}
	return nil
}

func (r *ArticleRepository) GetByAuthor(ctx context.Context, authorID int64) ([]*article.Article, error) {
	var rows []ArticleModel
	err := r.db.WithContext(ctx).
		Where("author_id = ?", authorID).
		Order("id DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list by author: %w", err)
	}
	out := make([]*article.Article, 0, len(rows))
	for i := range rows {
		out = append(out, rows[i].toEntity())
	}
	return out, nil
}
