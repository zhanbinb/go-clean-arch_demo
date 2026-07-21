package gorm

import "time"

// UserModel is the GORM database model for users.
type UserModel struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	Username     string    `gorm:"size:50;not null;uniqueIndex;column:username"`
	PasswordHash string    `gorm:"size:255;not null;column:password_hash"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

// TableName returns the underlying MySQL table name.
func (UserModel) TableName() string { return "users" }

func (m *UserModel) toEntity() *userEntity {
	return &userEntity{
		ID: int64(m.ID), Username: m.Username, PasswordHash: m.PasswordHash,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}
