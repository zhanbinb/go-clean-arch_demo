// Package user provides user-management use cases (separate from auth which
// handles token issuance).
package user

import (
	"context"
	"errors"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/user"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

type UserRepo = user.Repository

// Service is the user-management use case.
type Service struct {
	users UserRepo
	log   *logger.Logger
}

// NewService wires dependencies.
func NewService(users UserRepo, log *logger.Logger) *Service {
	return &Service{users: users, log: log}
}

// GetByID returns a user DTO by id (no password hash exposed).
func (s *Service) GetByID(ctx context.Context, id int64) (*UserDTO, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toDTO(u), nil
}

// ChangePassword lets a logged-in user rotate their own password.
func (s *Service) ChangePassword(ctx context.Context, id int64, oldPass, newPass string) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if err := u.ChangePassword(oldPass, newPass); err != nil {
		return errcode.ErrInvalidParam.WithCause(err)
	}
	if err := s.users.Update(ctx, u); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func toDTO(u *user.User) *UserDTO {
	return &UserDTO{
		ID: u.ID, Username: u.Username,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}
