package cache

import (
	"context"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/redis"
)

func newTestCache(t *testing.T) *RedisCache {
	t.Helper()
	rdb, err := redis.New(config.Redis{Host: "localhost", Port: 6379})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return NewRedisCache(rdb)
}

func TestRedisCacheSetGetDel(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	key := "test:cache:setget"

	if err := c.Set(ctx, key, "hello", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	t.Cleanup(func() { _ = c.Del(ctx, key) })

	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want hello", got)
	}

	if err := c.Del(ctx, key); err != nil {
		t.Fatalf("del: %v", err)
	}
	if _, err := c.Get(ctx, key); err != ErrMiss {
		t.Fatalf("get after del = %v, want ErrMiss", err)
	}
}

func TestRedisCacheMiss(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	if _, err := c.Get(ctx, "test:cache:missing"); err != ErrMiss {
		t.Fatalf("got %v, want ErrMiss", err)
	}
}

func TestRedisCacheDelPrefix(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	prefix := "test:cache:prefix:"

	t.Cleanup(func() { _ = c.DelPrefix(ctx, prefix) })

	for i := 0; i < 3; i++ {
		if err := c.Set(ctx, prefix+"k"+string(rune('a'+i)), "v", time.Minute); err != nil {
			t.Fatalf("set: %v", err)
		}
	}

	if err := c.DelPrefix(ctx, prefix); err != nil {
		t.Fatalf("del prefix: %v", err)
	}
	if _, err := c.Get(ctx, prefix+"ka"); err != ErrMiss {
		t.Fatalf("expected miss after prefix delete, got %v", err)
	}
}

func TestNoopNeverStores(t *testing.T) {
	var c Noop
	ctx := context.Background()

	if _, err := c.Get(ctx, "k"); err != ErrMiss {
		t.Fatalf("get = %v, want ErrMiss", err)
	}
	if err := c.Set(ctx, "k", "v", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := c.Get(ctx, "k"); err != ErrMiss {
		t.Fatalf("noop must not store, got %v", err)
	}
	if err := c.Del(ctx, "k"); err != nil {
		t.Fatalf("del: %v", err)
	}
	if err := c.DelPrefix(ctx, "k"); err != nil {
		t.Fatalf("del prefix: %v", err)
	}
}
