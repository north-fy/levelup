package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/north-fy/levelup/internal/domain"
)

const (
	refreshPrefix = "refresh:"
	blackPrefix   = "blacklist:"
	statePrefix   = "oauth:state:"
)

// TokenStore persists auth-related ephemeral data in Redis.
type TokenStore struct {
	rdb *redis.Client
}

// NewTokenStore creates the token store.
func NewTokenStore(rdb *redis.Client) *TokenStore {
	return &TokenStore{rdb: rdb}
}

// SaveRefresh stores a refresh token mapping its id to a user.
func (s *TokenStore) SaveRefresh(ctx context.Context, tokenID string, userID uint, ttl time.Duration) error {
	return s.rdb.Set(ctx, refreshPrefix+tokenID, strconv.FormatUint(uint64(userID), 10), ttl).Err()
}

// GetRefreshUser returns the user associated with a refresh token.
func (s *TokenStore) GetRefreshUser(ctx context.Context, tokenID string) (uint, error) {
	val, err := s.rdb.Get(ctx, refreshPrefix+tokenID).Result()
	if errors.Is(err, redis.Nil) {
		return 0, domain.ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	userID, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(userID), nil
}

// DeleteRefresh removes a stored refresh token.
func (s *TokenStore) DeleteRefresh(ctx context.Context, tokenID string) error {
	return s.rdb.Del(ctx, refreshPrefix+tokenID).Err()
}

// BlacklistAccess marks an access token as revoked until it expires.
func (s *TokenStore) BlacklistAccess(ctx context.Context, tokenID string, ttl time.Duration) error {
	return s.rdb.Set(ctx, blackPrefix+tokenID, "1", ttl).Err()
}

// IsBlacklisted reports whether an access token has been revoked.
func (s *TokenStore) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	n, err := s.rdb.Exists(ctx, blackPrefix+tokenID).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SaveOAuthState stores a GitHub OAuth state value.
func (s *TokenStore) SaveOAuthState(ctx context.Context, state string, ttl time.Duration) error {
	return s.rdb.Set(ctx, statePrefix+state, "1", ttl).Err()
}

// ValidateOAuthState verifies and consumes a GitHub OAuth state value.
func (s *TokenStore) ValidateOAuthState(ctx context.Context, state string) error {
	key := statePrefix + state
	n, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrInvalidState
	}
	return s.rdb.Del(ctx, key).Err()
}
