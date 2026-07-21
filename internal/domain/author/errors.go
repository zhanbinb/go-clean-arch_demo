package author

import "errors"

// Domain errors for the author aggregate.
var (
	ErrNotFound          = errors.New("author: not found")
	ErrAlreadyExists     = errors.New("author: already exists")
	ErrInvalidInput      = errors.New("author: invalid input")
	ErrEmailTaken        = errors.New("author: email already in use")
)
