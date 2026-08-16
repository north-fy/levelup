package services

import (
	"context"
	"strings"
	"time"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
)

const (
	maxShopTitleLen = 255
	maxShopDescLen  = 1024
)

// CreateShopItemInput holds the fields for creating a shop item.
type CreateShopItemInput struct {
	Title       string
	Description string
	PriceGold   int
}

// UpdateShopItemInput holds the optional fields for updating a shop item.
type UpdateShopItemInput struct {
	Title       *string
	Description *string
	PriceGold   *int
}

// ShopService manages the item market and purchases.
type ShopService struct {
	items     ShopItemStore
	purchases PurchaseStore
	users     UserStore
	events    PurchaseEventPublisher
	cache     cache.Cache
}

// NewShopService creates the shop service.
func NewShopService(items ShopItemStore, purchases PurchaseStore, users UserStore, events PurchaseEventPublisher, c cache.Cache) *ShopService {
	return &ShopService{
		items:     items,
		purchases: purchases,
		users:     users,
		events:    events,
		cache:     c,
	}
}

// Create lists a new item for sale.
func (s *ShopService) Create(ctx context.Context, sellerID uint, input CreateShopItemInput) (*domain.ShopItem, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, domain.NewValidationError("item title is required")
	}
	if len(title) > maxShopTitleLen {
		return nil, domain.NewValidationError("item title is too long")
	}
	if input.PriceGold < 0 {
		return nil, domain.NewValidationError("price cannot be negative")
	}

	item := &domain.ShopItem{
		SellerID:    sellerID,
		Title:       title,
		Description: input.Description,
		PriceGold:   input.PriceGold,
		IsActive:    true,
	}
	if err := s.items.Create(ctx, item); err != nil {
		return nil, err
	}
	//nolint:errcheck // cache invalidation is best-effort
	_ = s.cache.Del(ctx, cache.ShopItemsMineKey(sellerID))
	_ = s.cache.Del(ctx, cache.ShopItemsKey())
	return item, nil
}

// List returns active items, or the user's own items when mine is set.
func (s *ShopService) List(ctx context.Context, userID uint, mine bool) ([]domain.ShopItem, error) {
	if mine {
		return getOrSet(ctx, s.cache, cache.ShopItemsMineKey(userID), cache.ShopTTL, func() ([]domain.ShopItem, error) {
			return s.items.ListByUser(ctx, userID)
		})
	}
	return getOrSet(ctx, s.cache, cache.ShopItemsKey(), cache.ShopTTL, func() ([]domain.ShopItem, error) {
		return s.items.ListActive(ctx)
	})
}

// Update applies the provided changes to the seller's own item.
func (s *ShopService) Update(ctx context.Context, sellerID, itemID uint, input UpdateShopItemInput) (*domain.ShopItem, error) {
	item, err := s.items.GetByIDAndSeller(ctx, itemID, sellerID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, domain.NewValidationError("item title cannot be empty")
		}
		if len(title) > maxShopTitleLen {
			return nil, domain.NewValidationError("item title is too long")
		}
		item.Title = title
	}
	if input.Description != nil {
		item.Description = *input.Description
	}
	if input.PriceGold != nil {
		if *input.PriceGold < 0 {
			return nil, domain.NewValidationError("price cannot be negative")
		}
		item.PriceGold = *input.PriceGold
	}

	if err := s.items.Update(ctx, item); err != nil {
		return nil, err
	}
	//nolint:errcheck // cache invalidation is best-effort
	_ = s.cache.Del(ctx, cache.ShopItemsMineKey(sellerID))
	_ = s.cache.Del(ctx, cache.ShopItemsKey())
	return item, nil
}

// Delete deactivates the seller's own item, hiding it from the shop.
func (s *ShopService) Delete(ctx context.Context, sellerID, itemID uint) error {
	item, err := s.items.GetByIDAndSeller(ctx, itemID, sellerID)
	if err != nil {
		return err
	}
	if err := s.items.Deactivate(ctx, item); err != nil {
		return err
	}
	//nolint:errcheck // cache invalidation is best-effort
	_ = s.cache.Del(ctx, cache.ShopItemsMineKey(sellerID))
	_ = s.cache.Del(ctx, cache.ShopItemsKey())
	return nil
}

// Buy purchases an item: the buyer is debited and the seller credited atomically.
func (s *ShopService) Buy(ctx context.Context, buyerID, itemID uint) (*domain.Purchase, error) {
	item, err := s.items.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.SellerID == buyerID {
		return nil, domain.ErrCannotBuyOwnItem
	}
	if !item.IsActive {
		return nil, domain.ErrItemNotActive
	}

	buyer, err := s.users.GetByID(ctx, buyerID)
	if err != nil {
		return nil, err
	}
	if buyer.Gold < item.PriceGold {
		return nil, domain.ErrNotEnoughGold
	}

	purchase, err := s.items.Buy(ctx, itemID, buyerID)
	if err != nil {
		return nil, err
	}
	if err := s.events.PublishPurchase(ctx, domain.PurchaseEvent{
		UserID:      buyerID,
		ItemID:      itemID,
		Price:       item.PriceGold,
		PurchasedAt: time.Now(),
	}); err != nil {
		return nil, err
	}
	invalidateUser(ctx, s.cache, buyerID)
	invalidateUser(ctx, s.cache, item.SellerID)
	return purchase, nil
}

// PurchaseHistory returns the buyer's completed purchases.
func (s *ShopService) PurchaseHistory(ctx context.Context, buyerID uint) ([]domain.Purchase, error) {
	return s.purchases.ListByBuyer(ctx, buyerID)
}
