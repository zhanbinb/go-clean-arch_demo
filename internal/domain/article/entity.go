package article

import (
	"errors"
	"strings"
	"time"
)

// Article is the aggregate root for the article domain.
//
// It does NOT hold a full Author entity; it stores AuthorID plus a
// AuthorName snapshot (denormalised for read efficiency). For mutations,
// callers re-fetch author data via the author repository.
type Article struct {
	ID         int64
	Title      string
	Content    string
	AuthorID   int64
	AuthorName string // snapshot, may be empty when freshly created
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// MaxTitleLen is the maximum allowed article title length.
const MaxTitleLen = 200

// NewArticle is a factory that enforces invariants on creation.
// Returns ErrInvalidInput if validation fails.
func NewArticle(title, content string, authorID int64, authorName string) (*Article, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return nil, ErrInvalidInput
	}
	if len(title) > MaxTitleLen {
		return nil, ErrInvalidInput
	}
	if content == "" {
		return nil, ErrInvalidInput
	}
	if authorID <= 0 {
		return nil, ErrInvalidInput
	}
	now := time.Now()
	return &Article{
		Title:      title,
		Content:    content,
		AuthorID:   authorID,
		AuthorName: authorName,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Update applies new content to the article, refreshing UpdatedAt.
// Empty values mean "leave unchanged".
func (a *Article) Update(title, content *string) error {
	if title != nil {
		t := strings.TrimSpace(*title)
		if t == "" || len(t) > MaxTitleLen {
			return ErrInvalidInput
		}
		a.Title = t
	}
	if content != nil {
		c := strings.TrimSpace(*content)
		if c == "" {
			return ErrInvalidInput
		}
		a.Content = c
	}
	a.UpdatedAt = time.Now()
	return nil
}

// AssignAuthor changes the article's author (used when re-publishing under a
// different author). AuthorName should be the new author's display name.
func (a *Article) AssignAuthor(authorID int64, authorName string) error {
	if authorID <= 0 {
		return ErrInvalidInput
	}
	a.AuthorID = authorID
	a.AuthorName = authorName
	a.UpdatedAt = time.Now()
	return nil
}

// Validate is a defensive check used after loading from the repository.
func (a *Article) Validate() error {
	if a.ID <= 0 {
		return errors.New("article: invalid id")
	}
	if a.Title == "" {
		return errors.New("article: empty title")
	}
	if a.AuthorID <= 0 {
		return errors.New("article: invalid author id")
	}
	return nil
}
