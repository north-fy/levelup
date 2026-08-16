package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/services"
)

// handlerBranchStore is an in-memory BranchStore for handler tests.
type handlerBranchStore struct {
	mu       sync.Mutex
	nextID   uint
	byID     map[uint]*domain.Branch
	byUserID map[uint][]*domain.Branch
}

func newHandlerBranchStore() *handlerBranchStore {
	return &handlerBranchStore{
		byID:     make(map[uint]*domain.Branch),
		byUserID: make(map[uint][]*domain.Branch),
	}
}

func (f *handlerBranchStore) Create(_ context.Context, branch *domain.Branch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	branch.ID = f.nextID
	f.byID[branch.ID] = branch
	f.byUserID[branch.UserID] = append(f.byUserID[branch.UserID], branch)
	return nil
}

func (f *handlerBranchStore) GetByIDAndUser(_ context.Context, id, userID uint) (*domain.Branch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.byID[id]
	if !ok || b.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return b, nil
}

func (f *handlerBranchStore) ListByUser(_ context.Context, userID uint) ([]domain.Branch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.Branch
	for _, b := range f.byUserID[userID] {
		result = append(result, *b)
	}
	return result, nil
}

func (f *handlerBranchStore) Update(_ context.Context, branch *domain.Branch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[branch.ID] = branch
	return nil
}

func (f *handlerBranchStore) Delete(_ context.Context, branch *domain.Branch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, branch.ID)
	return nil
}

// handlerQuestStore is an in-memory QuestStore for handler tests.
type handlerQuestStore struct {
	mu         sync.Mutex
	nextID     uint
	byID       map[uint]*domain.Quest
	byBranchID map[uint][]*domain.Quest
}

func newHandlerQuestStore() *handlerQuestStore {
	return &handlerQuestStore{
		byID:       make(map[uint]*domain.Quest),
		byBranchID: make(map[uint][]*domain.Quest),
	}
}

func (f *handlerQuestStore) Create(_ context.Context, quest *domain.Quest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	quest.ID = f.nextID
	f.byID[quest.ID] = quest
	f.byBranchID[quest.BranchID] = append(f.byBranchID[quest.BranchID], quest)
	return nil
}

func (f *handlerQuestStore) GetByIDAndUser(_ context.Context, id, userID uint) (*domain.Quest, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	q, ok := f.byID[id]
	if !ok || q.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return q, nil
}

func (f *handlerQuestStore) ListByBranchAndUser(_ context.Context, branchID, userID uint) ([]domain.Quest, error) {
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

func (f *handlerQuestStore) Update(_ context.Context, quest *domain.Quest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[quest.ID] = quest
	return nil
}

func (f *handlerQuestStore) Delete(_ context.Context, quest *domain.Quest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, quest.ID)
	return nil
}

func (f *handlerQuestStore) HasActiveTimer(_ context.Context, userID uint) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, q := range f.byID {
		if q.UserID == userID && q.Status == domain.QuestStatusInProgress {
			return true, nil
		}
	}
	return false, nil
}

// noopQuestPublisher satisfies QuestEventPublisher in handler tests.
type noopQuestPublisher struct{}

func (noopQuestPublisher) PublishQuestCompleted(context.Context, domain.QuestCompletedEvent) error {
	return nil
}

// setupQuestRouter builds an authenticated router for branches and quests.
func setupQuestRouter() *gin.Engine {
	users := newHandlerUserStore()
	user := &domain.User{Email: "hero@example.com", Nickname: "Hero"}
	if err := users.Create(context.Background(), user); err != nil {
		panic(err)
	}

	branches := newHandlerBranchStore()
	quests := newHandlerQuestStore()
	branchSvc := services.NewBranchService(branches)
	questSvc := services.NewQuestService(quests, branches, users, noopQuestPublisher{})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })

	router.POST("/branches", NewBranchHandler(branchSvc).Create)
	router.GET("/branches", NewBranchHandler(branchSvc).List)
	router.POST("/branches/:branch_id/quests", NewQuestHandler(questSvc).Create)
	router.POST("/quests/:id/complete", NewQuestHandler(questSvc).Complete)
	return router
}

func postJSON(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestBranchCreateEndpoint(t *testing.T) {
	router := setupQuestRouter()

	rec := postJSON(router, "/branches", `{"name":"Finance","color":"#ff0000"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
}

func TestBranchCreateEndpointEmptyName(t *testing.T) {
	router := setupQuestRouter()

	rec := postJSON(router, "/branches", `{"name":"  "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}

func TestQuestCreateEndpoint(t *testing.T) {
	router := setupQuestRouter()

	createBranch := postJSON(router, "/branches", `{"name":"Finance"}`)
	if createBranch.Code != http.StatusCreated {
		t.Fatalf("create branch failed: %d %s", createBranch.Code, createBranch.Body.String())
	}

	rec := postJSON(router, "/branches/1/quests",
		`{"title":"Read a book","type":"simple","reward_xp":100,"reward_gold":50}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
}

func TestQuestCompleteEndpoint(t *testing.T) {
	router := setupQuestRouter()

	postJSON(router, "/branches", `{"name":"Finance"}`)
	postJSON(router, "/branches/1/quests", `{"title":"Read","type":"simple","reward_xp":10}`)

	rec := postJSON(router, "/quests/1/complete", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
}

func TestQuestCompleteEndpointUnknown(t *testing.T) {
	router := setupQuestRouter()

	rec := postJSON(router, "/quests/999/complete", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", rec.Code, rec.Body.String())
	}
}
