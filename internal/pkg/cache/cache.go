package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrMiss is returned when a key is not present in the cache.
var ErrMiss = errors.New("cache miss")

// Cache abstracts a key/value cache with TTL support.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	DelPrefix(ctx context.Context, prefix string) error
}

// TTLs used for cached data.
const (
	ProfileTTL  = 5 * time.Minute
	ShopTTL     = 5 * time.Minute
	WorkshopTTL = 5 * time.Minute
	StatsTTL    = time.Minute
)

const (
	profileKey      = "users:me:%d"
	shopKey         = "shop:items"
	shopMineKey     = "shop:items:mine:%d"
	workshopKey     = "workshop:published"
	workshopMineKey = "workshop:mine:%d"
	overviewKey     = "stats:overview:%d"
)

// ProfileKey returns the cache key for a user profile.
func ProfileKey(id uint) string {
	return fmt.Sprintf(profileKey, id)
}

// ShopItemsKey returns the cache key for the active shop item list.
func ShopItemsKey() string {
	return shopKey
}

// ShopItemsMineKey returns the cache key for a user's own items.
func ShopItemsMineKey(id uint) string {
	return fmt.Sprintf(shopMineKey, id)
}

// WorkshopKey returns the cache key for published workshop roadmaps.
func WorkshopKey() string {
	return workshopKey
}

// WorkshopMineKey returns the cache key for a user's own workshop roadmaps.
func WorkshopMineKey(id uint) string {
	return fmt.Sprintf(workshopMineKey, id)
}

// OverviewKey returns the cache key for a user's overview stats.
func OverviewKey(id uint) string {
	return fmt.Sprintf(overviewKey, id)
}
