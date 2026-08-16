package services

import (
	"context"
	"errors"
	"strings"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
)

// UpdateProfileInput holds the editable profile fields.
type UpdateProfileInput struct {
	Nickname  *string
	Status    *string
	AvatarURL *string
}

// UserService manages user profile operations.
type UserService struct {
	users UserStore
	cache cache.Cache
}

// NewUserService creates the user service.
func NewUserService(users UserStore, c cache.Cache) *UserService {
	return &UserService{users: users, cache: c}
}

// GetByID returns a user by id, caching the profile.
func (s *UserService) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	user, err := getOrSet(ctx, s.cache, cache.ProfileKey(id), cache.ProfileTTL, func() (*domain.User, error) {
		u, err := s.users.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		u.Level = domain.LevelFor(u.XP)
		return u, nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update applies profile changes for a user.
func (s *UserService) Update(ctx context.Context, userID uint, input UpdateProfileInput) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Nickname != nil {
		nickname := strings.TrimSpace(*input.Nickname)
		if nickname == "" {
			return nil, errors.New("nickname cannot be empty")
		}
		user.Nickname = nickname
	}
	if input.Status != nil {
		user.Status = *input.Status
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	invalidateUser(ctx, s.cache, userID)
	user.Level = domain.LevelFor(user.XP)
	return user, nil
}
