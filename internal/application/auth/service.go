// Package auth provides login / token-refresh use cases.
package auth

import (
	"context"
	"errors"

	"github.com/zhanbinb/go-clean-arch_demo/internal/domain/user"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/jwt"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

type UserRepo = user.Repository

// Service is the authentication use case.
type Service struct {
	users UserRepo
	jwt   *jwt.Manager
	log   *logger.Logger
}

// NewService wires dependencies.
func NewService(users UserRepo, jwtMgr *jwt.Manager, log *logger.Logger) *Service {
	return &Service{users: users, jwt: jwtMgr, log: log}
}

// Login validates credentials and returns a pair of tokens.
func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	u, err := s.users.GetByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, errcode.ErrUnauthorized.WithCause(user.ErrInvalidCredentials)
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := u.VerifyPassword(in.Password); err != nil {
		return nil, errcode.ErrUnauthorized.WithCause(user.ErrInvalidCredentials)
	}
	access, err := s.jwt.Sign(u.ID, u.Username)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	refresh, err := s.jwt.SignRefresh(u.ID, u.Username)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.jwt.TTL().Seconds()),
	}, nil
}

// Refresh issues a new access token given a valid refresh token.
func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*LoginResult, error) {
	claims, err := s.jwt.Parse(in.RefreshToken)
	if err != nil {
		return nil, errcode.ErrTokenInvalid.WithCause(err)
	}
	// Issue a fresh pair
	access, err := s.jwt.Sign(claims.UserID, claims.Username)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	refresh, err := s.jwt.SignRefresh(claims.UserID, claims.Username)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.jwt.TTL().Seconds()),
	}, nil
}

// Register creates a new user. Intended for admin / bootstrap flows.
func (s *Service) Register(ctx context.Context, username, password string) error {
	// Check duplicate username
	existing, err := s.users.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return errcode.ErrInternal.WithCause(err)
	}
	if existing != nil {
		return errcode.ErrConflict.WithCause(user.ErrAlreadyExists)
	}
	u, err := user.NewUser(username, password)
	if err != nil {
		return errcode.ErrInvalidParam.WithCause(err)
	}
	if _, err := s.users.Save(ctx, u); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
