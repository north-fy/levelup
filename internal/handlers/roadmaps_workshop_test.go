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

// handlerRoadmapStore is an in-memory RoadmapStore for handler tests.
type handlerRoadmapStore struct {
	mu             sync.Mutex
	nextID         uint
	nextNodeID     uint
	byID           map[uint]*domain.Roadmap
	nodesByRoadmap map[uint][]*domain.RoadmapNode
	edgesByRoadmap map[uint][]*domain.RoadmapEdge
	nodesByID      map[uint]*domain.RoadmapNode
}

func newHandlerRoadmapStore() *handlerRoadmapStore {
	return &handlerRoadmapStore{
		byID:           make(map[uint]*domain.Roadmap),
		nodesByRoadmap: make(map[uint][]*domain.RoadmapNode),
		edgesByRoadmap: make(map[uint][]*domain.RoadmapEdge),
		nodesByID:      make(map[uint]*domain.RoadmapNode),
	}
}

func (f *handlerRoadmapStore) Create(_ context.Context, roadmap *domain.Roadmap) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	roadmap.ID = f.nextID
	f.byID[roadmap.ID] = roadmap
	return nil
}

func (f *handlerRoadmapStore) GetByID(context.Context, uint) (*domain.Roadmap, error) {
	return nil, domain.ErrNotFound
}

func (f *handlerRoadmapStore) GetByIDAndUser(_ context.Context, id, userID uint) (*domain.Roadmap, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rm, ok := f.byID[id]
	if !ok || rm.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return rm, nil
}

func (f *handlerRoadmapStore) ListByUser(context.Context, uint) ([]domain.Roadmap, error) {
	return nil, nil
}

func (f *handlerRoadmapStore) Update(_ context.Context, roadmap *domain.Roadmap) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[roadmap.ID] = roadmap
	return nil
}

func (f *handlerRoadmapStore) Delete(_ context.Context, roadmap *domain.Roadmap) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, roadmap.ID)
	return nil
}

func (f *handlerRoadmapStore) AddNode(_ context.Context, roadmapID uint, node *domain.RoadmapNode, deps []uint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextNodeID++
	node.ID = f.nextNodeID
	node.RoadmapID = roadmapID
	f.nodesByRoadmap[roadmapID] = append(f.nodesByRoadmap[roadmapID], node)
	f.nodesByID[node.ID] = node
	for _, dep := range deps {
		f.edgesByRoadmap[roadmapID] = append(f.edgesByRoadmap[roadmapID], &domain.RoadmapEdge{
			ID:         uint(len(f.edgesByRoadmap[roadmapID]) + 1),
			RoadmapID:  roadmapID,
			FromNodeID: dep,
			ToNodeID:   node.ID,
		})
	}
	return nil
}

func (f *handlerRoadmapStore) UpdateNode(_ context.Context, node *domain.RoadmapNode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodesByID[node.ID] = node
	return nil
}

func (f *handlerRoadmapStore) UpdateNodeDeps(context.Context, uint, uint, []uint) error {
	return nil
}

func (f *handlerRoadmapStore) GetNodeByIDAndUser(_ context.Context, nodeID, userID uint) (*domain.RoadmapNode, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	node, ok := f.nodesByID[nodeID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	rm, ok := f.byID[node.RoadmapID]
	if !ok || rm.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return node, nil
}

func (f *handlerRoadmapStore) ListNodesByRoadmap(_ context.Context, roadmapID uint) ([]domain.RoadmapNode, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.RoadmapNode
	for _, n := range f.nodesByRoadmap[roadmapID] {
		result = append(result, *n)
	}
	return result, nil
}

func (f *handlerRoadmapStore) ListEdgesByRoadmap(_ context.Context, roadmapID uint) ([]domain.RoadmapEdge, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var result []domain.RoadmapEdge
	for _, e := range f.edgesByRoadmap[roadmapID] {
		result = append(result, *e)
	}
	return result, nil
}

func (f *handlerRoadmapStore) MarkNodeDone(_ context.Context, nodeID uint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	node, ok := f.nodesByID[nodeID]
	if !ok {
		return domain.ErrNotFound
	}
	node.Status = domain.QuestStatusDone
	return nil
}

// handlerWorkshopStore is an in-memory WorkshopStore for handler tests.
type handlerWorkshopStore struct {
	mu       sync.Mutex
	roadmaps *handlerRoadmapStore
	nextID   uint
	byID     map[uint]*domain.WorkshopRoadmap
}

func newHandlerWorkshopStore(roadmaps *handlerRoadmapStore) *handlerWorkshopStore {
	return &handlerWorkshopStore{
		roadmaps: roadmaps,
		byID:     make(map[uint]*domain.WorkshopRoadmap),
	}
}

func (f *handlerWorkshopStore) Create(_ context.Context, workshop *domain.WorkshopRoadmap) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	workshop.ID = f.nextID
	f.byID[workshop.ID] = workshop
	return nil
}

