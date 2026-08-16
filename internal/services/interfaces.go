package services

import (
	"context"
	"time"

	"github.com/north-fy/levelup/internal/domain"
)

// UserStore abstracts user persistence.
type UserStore interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	IsEmailTaken(ctx context.Context, email string, excludeID uint) (bool, error)
}

// TokenStore abstracts ephemeral auth data persistence.
type TokenStore interface {
	SaveRefresh(ctx context.Context, tokenID string, userID uint, ttl time.Duration) error
	GetRefreshUser(ctx context.Context, tokenID string) (uint, error)
	DeleteRefresh(ctx context.Context, tokenID string) error
	BlacklistAccess(ctx context.Context, tokenID string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	SaveOAuthState(ctx context.Context, state string, ttl time.Duration) error
	ValidateOAuthState(ctx context.Context, state string) error
}
