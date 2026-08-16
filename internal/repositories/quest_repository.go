package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// QuestRepository persists quests in PostgreSQL.
type QuestRepository struct {
	db *gorm.DB
}

// NewQuestRepository creates the repository.
func NewQuestRepository(db *gorm.DB) *QuestRepository {
	return &QuestRepository{db: db}
}

// Create stores a new quest.
func (r *QuestRepository) Create(ctx context.Context, quest *domain.Quest) error {
	return r.db.WithContext(ctx).Create(quest).Error
}

// GetByIDAndUser returns a quest owned by the given user.
func (r *QuestRepository) GetByIDAndUser(ctx context.Context, id, userID uint) (*domain.Quest, error) {
	var quest domain.Quest
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&quest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &quest, nil
}

// ListByBranchAndUser returns quests of a branch owned by the user.
func (r *QuestRepository) ListByBranchAndUser(ctx context.Context, branchID, userID uint) ([]domain.Quest, error) {
	var quests []domain.Quest
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND user_id = ?", branchID, userID).
		Order("created_at ASC").
		Find(&quests).Error; err != nil {
		return nil, err
	}
	return quests, nil
}

// Update persists quest changes.
func (r *QuestRepository) Update(ctx context.Context, quest *domain.Quest) error {
	return r.db.WithContext(ctx).Save(quest).Error
}

// Delete removes a quest.
func (r *QuestRepository) Delete(ctx context.Context, quest *domain.Quest) error {
	return r.db.WithContext(ctx).Delete(quest).Error
}

// HasActiveTimer reports whether the user already has a quest in progress.
func (r *QuestRepository) HasActiveTimer(ctx context.Context, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Quest{}).
		Where("user_id = ? AND status = ?", userID, domain.QuestStatusInProgress).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountDone returns the number of completed quests in a branch.
func (r *QuestRepository) CountDone(ctx context.Context, branchID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Quest{}).
		Where("branch_id = ? AND status = ?", branchID, domain.QuestStatusDone).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
