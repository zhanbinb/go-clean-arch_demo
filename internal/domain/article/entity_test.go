package article

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewArticle_Valid(t *testing.T) {
	a, err := NewArticle("Hello", "World", 1, "Alice")
	require.NoError(t, err)
	assert.Equal(t, "Hello", a.Title)
	assert.Equal(t, "World", a.Content)
	assert.Equal(t, int64(1), a.AuthorID)
	assert.Equal(t, "Alice", a.AuthorName)
	assert.False(t, a.CreatedAt.IsZero())
	assert.Equal(t, a.CreatedAt, a.UpdatedAt)
}

func TestNewArticle_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		author  int64
		wantErr bool
	}{
		{"empty title", "", "content", 1, true},
		{"whitespace title", "   ", "content", 1, true},
		{"empty content", "title", "", 1, true},
		{"zero author", "title", "content", 0, true},
		{"negative author", "title", "content", -1, true},
		{"too long title", strings.Repeat("x", MaxTitleLen+1), "content", 1, true},
		{"valid", "ok", "ok", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewArticle(tt.title, tt.content, tt.author, "name")
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArticle_Update_Partial(t *testing.T) {
	a, _ := NewArticle("old", "old content", 1, "alice")
	originalUpdated := a.UpdatedAt

	newTitle := "new"
	err := a.Update(&newTitle, nil)
	require.NoError(t, err)
	assert.Equal(t, "new", a.Title)
	assert.Equal(t, "old content", a.Content)
	assert.True(t, a.UpdatedAt.After(originalUpdated) || a.UpdatedAt.Equal(originalUpdated))
}

func TestArticle_Update_Empty(t *testing.T) {
	a, _ := NewArticle("ok", "ok", 1, "alice")
	err := a.Update(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", a.Title)
}

func TestArticle_Update_Invalid(t *testing.T) {
	a, _ := NewArticle("ok", "ok", 1, "alice")
	empty := ""
	err := a.Update(&empty, nil)
	assert.ErrorIs(t, err, ErrInvalidInput)
}
