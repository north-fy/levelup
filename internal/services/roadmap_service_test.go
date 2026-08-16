package services

import (
	"context"
	"testing"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
)

func newRoadmapTestEnv(t *testing.T) (*RoadmapService, *fakeUserStore, *fakeRoadmapStore, *recordingEventPublisher) {
	t.Helper()

	users := newFakeUserStore()
	if err := users.Create(context.Background(), &domain.User{Email: "hero@test.dev", Gold: 100}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	roadmaps := newFakeRoadmapStore()
	publisher := &recordingEventPublisher{}
	svc := NewRoadmapService(roadmaps, users, publisher, cache.Noop{})

	return svc, users, roadmaps, publisher
}

func createRoadmap(t *testing.T, svc *RoadmapService, userID uint) *domain.Roadmap {
	t.Helper()
	rm, err := svc.Create(context.Background(), userID, CreateRoadmapInput{Title: "Go Developer", Description: "Path"})
	if err != nil {
		t.Fatalf("create roadmap: %v", err)
	}
	return rm
}

func addNode(t *testing.T, svc *RoadmapService, userID, roadmapID uint, title string, deps []uint) *domain.RoadmapNode {
	t.Helper()
	node, err := svc.AddNode(context.Background(), userID, roadmapID, AddNodeInput{
		Title:        title,
		Type:         domain.QuestTypeSimple,
		RewardXP:     10,
		RewardGold:   5,
		Dependencies: deps,
	})
	if err != nil {
		t.Fatalf("add node %q: %v", title, err)
	}
	return node
}

func TestRoadmapService_Create(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newRoadmapTestEnv(t)

	rm, err := svc.Create(context.Background(), 1, CreateRoadmapInput{Title: "Finance"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if rm.SourceType != domain.RoadmapSourceOwn {
		t.Fatalf("expected source own, got %s", rm.SourceType)
	}
}

func TestRoadmapService_Create_EmptyTitle(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newRoadmapTestEnv(t)

	_, err := svc.Create(context.Background(), 1, CreateRoadmapInput{Title: " "})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRoadmapService_AddNode_UnknownDependency(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newRoadmapTestEnv(t)
	rm := createRoadmap(t, svc, 1)

	_, err := svc.AddNode(context.Background(), 1, rm.ID, AddNodeInput{
		Title:        "Node",
		Type:         domain.QuestTypeSimple,
		Dependencies: []uint{999},
	})
	if err == nil {
		t.Fatal("expected validation error for unknown dependency")
	}
}

func TestRoadmapService_CompleteNode_RequiresPrerequisites(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newRoadmapTestEnv(t)
	rm := createRoadmap(t, svc, 1)
	base := addNode(t, svc, 1, rm.ID, "Basics", nil)
	dependent := addNode(t, svc, 1, rm.ID, "Advanced", []uint{base.ID})

	if _, err := svc.CompleteNode(context.Background(), 1, rm.ID, dependent.ID); err != domain.ErrPrerequisitesNotMet {
		t.Fatalf("expected ErrPrerequisitesNotMet, got %v", err)
	}

	if _, err := svc.CompleteNode(context.Background(), 1, rm.ID, base.ID); err != nil {
		t.Fatalf("complete base: %v", err)
	}
	done, err := svc.CompleteNode(context.Background(), 1, rm.ID, dependent.ID)
	if err != nil {
		t.Fatalf("complete dependent: %v", err)
	}
	if done.Status != domain.QuestStatusDone {
		t.Fatalf("expected done, got %s", done.Status)
	}
}

func TestRoadmapService_CompleteNode_Rewards(t *testing.T) {
	t.Parallel()

	svc, users, _, publisher := newRoadmapTestEnv(t)
	rm := createRoadmap(t, svc, 1)
	node := addNode(t, svc, 1, rm.ID, "Basics", nil)

	if _, err := svc.CompleteNode(context.Background(), 1, rm.ID, node.ID); err != nil {
		t.Fatalf("complete: %v", err)
	}

	user, err := users.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.XP != 10 || user.Gold != 105 {
		t.Fatalf("expected xp=10 gold=105, got xp=%d gold=%d", user.XP, user.Gold)
	}
	if publisher.count() != 1 {
		t.Fatalf("expected 1 event, got %d", publisher.count())
	}
}

func TestRoadmapService_CompleteNode_Twice(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newRoadmapTestEnv(t)
	rm := createRoadmap(t, svc, 1)
	node := addNode(t, svc, 1, rm.ID, "Basics", nil)

	if _, err := svc.CompleteNode(context.Background(), 1, rm.ID, node.ID); err != nil {
		t.Fatalf("first complete: %v", err)
	}
	if _, err := svc.CompleteNode(context.Background(), 1, rm.ID, node.ID); err != domain.ErrQuestAlreadyDone {
		t.Fatalf("expected ErrQuestAlreadyDone, got %v", err)
	}
}

func TestRoadmapService_UpdateNode_CycleRejected(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newRoadmapTestEnv(t)
	rm := createRoadmap(t, svc, 1)
	a := addNode(t, svc, 1, rm.ID, "A", nil)
	b := addNode(t, svc, 1, rm.ID, "B", []uint{a.ID})

	// Make A depend on B, which would create the cycle A -> B -> A.
	deps := []uint{b.ID}
	if _, err := svc.UpdateNode(context.Background(), 1, rm.ID, a.ID, UpdateNodeInput{Dependencies: &deps}); err != domain.ErrGraphCycle {
		t.Fatalf("expected ErrGraphCycle, got %v", err)
	}
}

func TestRoadmapService_UpdateNode_ChangeDeps(t *testing.T) {
	t.Parallel()

	svc, _, roadmaps, _ := newRoadmapTestEnv(t)
	rm := createRoadmap(t, svc, 1)
	a := addNode(t, svc, 1, rm.ID, "A", nil)
	b := addNode(t, svc, 1, rm.ID, "B", []uint{a.ID})

	deps := []uint{}
	if _, err := svc.UpdateNode(context.Background(), 1, rm.ID, b.ID, UpdateNodeInput{Dependencies: &deps}); err != nil {
		t.Fatalf("UpdateNode returned error: %v", err)
	}

	edges, err := roadmaps.ListEdgesByRoadmap(context.Background(), rm.ID)
	if err != nil {
		t.Fatalf("list edges: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("expected no edges after clearing deps, got %d", len(edges))
	}
}
