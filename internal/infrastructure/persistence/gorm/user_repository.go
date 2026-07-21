package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/user"
)

type userEntity = user.User

// UserRepository is the GORM implementation of user.Repository.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) (*user.User, error) {
	m := &UserModel{Username: u.Username, PasswordHash: u.PasswordHash}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return m.toEntity(), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	return m.toEntity(), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("query user by username: %w", err)
	}
	return m.toEntity(), nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	res := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", u.ID).
		Updates(map[string]interface{}{
			"username":      u.Username,
			"password_hash": u.PasswordHash,
		})
	if res.Error != nil {
		return fmt.Errorf("update user: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&UserModel{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete user: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return user.ErrNotFound
	}
	return nil
}
