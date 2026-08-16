package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// BranchRepository persists branches in PostgreSQL.
type BranchRepository struct {
	db *gorm.DB
}

// NewBranchRepository creates the repository.
func NewBranchRepository(db *gorm.DB) *BranchRepository {
	return &BranchRepository{db: db}
}

// Create stores a new branch.
func (r *BranchRepository) Create(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Create(branch).Error
}

// GetByIDAndUser returns a branch owned by the given user.
func (r *BranchRepository) GetByIDAndUser(ctx context.Context, id, userID uint) (*domain.Branch, error) {
	var branch domain.Branch
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&branch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &branch, nil
}

// ListByUser returns all branches of a user ordered by creation.
func (r *BranchRepository) ListByUser(ctx context.Context, userID uint) ([]domain.Branch, error) {
	var branches []domain.Branch
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

// Update persists branch changes.
func (r *BranchRepository) Update(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

// Delete removes a branch.
func (r *BranchRepository) Delete(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Delete(branch).Error
}