func (f *handlerWorkshopStore) GetByID(_ context.Context, id uint) (*domain.WorkshopRoadmap, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return w, nil
}

func (f *handlerWorkshopStore) GetByIDAndAuthor(_ context.Context, id, authorID uint) (*domain.WorkshopRoadmap, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w, ok := f.byID[id]
	if !ok || w.AuthorID != authorID {
		return nil, domain.ErrNotFound
	}
	return w, nil
}

func (f *handlerWorkshopStore) ListPublished(context.Context) ([]domain.WorkshopRoadmap, error) {
	return nil, nil
}

func (f *handlerWorkshopStore) ListByAuthor(context.Context, uint) ([]domain.WorkshopRoadmap, error) {
	return nil, nil
}

func (f *handlerWorkshopStore) Update(_ context.Context, workshop *domain.WorkshopRoadmap) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[workshop.ID] = workshop
	return nil
}

func (f *handlerWorkshopStore) InstallCopy(ctx context.Context, installerID uint, workshop *domain.WorkshopRoadmap) (*domain.Roadmap, error) {
	installed := &domain.Roadmap{
		UserID:      installerID,
		Title:       workshop.Title,
		Description: workshop.Description,
		SourceType:  domain.RoadmapSourceWorkshop,
		SourceID:    workshop.ID,
	}
	if err := f.roadmaps.Create(ctx, installed); err != nil {
		return nil, err
	}
	return installed, nil
}

// handlerQuestPublisher satisfies QuestEventPublisher in handler tests.
type handlerQuestPublisher struct{}

func (handlerQuestPublisher) PublishQuestCompleted(context.Context, domain.QuestCompletedEvent) error {
	return nil
}

// setupRoadmapRouter builds an authenticated router for roadmap and workshop endpoints.
func setupRoadmapRouter() *gin.Engine {
	users := newHandlerUserStore()
	user := &domain.User{Email: "hero@example.com", Nickname: "Hero", Gold: 100}
	if err := users.Create(context.Background(), user); err != nil {
		panic(err)
	}

	roadmaps := newHandlerRoadmapStore()
	workshops := newHandlerWorkshopStore(roadmaps)
	roadmapSvc := services.NewRoadmapService(roadmaps, users, handlerQuestPublisher{}, cache.Noop{})
	workshopSvc := services.NewWorkshopService(workshops, roadmaps, cache.Noop{})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", user.ID); c.Next() })

	rh := NewRoadmapHandler(roadmapSvc)
	wh := NewWorkshopHandler(workshopSvc)

	router.POST("/roadmaps", rh.Create)
	router.POST("/roadmaps/:id/nodes", rh.AddNode)
	router.POST("/roadmaps/:id/nodes/:nodeId/complete", rh.CompleteNode)
	router.POST("/workshop/roadmaps", wh.Create)
	router.POST("/workshop/roadmaps/:id/install", wh.Install)
	return router
}

func TestRoadmapCreateEndpoint(t *testing.T) {
	router := setupRoadmapRouter()

	rec := postJSON(router, "/roadmaps", `{"title":"Go Developer"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}
}

func TestRoadmapCompleteEndpoint(t *testing.T) {
	router := setupRoadmapRouter()

	postJSON(router, "/roadmaps", `{"title":"Go Developer"}`)
	postJSON(router, "/roadmaps/1/nodes", `{"title":"Basics","type":"simple","reward_xp":10}`)

	rec := postJSON(router, "/roadmaps/1/nodes/1/complete", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
}

func TestRoadmapCompleteBlockedByPrerequisite(t *testing.T) {
	router := setupRoadmapRouter()

	postJSON(router, "/roadmaps", `{"title":"Go Developer"}`)
	postJSON(router, "/roadmaps/1/nodes", `{"title":"Basics","type":"simple"}`)
	postJSON(router, "/roadmaps/1/nodes", `{"title":"Advanced","type":"simple","dependencies":[1]}`)

	rec := postJSON(router, "/roadmaps/1/nodes/2/complete", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rec.Code, rec.Body.String())
	}
}

func TestWorkshopInstallEndpoint(t *testing.T) {
	router := setupRoadmapRouter()

	postJSON(router, "/roadmaps", `{"title":"Go Developer"}`)
	postJSON(router, "/roadmaps/1/nodes", `{"title":"Basics","type":"simple"}`)

	publish := postJSON(router, "/workshop/roadmaps", `{"roadmap_id":1}`)
	if publish.Code != http.StatusCreated {
		t.Fatalf("publish failed: %d %s", publish.Code, publish.Body.String())
	}

	rec := postJSON(router, "/workshop/roadmaps/1/install", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
}
