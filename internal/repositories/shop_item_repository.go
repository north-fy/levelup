package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/north-fy/levelup/internal/domain"
)

// ShopItemRepository persists shop items and executes the buy transaction.
type ShopItemRepository struct {
	db *gorm.DB
}

// NewShopItemRepository creates the repository.
func NewShopItemRepository(db *gorm.DB) *ShopItemRepository {
	return &ShopItemRepository{db: db}
}

// Create stores a new shop item.
func (r *ShopItemRepository) Create(ctx context.Context, item *domain.ShopItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetByID returns an item by id.
func (r *ShopItemRepository) GetByID(ctx context.Context, id uint) (*domain.ShopItem, error) {
	var item domain.ShopItem
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// GetByIDAndSeller returns an item owned by the given seller.
func (r *ShopItemRepository) GetByIDAndSeller(ctx context.Context, id, sellerID uint) (*domain.ShopItem, error) {
	var item domain.ShopItem
	if err := r.db.WithContext(ctx).Where("id = ? AND seller_id = ?", id, sellerID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// ListActive returns all items currently available for purchase.
func (r *ShopItemRepository) ListActive(ctx context.Context) ([]domain.ShopItem, error) {
	var items []domain.ShopItem
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ListByUser returns all items listed by a user, including inactive ones.
func (r *ShopItemRepository) ListByUser(ctx context.Context, userID uint) ([]domain.ShopItem, error) {
	var items []domain.ShopItem
	if err := r.db.WithContext(ctx).Where("seller_id = ?", userID).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Update persists item changes.
func (r *ShopItemRepository) Update(ctx context.Context, item *domain.ShopItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Deactivate removes an item from the shop while keeping its purchase history.
func (r *ShopItemRepository) Deactivate(ctx context.Context, item *domain.ShopItem) error {
	return r.db.WithContext(ctx).Model(item).Update("is_active", false).Error
}

// Buy atomically debits the buyer, credits the seller and records the purchase.
func (r *ShopItemRepository) Buy(ctx context.Context, itemID, buyerID uint) (*domain.Purchase, error) {
	var purchase domain.Purchase

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item domain.ShopItem
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", itemID).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		if !item.IsActive {
			return domain.ErrItemNotActive
		}
		if item.SellerID == buyerID {
			return domain.ErrCannotBuyOwnItem
		}

		var buyer domain.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", buyerID).First(&buyer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		if buyer.Gold < item.PriceGold {
			return domain.ErrNotEnoughGold
		}

		var seller domain.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", item.SellerID).First(&seller).Error; err != nil {
			return err
		}

		buyer.Gold -= item.PriceGold
		seller.Gold += item.PriceGold
		if err := tx.Save(&buyer).Error; err != nil {
			return err
		}
		if err := tx.Save(&seller).Error; err != nil {
			return err
		}

		purchase = domain.Purchase{
			ItemID:   item.ID,
			BuyerID:  buyerID,
			SellerID: item.SellerID,
			Price:    item.PriceGold,
		}
		return tx.Create(&purchase).Error
	})
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}
