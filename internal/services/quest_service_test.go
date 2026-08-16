package services

import (
	"context"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/domain"
)

func newQuestTestEnv(t *testing.T) (*QuestService, *fakeUserStore, *fakeQuestStore, *recordingEventPublisher, *domain.Branch) {
	t.Helper()

	users := newFakeUserStore()
	user := &domain.User{Email: "u@test.dev", Nickname: "tester"}
	if err := users.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	branches := newFakeBranchStore()
	branch := &domain.Branch{UserID: user.ID, Name: "Finance"}
	if err := branches.Create(context.Background(), branch); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	quests := newFakeQuestStore()
	publisher := &recordingEventPublisher{}
	svc := NewQuestService(quests, branches, users, publisher)

	return svc, users, quests, publisher, branch
}

func createSimpleQuest(t *testing.T, svc *QuestService, userID, branchID uint) *domain.Quest {
	t.Helper()
	quest, err := svc.Create(context.Background(), userID, branchID, CreateQuestInput{
		Title:      "Read a book",
		Type:       domain.QuestTypeSimple,
		RewardXP:   100,
		RewardGold: 50,
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}
	return quest
}

func createTimedQuest(t *testing.T, svc *QuestService, userID, branchID uint) *domain.Quest {
	t.Helper()
	quest, err := svc.Create(context.Background(), userID, branchID, CreateQuestInput{
		Title:         "Study Go",
		Type:          domain.QuestTypeTimed,
		RewardXP:      100,
		RewardGold:    50,
		DurationHours: 2,
	})
	if err != nil {
		t.Fatalf("create quest: %v", err)
	}
	return quest
}

func TestQuestService_Create_WrongBranchOwner(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)

	_, err := svc.Create(context.Background(), 2, branch.ID, CreateQuestInput{
		Title: "Sneaky",
		Type:  domain.QuestTypeSimple,
	})
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestQuestService_Create_InvalidType(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)

	_, err := svc.Create(context.Background(), 1, branch.ID, CreateQuestInput{
		Title: "Weird",
		Type:  "side_quest",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestQuestService_Complete(t *testing.T) {
	t.Parallel()

	svc, users, _, publisher, branch := newQuestTestEnv(t)
	quest := createSimpleQuest(t, svc, 1, branch.ID)

	done, err := svc.Complete(context.Background(), 1, quest.ID)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if done.Status != domain.QuestStatusDone {
		t.Fatalf("expected status done, got %s", done.Status)
	}
	if done.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}

	user, err := users.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.XP != 100 || user.Gold != 50 {
		t.Fatalf("expected xp=100 gold=50, got xp=%d gold=%d", user.XP, user.Gold)
	}
	if publisher.count() != 1 {
		t.Fatalf("expected 1 published event, got %d", publisher.count())
	}
}

func TestQuestService_Complete_Twice(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)
	quest := createSimpleQuest(t, svc, 1, branch.ID)

	if _, err := svc.Complete(context.Background(), 1, quest.ID); err != nil {
		t.Fatalf("first Complete returned error: %v", err)
	}
	if _, err := svc.Complete(context.Background(), 1, quest.ID); err != domain.ErrQuestAlreadyDone {
		t.Fatalf("expected ErrQuestAlreadyDone, got %v", err)
	}
}

func TestQuestService_Complete_TimedQuest(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)
	quest := createTimedQuest(t, svc, 1, branch.ID)

	if _, err := svc.Complete(context.Background(), 1, quest.ID); err != domain.ErrTimedQuestIncomplete {
		t.Fatalf("expected ErrTimedQuestIncomplete, got %v", err)
	}
}

func TestQuestService_Start_SimpleQuest(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)
	quest := createSimpleQuest(t, svc, 1, branch.ID)

	if _, err := svc.Start(context.Background(), 1, quest.ID); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestQuestService_StartStop(t *testing.T) {
	t.Parallel()

	svc, users, _, publisher, branch := newQuestTestEnv(t)
	quest := createTimedQuest(t, svc, 1, branch.ID)

	started, err := svc.Start(context.Background(), 1, quest.ID)
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if started.Status != domain.QuestStatusInProgress {
		t.Fatalf("expected in_progress, got %s", started.Status)
	}
	if started.StartedAt == nil {
		t.Fatal("expected started_at to be set")
	}

	// Move the clock back so the elapsed time simulates real progress.
	// The fake store keeps the same pointer, so the mutation persists.
	*started.StartedAt = time.Now().Add(-1 * time.Hour)

	stopped, err := svc.Stop(context.Background(), 1, quest.ID)
	if err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if stopped.Status != domain.QuestStatusDone {
		t.Fatalf("expected done, got %s", stopped.Status)
	}

	// 1 of 2 hours elapsed -> 50% of the reward.
	user, err := users.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.XP != 50 || user.Gold != 25 {
		t.Fatalf("expected xp=50 gold=25, got xp=%d gold=%d", user.XP, user.Gold)
	}
	if publisher.count() != 1 {
		t.Fatalf("expected 1 published event, got %d", publisher.count())
	}
}

func TestQuestService_Start_Conflict(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)
	first := createTimedQuest(t, svc, 1, branch.ID)
	second := createTimedQuest(t, svc, 1, branch.ID)

	if _, err := svc.Start(context.Background(), 1, first.ID); err != nil {
		t.Fatalf("first Start returned error: %v", err)
	}
	if _, err := svc.Start(context.Background(), 1, second.ID); err != domain.ErrActiveTimerConflict {
		t.Fatalf("expected ErrActiveTimerConflict, got %v", err)
	}
}

func TestQuestService_Stop_NotInProgress(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)
	quest := createTimedQuest(t, svc, 1, branch.ID)

	if _, err := svc.Stop(context.Background(), 1, quest.ID); err != domain.ErrQuestNotInProgress {
		t.Fatalf("expected ErrQuestNotInProgress, got %v", err)
	}
}

func TestQuestService_Update_DoneQuest(t *testing.T) {
	t.Parallel()

	svc, _, _, _, branch := newQuestTestEnv(t)
	quest := createSimpleQuest(t, svc, 1, branch.ID)

	if _, err := svc.Complete(context.Background(), 1, quest.ID); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	newTitle := "Renamed"
	if _, err := svc.Update(context.Background(), 1, quest.ID, UpdateQuestInput{Title: &newTitle}); err != domain.ErrQuestAlreadyDone {
		t.Fatalf("expected ErrQuestAlreadyDone, got %v", err)
	}
}
