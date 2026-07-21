package gorm

import "time"

// AuthorModel is the GORM database model for authors.
type AuthorModel struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"size:100;not null;column:name"`
	Email     string    `gorm:"size:255;not null;uniqueIndex;column:email"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the underlying MySQL table name.
func (AuthorModel) TableName() string { return "authors" }

func (m *AuthorModel) toEntity() *authorEntity {
	return &authorEntity{
		ID: int64(m.ID), Name: m.Name, Email: m.Email,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}
