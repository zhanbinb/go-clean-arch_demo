// Package jwt wraps golang-jwt/jwt v5 with a project-specific claim structure.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the payload embedded in every issued token.
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

// Manager signs and verifies tokens with a shared HMAC secret.
type Manager struct {
	secret      []byte
	ttl         time.Duration
	refreshTTL  time.Duration
	issuer      string
}

// New builds a Manager. secret should be at least 32 bytes for HMAC-SHA256.
func New(secret string, ttl, refreshTTL time.Duration) (*Manager, error) {
	if len(secret) < 16 {
		return nil, errors.New("jwt secret must be at least 16 bytes")
	}
	return &Manager{
		secret:     []byte(secret),
		ttl:        ttl,
		refreshTTL: refreshTTL,
		issuer:     "go-clean-arch-demo",
	}, nil
}

// Sign issues an access token for the given user.
func (m *Manager) Sign(userID int64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

// SignRefresh issues a longer-lived refresh token.
func (m *Manager) SignRefresh(userID int64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

// Parse verifies a token's signature and expiration and returns its claims.
func (m *Manager) Parse(token string) (*Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

// TTL returns the access-token time-to-live.
func (m *Manager) TTL() time.Duration { return m.ttl }
