package jwt

import (
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/config"
)

func newTestManager() *Manager {
	return New(config.JWT{
		AccessSecret:  "access-secret",
		RefreshSecret: "refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	})
}

func TestIssueAndParseAccess(t *testing.T) {
	m := newTestManager()

	token, expiresAt, err := m.IssueAccess(42, "token-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Error("expected future expiry")
	}

	claims, err := m.ParseAccess(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("user id = %d, want 42", claims.UserID)
	}
	if claims.TokenType != AccessToken {
		t.Errorf("token type = %q", claims.TokenType)
	}
	if claims.ID != "token-1" {
		t.Errorf("token id = %q", claims.ID)
	}
}

func TestRefreshTokenCannotBeUsedAsAccess(t *testing.T) {
	m := newTestManager()

	refresh, _, err := m.IssueRefresh(7, "refresh-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := m.ParseAccess(refresh); err == nil {
		t.Error("expected error parsing refresh token as access token")
	}
}

func TestAccessTokenCannotBeUsedAsRefresh(t *testing.T) {
	m := newTestManager()

	access, _, err := m.IssueAccess(7, "access-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := m.ParseRefresh(access); err == nil {
		t.Error("expected error parsing access token as refresh token")
	}
}

func TestParseInvalidToken(t *testing.T) {
	m := newTestManager()

	if _, err := m.ParseAccess("garbage"); err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestTamperedTokenRejected(t *testing.T) {
	m := newTestManager()

	token, _, err := m.IssueAccess(1, "access-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	tampered := token + "x"
	if _, err := m.ParseAccess(tampered); err == nil {
		t.Error("expected error for tampered token")
	}
}
