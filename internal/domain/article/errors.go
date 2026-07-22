package article

import "errors"

// Domain errors for the article aggregate. These are converted to *errcode.Error
// at the application / interface boundaries.
var (
	ErrNotFound      = errors.New("article: not found")
	ErrAlreadyExists = errors.New("article: already exists")
	ErrInvalidInput  = errors.New("article: invalid input")
)
