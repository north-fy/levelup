package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// UserRepository persists users in PostgreSQL.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates the repository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create stores a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID returns a user by primary key.
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail returns a user by email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByGitHubID returns a user by their linked GitHub id.
func (r *UserRepository) GetByGitHubID(ctx context.Context, githubID string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("github_id = ?", githubID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update persists profile changes for a user.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// IsEmailTaken reports whether the email belongs to a different user.
func (r *UserRepository) IsEmailTaken(ctx context.Context, email string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.User{}).Where("email = ?", email)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
