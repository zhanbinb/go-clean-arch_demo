package errcode

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_HTTPStatus(t *testing.T) {
	tests := []struct {
		err        *Error
		wantStatus int
	}{
		{ErrBadRequest, http.StatusBadRequest},
		{ErrUnauthorized, http.StatusUnauthorized},
		{ErrNotFound, http.StatusNotFound},
		{ErrConflict, http.StatusConflict},
		{ErrTooManyRequests, http.StatusTooManyRequests},
		{ErrInternal, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.err.Message, func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, tt.err.HTTPStatus())
		})
	}
}

func TestError_WithCause(t *testing.T) {
	root := errors.New("db down")
	e := ErrInternal.WithCause(root)
	assert.ErrorIs(t, e, root)
	assert.Contains(t, e.Error(), "db down")
}

func TestError_WithMsg(t *testing.T) {
	e := ErrBadRequest.WithMsg("title too long")
	assert.Equal(t, "title too long", e.Message)
	assert.Equal(t, ErrBadRequest.Code, e.Code)
}

func TestFromError(t *testing.T) {
	assert.Equal(t, OK, FromError(nil))

	e := ErrNotFound
	assert.Same(t, e, FromError(e))

	plain := errors.New("random")
	wrapped := FromError(plain)
	assert.ErrorIs(t, wrapped, plain)
	assert.Equal(t, ErrInternal.Code, wrapped.Code)
}
