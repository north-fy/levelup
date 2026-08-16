package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/north-fy/levelup/internal/pkg/cache"
)

// getOrSet returns the cached value for key, or computes it with fn and
// populates the cache with a TTL.
func getOrSet[T any](ctx context.Context, c cache.Cache, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	var zero T

	if raw, err := c.Get(ctx, key); err == nil {
		var v T
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			return v, nil
		}
	}

	v, err := fn()
	if err != nil {
		return zero, err
	}

	if data, err := json.Marshal(v); err == nil {
		//nolint:errcheck // caching is best-effort
		_ = c.Set(ctx, key, string(data), ttl)
	}
	return v, nil
}

// invalidateUser drops a user's cached profile and overview stats, typically
// after their XP or gold changed.
func invalidateUser(ctx context.Context, c cache.Cache, userID uint) {
	//nolint:errcheck // cache invalidation is best-effort
	_ = c.Del(ctx, cache.ProfileKey(userID))
	_ = c.Del(ctx, cache.OverviewKey(userID))
}
