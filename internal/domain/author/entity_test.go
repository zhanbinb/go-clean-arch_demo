package author

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthor_Valid(t *testing.T) {
	a, err := NewAuthor("Alice", "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "Alice", a.Name)
	assert.Equal(t, "alice@example.com", a.Email)
}

func TestNewAuthor_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		email string
	}{
		{"empty name", "", "a@b.com"},
		{"invalid email", "alice", "not-an-email"},
		{"empty email", "alice", ""},
		{"too long name", strings.Repeat("x", MaxNameLen+1), "a@b.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuthor(tt.in, tt.email)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}

func TestAuthor_Rename(t *testing.T) {
	a, _ := NewAuthor("Alice", "a@b.com")
	require.NoError(t, a.Rename("Bob"))
	assert.Equal(t, "Bob", a.Name)
}

func TestAuthor_ChangeEmail(t *testing.T) {
	a, _ := NewAuthor("Alice", "a@b.com")
	require.NoError(t, a.ChangeEmail("c@d.com"))
	assert.Equal(t, "c@d.com", a.Email)
}
