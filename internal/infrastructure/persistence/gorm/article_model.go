package gorm

import "time"

// ArticleModel is the GORM database model for articles.
// It is intentionally separate from domain/article.Article to avoid leaking
// ORM concerns into the domain layer.
type ArticleModel struct {
	ID         int64    `gorm:"primaryKey;column:id"`
	Title      string    `gorm:"size:200;not null;column:title"`
	Content    string    `gorm:"type:text;not null;column:content"`
	AuthorID   int64    `gorm:"not null;index;column:author_id"`
	AuthorName string    `gorm:"size:100;not null;default:'';column:author_name"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

// TableName returns the underlying MySQL table name.
func (ArticleModel) TableName() string { return "articles" }

// toEntity converts GORM model → domain entity.
func (m *ArticleModel) toEntity() *articleEntity {
	return &articleEntity{
		ID: m.ID, Title: m.Title, Content: m.Content,
		AuthorID: m.AuthorID, AuthorName: m.AuthorName,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}
