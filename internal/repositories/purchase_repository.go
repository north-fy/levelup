package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// PurchaseRepository persists purchase history.
type PurchaseRepository struct {
	db *gorm.DB
}

// NewPurchaseRepository creates the repository.
func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

// ListByBuyer returns the purchase history of a user ordered by date.
func (r *PurchaseRepository) ListByBuyer(ctx context.Context, buyerID uint) ([]domain.Purchase, error) {
	var purchases []domain.Purchase
	if err := r.db.WithContext(ctx).
		Where("buyer_id = ?", buyerID).
		Order("created_at DESC").
		Find(&purchases).Error; err != nil {
		return nil, err
	}
	return purchases, nil
}
