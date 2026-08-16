package services

import (
	"context"
	"sync"
	"time"

	"github.com/north-fy/levelup/internal/domain"
)

// fakeUserStore is an in-memory implementation of UserStore for tests.
type fakeUserStore struct {
	mu      sync.Mutex
	nextID  uint
	byID    map[uint]*domain.User
	byEmail map[string]*domain.User
	byGH    map[string]*domain.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		byID:    make(map[uint]*domain.User),
		byEmail: make(map[string]*domain.User),
		byGH:    make(map[string]*domain.User),
	}
}

func (f *fakeUserStore) Create(_ context.Context, user *domain.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	user.ID = f.nextID
	f.byID[user.ID] = user
	f.byEmail[user.Email] = user
	if user.GitHubID != "" {
		f.byGH[user.GitHubID] = user
	}
	return nil
}

func (f *fakeUserStore) GetByID(_ context.Context, id uint) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserStore) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserStore) GetByGitHubID(_ context.Context, githubID string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byGH[githubID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserStore) Update(_ context.Context, user *domain.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[user.ID] = user
	f.byEmail[user.Email] = user
	if user.GitHubID != "" {
		f.byGH[user.GitHubID] = user
	}
	return nil
}

func (f *fakeUserStore) IsEmailTaken(_ context.Context, email string, excludeID uint) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byEmail[email]
	return ok && u.ID != excludeID, nil
}

// fakeTokenStore is an in-memory implementation of TokenStore for tests.
type fakeTokenStore struct {
	mu        sync.Mutex
	refresh   map[string]uint
	states    map[string]struct{}
	blacklist map[string]struct{}
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{
		refresh:   make(map[string]uint),
		states:    make(map[string]struct{}),
		blacklist: make(map[string]struct{}),
	}
}

func (f *fakeTokenStore) SaveRefresh(_ context.Context, tokenID string, userID uint, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refresh[tokenID] = userID
	return nil
}

func (f *fakeTokenStore) GetRefreshUser(_ context.Context, tokenID string) (uint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	uid, ok := f.refresh[tokenID]
	if !ok {
		return 0, domain.ErrNotFound
	}
	return uid, nil
}

func (f *fakeTokenStore) DeleteRefresh(_ context.Context, tokenID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.refresh, tokenID)
	return nil
}

func (f *fakeTokenStore) BlacklistAccess(_ context.Context, tokenID string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blacklist[tokenID] = struct{}{}
	return nil
}

func (f *fakeTokenStore) IsBlacklisted(_ context.Context, tokenID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.blacklist[tokenID]
	return ok, nil
}

func (f *fakeTokenStore) SaveOAuthState(_ context.Context, state string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states[state] = struct{}{}
	return nil
}

func (f *fakeTokenStore) ValidateOAuthState(_ context.Context, state string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.states[state]; !ok {
		return domain.ErrInvalidState
	}
	delete(f.states, state)
	return nil
}

// fakeBranchStore is an in-memory implementation of BranchStore for tests.
type fakeBranchStore struct {
	mu       sync.Mutex
	nextID   uint
	byID     map[uint]*domain.Branch
	byUserID map[uint][]*domain.Branch
}

func newFakeBranchStore() *fakeBranchStore {
	return &fakeBranchStore{
		byID:     make(map[uint]*domain.Branch),
		byUserID: make(map[uint][]*domain.Branch),
	}
}

func (f *fakeBranchStore) Create(_ context.Context, branch *domain.Branch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	branch.ID = f.nextID
	f.byID[branch.ID] = branch
	f.byUserID[branch.UserID] = append(f.byUserID[branch.UserID], branch)
	return nil
}

func (f *fakeBranchStore) GetByIDAndUser(_ context.Context, id, userID uint) (*domain.Branch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.byID[id]
	if !ok || b.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return b, nil
}

func (f *fakeBranchStore) ListByUser(_ context.Context, userID uint) ([]domain.Branch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := f.byUserID[userID]
	result := make([]domain.Branch, 0, len(items))
	for _, b := range items {
		result = append(result, *b)
	}
	return result, nil
}

func (f *fakeBranchStore) Update(_ context.Context, branch *domain.Branch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[branch.ID] = branch
	return nil
}

func (f *fakeBranchStore) Delete(_ context.Context, branch *domain.Branch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, branch.ID)
	items := f.byUserID[branch.UserID]
	filtered := items[:0]
	for _, b := range items {
		if b.ID != branch.ID {
			filtered = append(filtered, b)
		}
	}
	f.byUserID[branch.UserID] = filtered
	return nil
}

// fakeQuestStore is an in-memory implementation of QuestStore for tests.
type fakeQuestStore struct {
	mu         sync.Mutex
	nextID     uint
	byID       map[uint]*domain.Quest
	byBranchID map[uint][]*domain.Quest
}

func newFakeQuestStore() *fakeQuestStore {
	return &fakeQuestStore{
		byID:       make(map[uint]*domain.Quest),
		byBranchID: make(map[uint][]*domain.Quest),
	}
}

