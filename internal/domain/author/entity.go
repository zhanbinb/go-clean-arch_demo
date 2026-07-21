package author

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// Author is the aggregate root for the author domain.
type Author struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MaxNameLen is the maximum allowed author name length.
const MaxNameLen = 100

// emailRegex is intentionally simple; production code may want full RFC 5322.
var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// NewAuthor is a factory that enforces invariants.
func NewAuthor(name, email string) (*Author, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || len(name) > MaxNameLen {
		return nil, ErrInvalidInput
	}
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidInput
	}
	now := time.Now()
	return &Author{Name: name, Email: email, CreatedAt: now, UpdatedAt: now}, nil
}

// Rename updates the author's display name.
func (a *Author) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > MaxNameLen {
		return ErrInvalidInput
	}
	a.Name = name
	a.UpdatedAt = time.Now()
	return nil
}

// ChangeEmail updates the author's email.
func (a *Author) ChangeEmail(email string) error {
	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidInput
	}
	a.Email = email
	a.UpdatedAt = time.Now()
	return nil
}

// Validate is a defensive check used after loading from the repository.
func (a *Author) Validate() error {
	if a.ID <= 0 {
		return errors.New("author: invalid id")
	}
	if a.Name == "" {
		return errors.New("author: empty name")
	}
	if !emailRegex.MatchString(a.Email) {
		return errors.New("author: invalid email")
	}
	return nil
}
