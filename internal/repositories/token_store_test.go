package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/redis"
)

func newTestTokenStore(t *testing.T) *TokenStore {
	t.Helper()
	rdb, err := redis.New(config.Redis{Host: "localhost", Port: 6379})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return NewTokenStore(rdb)
}

func TestTokenStoreRefreshLifecycle(t *testing.T) {
	store := newTestTokenStore(t)
	ctx := context.Background()

	if err := store.SaveRefresh(ctx, "rt-1", 42, time.Minute); err != nil {
		t.Fatalf("save: %v", err)
	}

	userID, err := store.GetRefreshUser(ctx, "rt-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if userID != 42 {
		t.Errorf("user id = %d, want 42", userID)
	}

	if err := store.DeleteRefresh(ctx, "rt-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.GetRefreshUser(ctx, "rt-1"); err != domain.ErrNotFound {
		t.Errorf("get after delete = %v, want ErrNotFound", err)
	}
}

func TestTokenStoreBlacklist(t *testing.T) {
	store := newTestTokenStore(t)
	ctx := context.Background()

	if err := store.BlacklistAccess(ctx, "at-1", time.Minute); err != nil {
		t.Fatalf("blacklist: %v", err)
	}

	revoked, err := store.IsBlacklisted(ctx, "at-1")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !revoked {
		t.Error("expected token to be blacklisted")
	}

	revoked, err = store.IsBlacklisted(ctx, "at-2")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if revoked {
		t.Error("unexpected blacklist hit")
	}
}

func TestTokenStoreOAuthState(t *testing.T) {
	store := newTestTokenStore(t)
	ctx := context.Background()

	if err := store.SaveOAuthState(ctx, "state-1", time.Minute); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := store.ValidateOAuthState(ctx, "state-1"); err != nil {
		t.Fatalf("validate: %v", err)
	}
	// consumed after one use
	if err := store.ValidateOAuthState(ctx, "state-1"); err != domain.ErrInvalidState {
		t.Errorf("reuse = %v, want ErrInvalidState", err)
	}
}
