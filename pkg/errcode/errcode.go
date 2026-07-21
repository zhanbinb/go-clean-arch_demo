// Package errcode defines the unified business error code system used across
// REST and gRPC interfaces. Codes are 5-digit integers: first 3 = HTTP status,
// last 2 = business-specific.
package errcode

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is the application-wide error code.
type Code int

// Error is a structured business error carrying a Code, HTTP status and message.
type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.cause }

// WithCause attaches a root cause (e.g. DB error) for logging without leaking
// it to the client.
func (e *Error) WithCause(cause error) *Error {
	clone := *e
	clone.cause = cause
	return &clone
}

// WithMsg replaces the human-readable message.
func (e *Error) WithMsg(msg string) *Error {
	clone := *e
	clone.Message = msg
	return &clone
}

// HTTPStatus returns the recommended HTTP status for this error.
func (e *Error) HTTPStatus() int { return httpStatusFromCode(int(e.Code)) }

// New builds a custom Error with explicit code and message.
func New(code int, msg string) *Error {
	return &Error{Code: Code(code), Message: msg}
}

// ----------------------------------------------------------------------
// Predefined codes
// ----------------------------------------------------------------------

var (
	// OK is the success code (never actually returned, used for symmetry).
	OK = New(0, "OK")

	// 400xx — client errors
	ErrBadRequest          = New(40000, "bad request")
	ErrInvalidParam        = New(40001, "invalid parameter")
	ErrValidation          = New(40002, "validation failed")
	ErrUnauthorized        = New(40100, "unauthorized")
	ErrTokenExpired        = New(40101, "token expired")
	ErrTokenInvalid        = New(40102, "token invalid")
	ErrForbidden           = New(40300, "forbidden")
	ErrNotFound            = New(40400, "not found")
	ErrConflict            = New(40900, "conflict")
	ErrTooManyRequests     = New(42900, "too many requests")

	// 500xx — server errors
	ErrInternal            = New(50000, "internal server error")
	ErrServiceUnavailable  = New(50300, "service unavailable")
)

// httpStatusFromCode extracts HTTP status from a 5-digit error code.
// First 3 digits map to standard HTTP status.
func httpStatusFromCode(code int) int {
	switch {
	case code >= 40000 && code < 50000:
		first := code / 100
		if first >= 400 && first < 500 {
			return first
		}
		return http.StatusBadRequest
	case code >= 50000 && code < 60000:
		first := code / 100
		if first >= 500 && first < 600 {
			return first
		}
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

// FromError converts any error to an *Error. If err is already *Error, it is
// returned as-is. Otherwise it is wrapped into ErrInternal.
func FromError(err error) *Error {
	if err == nil {
		return OK
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrInternal.WithCause(err)
}
