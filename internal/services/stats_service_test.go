package services

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/north-fy/levelup/internal/domain"
)

func TestPeriodStart(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 16, 15, 30, 0, 0, time.UTC)

	if got := PeriodStart("day", now); !got.Equal(time.Date(2026, time.August, 16, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("day: expected start of today, got %v", got)
	}
	if got := PeriodStart("week", now); !got.Equal(now.AddDate(0, 0, -7)) {
		t.Fatalf("week: expected now-7d, got %v", got)
	}
	if got := PeriodStart("month", now); !got.Equal(now.AddDate(0, 0, -30)) {
		t.Fatalf("month: expected now-30d, got %v", got)
	}
	if got := PeriodStart("bogus", now); !got.Equal(PeriodStart("day", now)) {
		t.Fatalf("unknown period must fall back to day, got %v", got)
	}
}

func TestQuestAggregationSQL(t *testing.T) {
	t.Parallel()

	queries := map[string]string{
		"branch":  questBranchAggSQL,
		"roadmap": questRoadmapAggSQL,
		"daily":   questDailyAggSQL,
	}

	for name, query := range queries {
		query := query
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, fragment := range []string{"FROM quest_completed", "GROUP BY", "WHERE user_id = ?"} {
				if !strings.Contains(query, fragment) {
					t.Fatalf("query %q missing fragment %q", query, fragment)
				}
			}
			if !strings.Contains(query, "ORDER BY") {
				t.Fatalf("query %q must be ordered", query)
			}
		})
	}
}

// fakeOutboxStore is an in-memory OutboxStore for tests.
type fakeOutboxStore struct {
	mu     sync.Mutex
	events []domain.OutboxEvent
}

func (f *fakeOutboxStore) Insert(_ context.Context, eventType, payload string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, domain.OutboxEvent{Type: eventType, Payload: payload})
	return nil
}

func (f *fakeOutboxStore) Pending(context.Context, int) ([]domain.OutboxEvent, error) {
	return nil, nil
}

func (f *fakeOutboxStore) MarkProcessed(context.Context, uint) error {
	return nil
}

func (f *fakeOutboxStore) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func TestOutboxQuestPublisher(t *testing.T) {
	t.Parallel()

	store := &fakeOutboxStore{}
	publisher := NewOutboxQuestPublisher(store)

	completedAt := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
	err := publisher.PublishQuestCompleted(context.Background(), domain.QuestCompletedEvent{
		UserID:      1,
		QuestID:     5,
		BranchID:    2,
		XP:          100,
		Gold:        10,
		Hours:       2,
		CompletedAt: completedAt,
	})
	if err != nil {
		t.Fatalf("PublishQuestCompleted: %v", err)
	}
	if store.count() != 1 {
		t.Fatalf("expected 1 event, got %d", store.count())
	}

	event := store.events[0]
	if event.Type != "quest_completed" {
		t.Fatalf("expected type quest_completed, got %q", event.Type)
	}
	if !strings.Contains(event.Payload, `"user_id":1`) {
		t.Fatalf("payload %q missing user_id", event.Payload)
	}
}

func TestOutboxPurchasePublisher(t *testing.T) {
	t.Parallel()

	store := &fakeOutboxStore{}
	publisher := NewOutboxPurchasePublisher(store)

	err := publisher.PublishPurchase(context.Background(), domain.PurchaseEvent{
		UserID: 3,
		ItemID: 7,
		Price:  50,
	})
	if err != nil {
		t.Fatalf("PublishPurchase: %v", err)
	}
	if store.count() != 1 {
		t.Fatalf("expected 1 event, got %d", store.count())
	}
	if store.events[0].Type != "purchase" {
		t.Fatalf("expected type purchase, got %q", store.events[0].Type)
	}
}
