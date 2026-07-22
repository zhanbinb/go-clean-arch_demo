package user

import "errors"

// Domain errors for the user aggregate.
var (
	ErrNotFound            = errors.New("user: not found")
	ErrAlreadyExists       = errors.New("user: already exists")
	ErrInvalidInput        = errors.New("user: invalid input")
	ErrWeakPassword        = errors.New("user: password too weak (min 8 chars)")
	ErrInvalidCredentials  = errors.New("user: invalid username or password")
)
