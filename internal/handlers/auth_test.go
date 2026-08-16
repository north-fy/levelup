package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/jwt"
	"github.com/north-fy/levelup/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// handlerUserStore is an in-memory UserStore for handler tests.
type handlerUserStore struct {
	mu      sync.Mutex
	nextID  uint
	byID    map[uint]*domain.User
	byEmail map[string]*domain.User
}

func newHandlerUserStore() *handlerUserStore {
	return &handlerUserStore{
		byID:    make(map[uint]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (f *handlerUserStore) Create(_ context.Context, user *domain.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	user.ID = f.nextID
	f.byID[user.ID] = user
	f.byEmail[user.Email] = user
	return nil
}

func (f *handlerUserStore) GetByID(_ context.Context, id uint) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *handlerUserStore) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (f *handlerUserStore) GetByGitHubID(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

func (f *handlerUserStore) Update(_ context.Context, user *domain.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[user.ID] = user
	f.byEmail[user.Email] = user
	return nil
}

func (f *handlerUserStore) IsEmailTaken(_ context.Context, email string, excludeID uint) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.byEmail[email]
	return ok && u.ID != excludeID, nil
}

// handlerTokenStore is an in-memory TokenStore for handler tests.
type handlerTokenStore struct {
	mu      sync.Mutex
	refresh map[string]uint
}

func newHandlerTokenStore() *handlerTokenStore {
	return &handlerTokenStore{refresh: make(map[string]uint)}
}

func (f *handlerTokenStore) SaveRefresh(_ context.Context, tokenID string, userID uint, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refresh[tokenID] = userID
	return nil
}

func (f *handlerTokenStore) GetRefreshUser(_ context.Context, tokenID string) (uint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	uid, ok := f.refresh[tokenID]
	if !ok {
		return 0, domain.ErrNotFound
	}
	return uid, nil
}

func (f *handlerTokenStore) DeleteRefresh(_ context.Context, tokenID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.refresh, tokenID)
	return nil
}

func (f *handlerTokenStore) BlacklistAccess(context.Context, string, time.Duration) error {
	return nil
}

func (f *handlerTokenStore) IsBlacklisted(context.Context, string) (bool, error) {
	return false, nil
}

func (f *handlerTokenStore) SaveOAuthState(context.Context, string, time.Duration) error {
	return nil
}

func (f *handlerTokenStore) ValidateOAuthState(context.Context, string) error {
	return nil
}

func setupAuthRouter() *gin.Engine {
	users := newHandlerUserStore()
	tokens := newHandlerTokenStore()
	jwtMgr := jwt.New(config.JWT{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    720 * time.Hour,
	})
	authSvc := services.NewAuthService(users, tokens, jwtMgr, config.GitHub{})
	handler := NewAuthHandler(authSvc)

	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	return router
}

func TestRegisterEndpoint(t *testing.T) {
	router := setupAuthRouter()

	body := `{"email":"hero@example.com","password":"password123","nickname":"Hero"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterEndpointInvalidPayload(t *testing.T) {
	router := setupAuthRouter()

	body := `{"email":"not-an-email","password":"123","nickname":"H"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestLoginEndpointWrongPassword(t *testing.T) {
	router := setupAuthRouter()

	registerBody := `{"email":"hero@example.com","password":"password123","nickname":"Hero"}`
	regReq := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(registerBody))
	regReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), regReq)

	loginBody := `{"email":"hero@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
