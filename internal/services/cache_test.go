package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
)

// fakeCache is an in-memory cache.Cache recording operations for assertions.
type fakeCache struct {
	mu   sync.Mutex
	data map[string]string
	dels map[string]int
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		data: make(map[string]string),
		dels: make(map[string]int),
	}
}

func (f *fakeCache) Get(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[key]
	if !ok {
		return "", cache.ErrMiss
	}
	return v, nil
}

func (f *fakeCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = value
	return nil
}

func (f *fakeCache) Del(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dels[key]++
	delete(f.data, key)
	return nil
}

func (f *fakeCache) DelPrefix(_ context.Context, prefix string) error {
	return nil
}

func (f *fakeCache) delCount(key string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.dels[key]
}

func TestUserServiceGetByIDCachesAndUpdateInvalidates(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserStore()
	u := &domain.User{Email: "a@b.c", Nickname: "A", XP: 100, Gold: 5}
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	fc := newFakeCache()
	svc := NewUserService(users, fc)

	first, err := svc.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if first.Level != domain.LevelFor(100) {
		t.Fatalf("expected level %d, got %d", domain.LevelFor(100), first.Level)
	}

	users.mu.Lock()
	users.byID[u.ID].XP = 900
	users.mu.Unlock()

	cached, err := svc.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if cached.XP != 100 {
		t.Fatalf("expected cached XP 100, got %d", cached.XP)
	}

	_, err = svc.Update(ctx, u.ID, UpdateProfileInput{Nickname: stringPtr("B")})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if fc.delCount(cache.ProfileKey(u.ID)) == 0 {
		t.Fatal("expected profile cache invalidation")
	}
	if fc.delCount(cache.OverviewKey(u.ID)) == 0 {
		t.Fatal("expected overview cache invalidation")
	}

	fresh, err := svc.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fresh.XP != 900 {
		t.Fatalf("expected fresh XP 900 after invalidation, got %d", fresh.XP)
	}
}

func TestShopServiceListCachesAndCreateInvalidates(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserStore()
	seller := &domain.User{Email: "s@x.y", Nickname: "S", Gold: 1000}
	if err := users.Create(ctx, seller); err != nil {
		t.Fatalf("create seller: %v", err)
	}

	items := newFakeShopItemStore(users)
	if err := items.Create(ctx, &domain.ShopItem{SellerID: seller.ID, Title: "A", PriceGold: 10, IsActive: true}); err != nil {
		t.Fatalf("create item: %v", err)
	}

	fc := newFakeCache()
	svc := NewShopService(items, &fakePurchaseStore{}, users, &recordingPurchasePublisher{}, fc)

	if _, err := svc.List(ctx, 0, false); err != nil {
		t.Fatalf("List: %v", err)
	}

	if err := items.Create(ctx, &domain.ShopItem{SellerID: seller.ID, Title: "B", PriceGold: 20, IsActive: true}); err != nil {
		t.Fatalf("create item: %v", err)
	}

	cached, err := svc.List(ctx, 0, false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cached) != 1 {
		t.Fatalf("expected cached list with 1 item, got %d", len(cached))
	}

	if _, err := svc.Create(ctx, seller.ID, CreateShopItemInput{Title: "C", PriceGold: 5}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	fresh, err := svc.List(ctx, 0, false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(fresh) != 3 {
		t.Fatalf("expected 3 items after invalidation, got %d", len(fresh))
	}
}

func TestShopServiceBuyInvalidatesBuyerAndSeller(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserStore()
	buyer := &domain.User{Email: "b@x.y", Nickname: "B", Gold: 100}
	seller := &domain.User{Email: "s@x.y", Nickname: "S", Gold: 100}
	if err := users.Create(ctx, buyer); err != nil {
		t.Fatalf("create buyer: %v", err)
	}
	if err := users.Create(ctx, seller); err != nil {
		t.Fatalf("create seller: %v", err)
	}

	items := newFakeShopItemStore(users)
	if err := items.Create(ctx, &domain.ShopItem{SellerID: seller.ID, Title: "A", PriceGold: 30, IsActive: true}); err != nil {
		t.Fatalf("create item: %v", err)
	}
	item, err := items.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}

	fc := newFakeCache()
	svc := NewShopService(items, &fakePurchaseStore{}, users, &recordingPurchasePublisher{}, fc)

	if _, err := svc.Buy(ctx, buyer.ID, item.ID); err != nil {
		t.Fatalf("Buy: %v", err)
	}

	if fc.delCount(cache.ProfileKey(buyer.ID)) == 0 {
		t.Fatal("expected buyer profile invalidation")
	}
	if fc.delCount(cache.ProfileKey(seller.ID)) == 0 {
		t.Fatal("expected seller profile invalidation")
	}
	if fc.delCount(cache.OverviewKey(buyer.ID)) == 0 {
		t.Fatal("expected buyer overview invalidation")
	}
}

func TestWorkshopServiceListCachesAndCreateInvalidates(t *testing.T) {
	ctx := context.Background()
	roadmaps := newFakeRoadmapStore()
	if err := roadmaps.Create(ctx, &domain.Roadmap{UserID: 1, Title: "R"}); err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	workshops := newFakeWorkshopStore(roadmaps)
	if err := workshops.Create(ctx, &domain.WorkshopRoadmap{AuthorID: 1, Title: "W", IsPublished: true}); err != nil {
		t.Fatalf("create workshop: %v", err)
	}

	fc := newFakeCache()
	svc := NewWorkshopService(workshops, roadmaps, fc)

	if _, err := svc.List(ctx, 0, false); err != nil {
		t.Fatalf("List: %v", err)
	}

	if err := workshops.Create(ctx, &domain.WorkshopRoadmap{AuthorID: 1, Title: "W2", IsPublished: true}); err != nil {
		t.Fatalf("create workshop: %v", err)
	}

	cached, err := svc.List(ctx, 0, false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cached) != 1 {
		t.Fatalf("expected cached list with 1 item, got %d", len(cached))
	}

	if _, err := svc.Create(ctx, 1, CreateWorkshopInput{RoadmapID: 1, Title: "W3"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	fresh, err := svc.List(ctx, 0, false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(fresh) != 3 {
		t.Fatalf("expected 3 items after invalidation, got %d", len(fresh))
	}
}

func TestQuestServiceCompleteInvalidatesUser(t *testing.T) {
	ctx := context.Background()

	users := newFakeUserStore()
	u := &domain.User{Email: "u@t.dev", Nickname: "tester"}
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	branches := newFakeBranchStore()
	branch := &domain.Branch{UserID: u.ID, Name: "Finance"}
	if err := branches.Create(ctx, branch); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	quests := newFakeQuestStore()
	fc := newFakeCache()
	svc := NewQuestService(quests, branches, users, &recordingEventPublisher{}, fc)

	quest, err := svc.Create(ctx, u.ID, branch.ID, CreateQuestInput{
		Title:      "Read",
		Type:       domain.QuestTypeSimple,
		RewardXP:   100,
		RewardGold: 50,
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}

	if _, err := svc.Complete(ctx, u.ID, quest.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if fc.delCount(cache.ProfileKey(u.ID)) == 0 {
		t.Fatal("expected profile invalidation on quest completion")
	}
	if fc.delCount(cache.OverviewKey(u.ID)) == 0 {
		t.Fatal("expected overview invalidation on quest completion")
	}
}

func stringPtr(s string) *string {
	return &s
}
