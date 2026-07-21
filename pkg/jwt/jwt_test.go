package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m, err := New("test-secret-must-be-at-least-16-bytes-long", time.Hour, 24*time.Hour)
	require.NoError(t, err)
	return m
}

func TestNew_ShortSecret(t *testing.T) {
	_, err := New("short", time.Hour, time.Hour)
	assert.Error(t, err)
}

func TestSignAndParse(t *testing.T) {
	m := newTestManager(t)
	tok, err := m.Sign(42, "alice")
	require.NoError(t, err)
	assert.NotEmpty(t, tok)

	claims, err := m.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
}

func TestParse_Invalid(t *testing.T) {
	m := newTestManager(t)
	_, err := m.Parse("garbage.token.here")
	assert.Error(t, err)
}

func TestParse_WrongSecret(t *testing.T) {
	m1 := newTestManager(t)
	m2, _ := New("different-secret-also-16-bytes-long", time.Hour, time.Hour)

	tok, _ := m1.Sign(1, "alice")
	_, err := m2.Parse(tok)
	assert.Error(t, err)
}

func TestSignRefresh(t *testing.T) {
	m := newTestManager(t)
	tok, err := m.SignRefresh(1, "alice")
	require.NoError(t, err)
	claims, err := m.Parse(tok)
	require.NoError(t, err)
	assert.True(t, claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) > time.Hour)
}
