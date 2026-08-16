package domain

import "time"

// ShopItem is an item listed for sale by a user.
type ShopItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SellerID    uint      `gorm:"index" json:"seller_id"`
	Title       string    `gorm:"size:255" json:"title"`
	Description string    `gorm:"size:1024" json:"description"`
	PriceGold   int       `json:"price_gold"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Purchase records a completed shop transaction.
type Purchase struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ItemID    uint      `gorm:"index" json:"item_id"`
	BuyerID   uint      `gorm:"index" json:"buyer_id"`
	SellerID  uint      `gorm:"index" json:"seller_id"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

// PurchaseEvent is emitted when an item is bought.
type PurchaseEvent struct {
	UserID      uint
	ItemID      uint
	Price       int
	PurchasedAt time.Time
}
