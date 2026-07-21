package user

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is the aggregate root for authentication.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// MaxUsernameLen is the maximum allowed username length.
const (
	MaxUsernameLen = 50
	MinPasswordLen = 8
	bcryptCost     = 12
)

// NewUser is a factory that creates a User with a bcrypt-hashed password.
func NewUser(username, plaintextPassword string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > MaxUsernameLen {
		return nil, ErrInvalidInput
	}
	if len(plaintextPassword) < MinPasswordLen {
		return nil, ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcryptCost)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &User{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// VerifyPassword returns nil if the plaintext matches the stored hash.
func (u *User) VerifyPassword(plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext))
}

// ChangePassword validates the old password and updates to the new one.
func (u *User) ChangePassword(oldPlaintext, newPlaintext string) error {
	if err := u.VerifyPassword(oldPlaintext); err != nil {
		return ErrInvalidCredentials
	}
	if len(newPlaintext) < MinPasswordLen {
		return ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPlaintext), bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// Validate is a defensive check used after loading from the repository.
func (u *User) Validate() error {
	if u.ID <= 0 {
		return errors.New("user: invalid id")
	}
	if u.Username == "" {
		return errors.New("user: empty username")
	}
	if u.PasswordHash == "" {
		return errors.New("user: empty password hash")
	}
	return nil
}
