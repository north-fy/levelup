package ratelimit

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/redis"
)

func newTestLimiter(t *testing.T) *RedisLimiter {
	t.Helper()
	rdb, err := redis.New(config.Redis{Host: "localhost", Port: 6379})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return NewRedisLimiter(rdb)
}

func TestRedisLimiterAllowsWithinLimit(t *testing.T) {
	l := newTestLimiter(t)
	ctx := context.Background()
	key := "test:rl:" + strconv.FormatInt(time.Now().UnixNano(), 10)
	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup
		_ = l.rdb.Del(ctx, key)
	})

	for i := 0; i < 3; i++ {
		ok, err := l.Allow(ctx, key, 3, time.Minute)
		if err != nil {
			t.Fatalf("allow #%d: %v", i, err)
		}
		if !ok {
			t.Fatalf("allow #%d should be allowed", i)
		}
	}

	ok, err := l.Allow(ctx, key, 3, time.Minute)
	if err != nil {
		t.Fatalf("allow: %v", err)
	}
	if ok {
		t.Fatal("expected limit exceeded")
	}
}

func TestRedisLimiterDistinctKeys(t *testing.T) {
	l := newTestLimiter(t)
	ctx := context.Background()

	if ok, err := l.Allow(ctx, "test:rl:a:"+strconv.FormatInt(time.Now().UnixNano(), 10), 1, time.Minute); err != nil || !ok {
		t.Fatalf("first key should be allowed: %v %v", ok, err)
	}
	if ok, err := l.Allow(ctx, "test:rl:b:"+strconv.FormatInt(time.Now().UnixNano(), 10), 1, time.Minute); err != nil || !ok {
		t.Fatalf("second key should be allowed: %v %v", ok, err)
	}
}
