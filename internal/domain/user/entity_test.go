package user

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := NewUser("alice", "supersecret")
	require.NoError(t, err)
	assert.Equal(t, "alice", u.Username)
	assert.NotEmpty(t, u.PasswordHash)
	assert.NotEqual(t, "supersecret", u.PasswordHash) // must be hashed
}

func TestNewUser_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		password string
	}{
		{"empty user", "", "longenough"},
		{"short password", "alice", "short"},
		{"too long user", strings.Repeat("x", MaxUsernameLen+1), "longenough"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser(tt.user, tt.password)
			assert.Error(t, err)
		})
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	u, _ := NewUser("alice", "supersecret")
	assert.NoError(t, u.VerifyPassword("supersecret"))
	assert.Error(t, u.VerifyPassword("wrong"))
}

func TestUser_ChangePassword(t *testing.T) {
	u, _ := NewUser("alice", "supersecret")
	require.NoError(t, u.ChangePassword("supersecret", "newpassword1"))
	assert.NoError(t, u.VerifyPassword("newpassword1"))
	assert.Error(t, u.VerifyPassword("supersecret"))
}

func TestUser_ChangePassword_WrongOld(t *testing.T) {
	u, _ := NewUser("alice", "supersecret")
	err := u.ChangePassword("wrong", "newpassword1")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}
