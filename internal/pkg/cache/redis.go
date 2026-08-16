package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-backed implementation of Cache.
type RedisCache struct {
	rdb *redis.Client
}

// NewRedisCache creates a Redis-backed cache.
func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{rdb: rdb}
}

// Get returns the raw value for key, or ErrMiss.
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrMiss
	}
	return val, err
}

// Set stores value under key with a TTL.
func (c *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Del removes a single key.
func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// DelPrefix removes every key starting with prefix.
func (c *RedisCache) DelPrefix(ctx context.Context, prefix string) error {
	keys, err := c.rdb.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}
