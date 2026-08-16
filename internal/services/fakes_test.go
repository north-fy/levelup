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
