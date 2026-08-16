package services

import (
	"context"
	"testing"

	"github.com/north-fy/levelup/internal/domain"
)

func newWorkshopTestEnv(t *testing.T) (*WorkshopService, *RoadmapService, *fakeRoadmapStore, *fakeUserStore) {
	t.Helper()

	users := newFakeUserStore()
	if err := users.Create(context.Background(), &domain.User{Email: "author@test.dev"}); err != nil {
		t.Fatalf("create author: %v", err)
	}
	if err := users.Create(context.Background(), &domain.User{Email: "installer@test.dev"}); err != nil {
		t.Fatalf("create installer: %v", err)
	}

	roadmaps := newFakeRoadmapStore()
	roadmapSvc := NewRoadmapService(roadmaps, users, &recordingEventPublisher{})
	workshops := newFakeWorkshopStore(roadmaps)
	workshopSvc := NewWorkshopService(workshops, roadmaps)

	return workshopSvc, roadmapSvc, roadmaps, users
}

func publishWorkshop(t *testing.T, workshopSvc *WorkshopService, roadmapSvc *RoadmapService, authorID, roadmapID uint) *domain.WorkshopRoadmap {
	t.Helper()
	w, err := workshopSvc.Create(context.Background(), authorID, CreateWorkshopInput{RoadmapID: roadmapID})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	return w
}

func TestWorkshopService_Create(t *testing.T) {
	t.Parallel()

	ws, rs, _, _ := newWorkshopTestEnv(t)
	rm, err := rs.Create(context.Background(), 1, CreateRoadmapInput{Title: "Go Path"})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	w := publishWorkshop(t, ws, rs, 1, rm.ID)
	if !w.IsPublished {
		t.Fatal("expected workshop to be published")
	}
	if w.AuthorID != 1 {
		t.Fatalf("expected author 1, got %d", w.AuthorID)
	}
}

func TestWorkshopService_Create_NotOwned(t *testing.T) {
	t.Parallel()

	ws, rs, _, _ := newWorkshopTestEnv(t)
	rm, err := rs.Create(context.Background(), 1, CreateRoadmapInput{Title: "Go Path"})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	_, err = ws.Create(context.Background(), 2, CreateWorkshopInput{RoadmapID: rm.ID})
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestWorkshopService_Install_CopiesGraph(t *testing.T) {
	t.Parallel()

	ws, rs, _, _ := newWorkshopTestEnv(t)
	rm, err := rs.Create(context.Background(), 1, CreateRoadmapInput{Title: "Go Path"})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}

	base, err := rs.AddNode(context.Background(), 1, rm.ID, AddNodeInput{Title: "Basics", Type: domain.QuestTypeSimple, RewardXP: 10})
	if err != nil {
		t.Fatalf("add base node: %v", err)
	}
	if _, err := rs.AddNode(context.Background(), 1, rm.ID, AddNodeInput{
		Title:        "Advanced",
		Type:         domain.QuestTypeSimple,
		RewardXP:     20,
		Dependencies: []uint{base.ID},
	}); err != nil {
		t.Fatalf("add dependent node: %v", err)
	}

	w := publishWorkshop(t, ws, rs, 1, rm.ID)

	detail, err := ws.Install(context.Background(), 2, w.ID)
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if detail.UserID != 2 {
		t.Fatalf("expected installed user 2, got %d", detail.UserID)
	}
	if detail.SourceType != domain.RoadmapSourceWorkshop {
		t.Fatalf("expected workshop source, got %s", detail.SourceType)
	}
	if len(detail.Nodes) != 2 {
		t.Fatalf("expected 2 copied nodes, got %d", len(detail.Nodes))
	}
	if len(detail.Edges) != 1 {
		t.Fatalf("expected 1 copied edge, got %d", len(detail.Edges))
	}

	// The copied graph must not be shared with the source roadmap.
	if detail.Nodes[0].ID == base.ID {
		t.Fatal("copied node must have a fresh id")
	}
	from, to := detail.Edges[0].FromNodeID, detail.Edges[0].ToNodeID
	if from == to {
		t.Fatal("edge endpoints must be distinct after remapping")
	}
	for _, n := range detail.Nodes {
		if n.RoadmapID != detail.ID {
			t.Fatalf("node %d must belong to installed roadmap %d", n.ID, detail.ID)
		}
		if n.Status != domain.QuestStatusTodo {
			t.Fatalf("copied node must be todo, got %s", n.Status)
		}
	}
}

func TestWorkshopService_Install_Unpublished(t *testing.T) {
	t.Parallel()

	ws, rs, _, _ := newWorkshopTestEnv(t)
	rm, err := rs.Create(context.Background(), 1, CreateRoadmapInput{Title: "Go Path"})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	w := publishWorkshop(t, ws, rs, 1, rm.ID)

	unpublished := false
	if _, err := ws.Update(context.Background(), 1, w.ID, UpdateWorkshopInput{IsPublished: &unpublished}); err != nil {
		t.Fatalf("unpublish: %v", err)
	}

	if _, err := ws.Install(context.Background(), 2, w.ID); err != domain.ErrWorkshopNotPublished {
		t.Fatalf("expected ErrWorkshopNotPublished, got %v", err)
	}
}

func TestWorkshopService_Update_WrongAuthor(t *testing.T) {
	t.Parallel()

	ws, rs, _, _ := newWorkshopTestEnv(t)
	rm, err := rs.Create(context.Background(), 1, CreateRoadmapInput{Title: "Go Path"})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	w := publishWorkshop(t, ws, rs, 1, rm.ID)

	title := "Hijacked"
	if _, err := ws.Update(context.Background(), 2, w.ID, UpdateWorkshopInput{Title: &title}); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