func (f *fakeQuestStore) Create(_ context.Context, quest *domain.Quest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	quest.ID = f.nextID
	f.byID[quest.ID] = quest
	f.byBranchID[quest.BranchID] = append(f.byBranchID[quest.BranchID], quest)
	return nil
}

func (f *fakeQuestStore) GetByIDAndUser(_ context.Context, id, userID uint) (*domain.Quest, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	q, ok := f.byID[id]
	if !ok || q.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return q, nil
}

func (f *fakeQuestStore) ListByBranchAndUser(_ context.Context, branchID, userID uint) ([]domain.Quest, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.Quest
	for _, q := range f.byBranchID[branchID] {
		if q.UserID == userID {
			result = append(result, *q)
		}
	}
	return result, nil
}

func (f *fakeQuestStore) Update(_ context.Context, quest *domain.Quest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[quest.ID] = quest
	return nil
}

func (f *fakeQuestStore) Delete(_ context.Context, quest *domain.Quest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, quest.ID)
	items := f.byBranchID[quest.BranchID]
	filtered := items[:0]
	for _, q := range items {
		if q.ID != quest.ID {
			filtered = append(filtered, q)
		}
	}
	f.byBranchID[quest.BranchID] = filtered
	return nil
}

func (f *fakeQuestStore) HasActiveTimer(_ context.Context, userID uint) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, q := range f.byID {
		if q.UserID == userID && q.Status == domain.QuestStatusInProgress {
			return true, nil
		}
	}
	return false, nil
}

// recordingEventPublisher records published events for assertions.
type recordingEventPublisher struct {
	mu     sync.Mutex
	events []domain.QuestCompletedEvent
}

func (r *recordingEventPublisher) PublishQuestCompleted(_ context.Context, event domain.QuestCompletedEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}

func (r *recordingEventPublisher) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

// fakeShopItemStore is an in-memory implementation of ShopItemStore for tests.
type fakeShopItemStore struct {
	mu         sync.Mutex
	users      *fakeUserStore
	nextID     uint
	nextPurID  uint
	byID       map[uint]*domain.ShopItem
	bySellerID map[uint][]*domain.ShopItem
	purchases  []domain.Purchase
}

func newFakeShopItemStore(users *fakeUserStore) *fakeShopItemStore {
	return &fakeShopItemStore{
		users:      users,
		byID:       make(map[uint]*domain.ShopItem),
		bySellerID: make(map[uint][]*domain.ShopItem),
	}
}

func (f *fakeShopItemStore) Create(_ context.Context, item *domain.ShopItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	item.ID = f.nextID
	f.byID[item.ID] = item
	f.bySellerID[item.SellerID] = append(f.bySellerID[item.SellerID], item)
	return nil
}

func (f *fakeShopItemStore) GetByID(_ context.Context, id uint) (*domain.ShopItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeShopItemStore) GetByIDAndSeller(_ context.Context, id, sellerID uint) (*domain.ShopItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.byID[id]
	if !ok || item.SellerID != sellerID {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeShopItemStore) ListActive(_ context.Context) ([]domain.ShopItem, error) {
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

func (f *fakeShopItemStore) ListByUser(_ context.Context, userID uint) ([]domain.ShopItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.ShopItem
	for _, item := range f.bySellerID[userID] {
		result = append(result, *item)
	}
	return result, nil
}

func (f *fakeShopItemStore) Update(_ context.Context, item *domain.ShopItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[item.ID] = item
	return nil
}

func (f *fakeShopItemStore) Deactivate(_ context.Context, item *domain.ShopItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item.IsActive = false
	f.byID[item.ID] = item
	return nil
}

func (f *fakeShopItemStore) Buy(_ context.Context, itemID, buyerID uint) (*domain.Purchase, error) {
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
	seller, err := f.users.GetByID(context.Background(), item.SellerID)
	if err != nil {
		return nil, err
	}

	buyer.Gold -= item.PriceGold
	seller.Gold += item.PriceGold
	if err := f.users.Update(context.Background(), buyer); err != nil {
		return nil, err
	}
	if err := f.users.Update(context.Background(), seller); err != nil {
		return nil, err
	}

	f.nextPurID++
	purchase := &domain.Purchase{
		ID:        f.nextPurID,
		ItemID:    item.ID,
		BuyerID:   buyerID,
		SellerID:  item.SellerID,
		Price:     item.PriceGold,
		CreatedAt: time.Now(),
	}
	f.purchases = append(f.purchases, *purchase)
	return purchase, nil
}

// fakePurchaseStore is an in-memory implementation of PurchaseStore for tests.
type fakePurchaseStore struct {
	mu        sync.Mutex
	purchases []domain.Purchase
}

func (f *fakePurchaseStore) ListByBuyer(_ context.Context, buyerID uint) ([]domain.Purchase, error) {
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

// recordingPurchasePublisher records published purchase events for assertions.
type recordingPurchasePublisher struct {
	mu     sync.Mutex
	events []domain.PurchaseEvent
}

func (r *recordingPurchasePublisher) PublishPurchase(_ context.Context, event domain.PurchaseEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}

func (r *recordingPurchasePublisher) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}
