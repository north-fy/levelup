package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/jwt"
)

func newTestAuthService(users UserStore, tokens TokenStore) *AuthService {
	return NewAuthService(users, tokens, jwt.New(config.JWT{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    720 * time.Hour,
	}), config.GitHub{})
}

func TestRegister(t *testing.T) {
	users := newFakeUserStore()
	tokens := newFakeTokenStore()
	svc := newTestAuthService(users, tokens)

	user, pair, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected assigned user id")
	}
	if user.Nickname != "Hero" {
		t.Errorf("nickname = %q, want %q", user.Nickname, "Hero")
	}
	if user.Level != 1 {
		t.Errorf("level = %d, want 1", user.Level)
	}
	if user.PasswordHash == "password123" {
		t.Error("password must be hashed")
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("expected tokens to be issued")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	users := newFakeUserStore()
	svc := newTestAuthService(users, newFakeTokenStore())

	if _, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero"); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero2")
	if !errors.Is(err, domain.ErrEmailAlreadyUsed) {
		t.Errorf("err = %v, want ErrEmailAlreadyUsed", err)
	}
}

func TestLogin(t *testing.T) {
	users := newFakeUserStore()
	svc := newTestAuthService(users, newFakeTokenStore())

	if _, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	user, pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if user.Nickname != "Hero" {
		t.Errorf("nickname = %q", user.Nickname)
	}
	if pair.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	users := newFakeUserStore()
	svc := newTestAuthService(users, newFakeTokenStore())

	if _, _, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, _, err := svc.Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestRefreshRotation(t *testing.T) {
	users := newFakeUserStore()
	tokens := newFakeTokenStore()
	svc := newTestAuthService(users, tokens)

	_, first, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	user, rotated, err := svc.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected user from refresh")
	}
	if rotated.RefreshToken == first.RefreshToken {
		t.Error("refresh token must be rotated")
	}

	// the old refresh token must no longer be valid
	if _, _, err := svc.Refresh(context.Background(), first.RefreshToken); !errors.Is(err, domain.ErrTokenRevoked) {
		t.Errorf("old refresh token err = %v, want ErrTokenRevoked", err)
	}
}

func TestRefreshInvalidToken(t *testing.T) {
	svc := newTestAuthService(newFakeUserStore(), newFakeTokenStore())

	_, _, err := svc.Refresh(context.Background(), "not-a-valid-token")
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestLogoutRevokesTokens(t *testing.T) {
	users := newFakeUserStore()
	tokens := newFakeTokenStore()
	svc := newTestAuthService(users, tokens)

	_, pair, err := svc.Register(context.Background(), "user@example.com", "password123", "Hero")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if err := svc.Logout(context.Background(), pair.AccessToken, pair.RefreshToken); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// access token is blacklisted
	claims, err := svc.jwt.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	blacklisted, err := tokens.IsBlacklisted(context.Background(), claims.ID)
	if err != nil {
		t.Fatalf("is blacklisted: %v", err)
	}
	if !blacklisted {
		t.Error("expected access token to be blacklisted")
	}

	// refresh token is revoked
	if _, _, err := svc.Refresh(context.Background(), pair.RefreshToken); !errors.Is(err, domain.ErrTokenRevoked) {
		t.Errorf("refresh after logout err = %v, want ErrTokenRevoked", err)
	}
}

func TestOAuthState(t *testing.T) {
	svc := newTestAuthService(newFakeUserStore(), newFakeTokenStore())

	state := svc.NewOAuthState()
	if err := svc.SaveOAuthState(context.Background(), state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	if err := svc.ValidateOAuthState(context.Background(), state); err != nil {
		t.Fatalf("validate state: %v", err)
	}
	// state is consumed after one use
	if err := svc.ValidateOAuthState(context.Background(), state); !errors.Is(err, domain.ErrInvalidState) {
		t.Errorf("reuse state err = %v, want ErrInvalidState", err)
	}
}
