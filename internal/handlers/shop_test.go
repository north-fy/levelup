package handlers

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
	"github.com/north-fy/levelup/internal/services"
)

// handlerShopItemStore is an in-memory ShopItemStore for handler tests.
type handlerShopItemStore struct {
	mu        sync.Mutex
	users     *handlerUserStore
	nextID    uint
	byID      map[uint]*domain.ShopItem
	purchases []domain.Purchase
}

func newHandlerShopItemStore(users *handlerUserStore) *handlerShopItemStore {
	return &handlerShopItemStore{
		users: users,
		byID:  make(map[uint]*domain.ShopItem),
	}
}

func (f *handlerShopItemStore) Create(_ context.Context, item *domain.ShopItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	item.ID = f.nextID
	f.byID[item.ID] = item
	return nil
}

func (f *handlerShopItemStore) GetByID(_ context.Context, id uint) (*domain.ShopItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (f *handlerShopItemStore) GetByIDAndSeller(_ context.Context, id, sellerID uint) (*domain.ShopItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.byID[id]
	if !ok || item.SellerID != sellerID {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (f *handlerShopItemStore) ListActive(context.Context) ([]domain.ShopItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.ShopItem
	for _, item := range f.byID {
		if item.IsActive {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (f *handlerShopItemStore) ListByUser(context.Context, uint) ([]domain.ShopItem, error) {
	return nil, nil
}

func (f *handlerShopItemStore) Update(_ context.Context, item *domain.ShopItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[item.ID] = item
	return nil
}

func (f *handlerShopItemStore) Deactivate(_ context.Context, item *domain.ShopItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item.IsActive = false
	f.byID[item.ID] = item
	return nil
}

func (f *handlerShopItemStore) Buy(_ context.Context, itemID, buyerID uint) (*domain.Purchase, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.byID[itemID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if !item.IsActive {
		return nil, domain.ErrItemNotActive
	}
	if item.SellerID == buyerID {
		return nil, domain.ErrCannotBuyOwnItem
	}
	buyer, err := f.users.GetByID(context.Background(), buyerID)
	if err != nil {
		return nil, err
	}
	if buyer.Gold < item.PriceGold {
		return nil, domain.ErrNotEnoughGold
	}
	buyer.Gold -= item.PriceGold
	if err := f.users.Update(context.Background(), buyer); err != nil {
		return nil, err
	}
	purchase := &domain.Purchase{ItemID: item.ID, BuyerID: buyerID, SellerID: item.SellerID, Price: item.PriceGold}
	f.purchases = append(f.purchases, *purchase)
	return purchase, nil
}

// handlerPurchaseStore is an in-memory PurchaseStore for handler tests.
type handlerPurchaseStore struct {
	mu        sync.Mutex
	purchases []domain.Purchase
}

func (f *handlerPurchaseStore) ListByBuyer(_ context.Context, buyerID uint) ([]domain.Purchase, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.Purchase
	for _, p := range f.purchases {
		if p.BuyerID == buyerID {
			result = append(result, p)
		}
	}
	return result, nil
}

// handlerPurchasePublisher satisfies PurchaseEventPublisher in handler tests.
type handlerPurchasePublisher struct{}

func (handlerPurchasePublisher) PublishPurchase(context.Context, domain.PurchaseEvent) error {
	return nil
}

// setupShopRouter builds an authenticated router for shop endpoints.
// A "Magic Sword" priced 50 is pre-listed by the seller (user 2).
func setupShopRouter() *gin.Engine {
	users := newHandlerUserStore()
	buyer := &domain.User{Email: "buyer@example.com", Nickname: "Buyer", Gold: 200}
	seller := &domain.User{Email: "seller@example.com", Nickname: "Seller"}
	if err := users.Create(context.Background(), buyer); err != nil {
		panic(err)
	}
	if err := users.Create(context.Background(), seller); err != nil {
		panic(err)
	}

	items := newHandlerShopItemStore(users)
	purchases := &handlerPurchaseStore{}
	shopSvc := services.NewShopService(items, purchases, users, handlerPurchasePublisher{}, cache.Noop{})

	if _, err := shopSvc.Create(context.Background(), seller.ID, services.CreateShopItemInput{
		Title:     "Magic Sword",
		PriceGold: 50,
	}); err != nil {
		panic(err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", buyer.ID); c.Next() })
	handler := NewShopHandler(shopSvc)

	router.POST("/shop/items", handler.Create)
	router.GET("/shop/items", handler.List)
	router.PATCH("/shop/items/:id", handler.Update)
	router.DELETE("/shop/items/:id", handler.Delete)
	router.POST("/shop/items/:id/buy", handler.Buy)
	router.GET("/shop/purchases", handler.Purchases)
	return router
}

func TestShopCreateEndpoint(t *testing.T) {
	router := setupShopRouter()

	rec := postJSON(router, "/shop/items", `{"title":"Sword","price_gold":50}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
}

func TestShopBuyEndpoint(t *testing.T) {
	router := setupShopRouter()

	// The item is pre-listed by the seller; the authenticated buyer buys it.
	rec := postJSON(router, "/shop/items/1/buy", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
}

func TestShopBuyEndpointOwnItem(t *testing.T) {
	// The authenticated user both lists and buys, which must be rejected.
	router := buildShopRouterAsUser(1)

	create := postJSON(router, "/shop/items", `{"title":"Mine","price_gold":10}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create as seller failed: %d %s", create.Code, create.Body.String())
	}

	buy := postJSON(router, "/shop/items/1/buy", "")
	if buy.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", buy.Code, buy.Body.String())
	}
}

func buildShopRouterAsUser(userID uint) *gin.Engine {
	users := newHandlerUserStore()
	buyer := &domain.User{Email: "buyer@example.com", Nickname: "Buyer", Gold: 200}
	if err := users.Create(context.Background(), buyer); err != nil {
		panic(err)
	}

	items := newHandlerShopItemStore(users)
	purchases := &handlerPurchaseStore{}
	shopSvc := services.NewShopService(items, purchases, users, handlerPurchasePublisher{}, cache.Noop{})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", userID); c.Next() })
	handler := NewShopHandler(shopSvc)

	router.POST("/shop/items", handler.Create)
	router.POST("/shop/items/:id/buy", handler.Buy)
	return router
}
