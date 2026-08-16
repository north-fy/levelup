package services

import (
	"context"
	"testing"

	"github.com/north-fy/levelup/internal/domain"
)

func newShopTestEnv(t *testing.T) (*ShopService, *fakeUserStore, *fakeShopItemStore, *recordingPurchasePublisher) {
	t.Helper()

	users := newFakeUserStore()
	buyer := &domain.User{Email: "buyer@test.dev", Gold: 200}
	seller := &domain.User{Email: "seller@test.dev", Gold: 10}
	if err := users.Create(context.Background(), buyer); err != nil {
		t.Fatalf("create buyer: %v", err)
	}
	if err := users.Create(context.Background(), seller); err != nil {
		t.Fatalf("create seller: %v", err)
	}

	items := newFakeShopItemStore(users)
	purchases := &fakePurchaseStore{}
	publisher := &recordingPurchasePublisher{}
	svc := NewShopService(items, purchases, users, publisher)

	return svc, users, items, publisher
}

func createShopItem(t *testing.T, svc *ShopService, sellerID uint, price int) *domain.ShopItem {
	t.Helper()
	item, err := svc.Create(context.Background(), sellerID, CreateShopItemInput{
		Title:     "Magic Sword",
		PriceGold: price,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	return item
}

func TestShopService_Create(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)

	item, err := svc.Create(context.Background(), 1, CreateShopItemInput{Title: "Shield", PriceGold: 50})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected generated id")
	}
	if !item.IsActive {
		t.Fatal("expected new item to be active")
	}
}

func TestShopService_Create_EmptyTitle(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)

	_, err := svc.Create(context.Background(), 1, CreateShopItemInput{Title: " ", PriceGold: 10})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestShopService_Create_NegativePrice(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)

	_, err := svc.Create(context.Background(), 1, CreateShopItemInput{Title: "Shield", PriceGold: -5})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestShopService_Update_WrongOwner(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)
	item := createShopItem(t, svc, 2, 50)

	newPrice := 60
	_, err := svc.Update(context.Background(), 1, item.ID, UpdateShopItemInput{PriceGold: &newPrice})
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestShopService_Delete(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)
	item := createShopItem(t, svc, 2, 50)

	if err := svc.Delete(context.Background(), 2, item.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	deactivated, err := svc.items.GetByID(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if deactivated.IsActive {
		t.Fatal("expected item to be deactivated")
	}
}

func TestShopService_Buy(t *testing.T) {
	t.Parallel()

	svc, users, _, publisher := newShopTestEnv(t)
	item := createShopItem(t, svc, 2, 50)

	purchase, err := svc.Buy(context.Background(), 1, item.ID)
	if err != nil {
		t.Fatalf("Buy returned error: %v", err)
	}
	if purchase.ItemID != item.ID {
		t.Fatalf("expected item id %d, got %d", item.ID, purchase.ItemID)
	}
	if purchase.Price != 50 {
		t.Fatalf("expected price 50, got %d", purchase.Price)
	}

	buyer, _ := users.GetByID(context.Background(), 1)
	seller, _ := users.GetByID(context.Background(), 2)
	if buyer.Gold != 150 {
		t.Fatalf("expected buyer gold 150, got %d", buyer.Gold)
	}
	if seller.Gold != 60 {
		t.Fatalf("expected seller gold 60, got %d", seller.Gold)
	}
	if publisher.count() != 1 {
		t.Fatalf("expected 1 published event, got %d", publisher.count())
	}
}

func TestShopService_Buy_OwnItem(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)
	item := createShopItem(t, svc, 1, 50)

	if _, err := svc.Buy(context.Background(), 1, item.ID); err != domain.ErrCannotBuyOwnItem {
		t.Fatalf("expected ErrCannotBuyOwnItem, got %v", err)
	}
}

func TestShopService_Buy_InactiveItem(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)
	item := createShopItem(t, svc, 2, 50)
	if err := svc.Delete(context.Background(), 2, item.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if _, err := svc.Buy(context.Background(), 1, item.ID); err != domain.ErrItemNotActive {
		t.Fatalf("expected ErrItemNotActive, got %v", err)
	}
}

func TestShopService_Buy_NotEnoughGold(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)
	item := createShopItem(t, svc, 2, 1000)

	if _, err := svc.Buy(context.Background(), 1, item.ID); err != domain.ErrNotEnoughGold {
		t.Fatalf("expected ErrNotEnoughGold, got %v", err)
	}
}

func TestShopService_Buy_UnknownItem(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newShopTestEnv(t)

	if _, err := svc.Buy(context.Background(), 1, 999); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
