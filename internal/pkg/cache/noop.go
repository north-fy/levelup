package cache

import (
	"context"
	"time"
)

// Noop is a Cache that never stores anything. It is used in tests and as a
// safe default when Redis is unavailable.
type Noop struct{}

// Get always reports a miss.
func (Noop) Get(ctx context.Context, key string) (string, error) {
	return "", ErrMiss
}

// Set is a no-op.
func (Noop) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return nil
}

// Del is a no-op.
func (Noop) Del(ctx context.Context, key string) error {
	return nil
}

// DelPrefix is a no-op.
func (Noop) DelPrefix(ctx context.Context, prefix string) error {
	return nil
}
