package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/north-fy/levelup/internal/config"
)

// TokenType distinguishes access and refresh tokens.
type TokenType string

const (
	// AccessToken marks a short-lived access token.
	AccessToken TokenType = "access"
	// RefreshToken marks a long-lived refresh token.
	RefreshToken TokenType = "refresh"
)

// Claims carries the payload of a signed token.
type Claims struct {
	UserID    uint      `json:"uid"`
	TokenType TokenType `json:"typ"`
	jwt.RegisteredClaims
}

// Manager issues and parses signed JWT tokens.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// New creates a token manager from configuration.
func New(cfg config.JWT) *Manager {
	return &Manager{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

// AccessTTL returns the lifetime of access tokens.
func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}

// RefreshTTL returns the lifetime of refresh tokens.
func (m *Manager) RefreshTTL() time.Duration {
	return m.refreshTTL
}

// IssueAccess creates a new access token for the user.
func (m *Manager) IssueAccess(userID uint, tokenID string) (string, time.Time, error) {
	return m.issue(m.accessSecret, m.accessTTL, userID, tokenID, AccessToken)
}

// IssueRefresh creates a new refresh token for the user.
func (m *Manager) IssueRefresh(userID uint, tokenID string) (string, time.Time, error) {
	return m.issue(m.refreshSecret, m.refreshTTL, userID, tokenID, RefreshToken)
}

// ParseAccess validates an access token and returns its claims.
func (m *Manager) ParseAccess(token string) (*Claims, error) {
	return m.parse(m.accessSecret, AccessToken, token)
}

// ParseRefresh validates a refresh token and returns its claims.
func (m *Manager) ParseRefresh(token string) (*Claims, error) {
	return m.parse(m.refreshSecret, RefreshToken, token)
}

func (m *Manager) issue(secret []byte, ttl time.Duration, userID uint, tokenID string, tokenType TokenType) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (m *Manager) parse(secret []byte, expected TokenType, tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType != expected {
		return nil, errors.New("unexpected token type")
	}
	return claims, nil
}
