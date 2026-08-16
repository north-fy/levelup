package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter decides whether a request identified by key is allowed.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// RedisLimiter is a fixed-window counter backed by Redis. The caller is
// responsible for embedding a window bucket in the key.
type RedisLimiter struct {
	rdb *redis.Client
}

// NewRedisLimiter creates a Redis-backed limiter.
func NewRedisLimiter(rdb *redis.Client) *RedisLimiter {
	return &RedisLimiter{rdb: rdb}
}

// Allow increments the counter for key and reports whether it stays within
// limit during the given window. The key must already encode the window
// bucket so that counters reset naturally as the window rolls over.
func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	res, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if res == 1 {
		//nolint:errcheck // best-effort expiry on first increment
		_ = r.rdb.Expire(ctx, key, window).Err()
	}
	return res <= int64(limit), nil
}
