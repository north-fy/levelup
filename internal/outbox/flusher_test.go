package outbox

import (
	"strings"
	"testing"

	"github.com/north-fy/levelup/internal/domain"
)

func TestBuildInsertQuestCompleted(t *testing.T) {
	t.Parallel()

	event := domain.OutboxEvent{
		Type:    "quest_completed",
		Payload: `{"user_id":1,"quest_id":5,"branch_id":2,"roadmap_id":0,"xp":100,"gold":10,"hours":2,"completed_at":"2026-08-16T12:00:00Z"}`,
	}

	query, args, err := buildInsert(event)
	if err != nil {
		t.Fatalf("buildInsert: %v", err)
	}
	if !strings.HasPrefix(query, "INSERT INTO quest_completed") {
		t.Fatalf("unexpected query %q", query)
	}
	if len(args) != 8 {
		t.Fatalf("expected 8 args, got %d", len(args))
	}
	if args[0] != uint(1) || args[1] != uint(2) || args[3] != uint(5) {
		t.Fatalf("unexpected args %v", args)
	}
	if args[4] != 100 || args[5] != 10 || args[6] != 2 {
		t.Fatalf("unexpected reward args %v", args)
	}
}

func TestBuildInsertPurchase(t *testing.T) {
	t.Parallel()

	event := domain.OutboxEvent{
		Type:    "purchase",
		Payload: `{"user_id":3,"item_id":7,"price":50,"purchased_at":"2026-08-16T12:00:00Z"}`,
	}

	query, args, err := buildInsert(event)
	if err != nil {
		t.Fatalf("buildInsert: %v", err)
	}
	if !strings.HasPrefix(query, "INSERT INTO purchase") {
		t.Fatalf("unexpected query %q", query)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(args))
	}
	if args[0] != uint(3) || args[1] != uint(7) || args[2] != 50 {
		t.Fatalf("unexpected args %v", args)
	}
}

func TestBuildInsertUnknownType(t *testing.T) {
	t.Parallel()

	event := domain.OutboxEvent{Type: "mystery", Payload: "{}"}
	if _, _, err := buildInsert(event); err == nil {
		t.Fatal("expected error for unknown event type")
	}
}

func TestBuildInsertBadPayload(t *testing.T) {
	t.Parallel()

	event := domain.OutboxEvent{Type: "quest_completed", Payload: "not json"}
	if _, _, err := buildInsert(event); err == nil {
		t.Fatal("expected error for malformed payload")
	}
}
