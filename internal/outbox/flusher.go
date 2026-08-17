package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/north-fy/levelup/internal/domain"
)

// EventStore abstracts the pending events needed by the flusher.
type EventStore interface {
	Pending(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uint) error
}

// CHExecer abstracts ClickHouse writes for the flusher.
type CHExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Flusher drains outbox events into ClickHouse on a schedule.
type Flusher struct {
	store     EventStore
	ch        CHExecer
	log       *zap.Logger
	interval  time.Duration
	batchSize int
}

// NewFlusher creates the outbox flusher.
func NewFlusher(store EventStore, ch CHExecer, log *zap.Logger) *Flusher {
	return &Flusher{
		store:     store,
		ch:        ch,
		log:       log,
		interval:  5 * time.Second,
		batchSize: 100,
	}
}

// Run flushes pending events until the context is cancelled.
func (f *Flusher) Run(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := f.FlushOnce(ctx); err != nil {
				f.log.Error("outbox flush failed", zap.Error(err))
			}
		}
	}
}

// FlushOnce exports one batch of pending events to ClickHouse.
func (f *Flusher) FlushOnce(ctx context.Context) error {
	events, err := f.store.Pending(ctx, f.batchSize)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := f.flushEvent(ctx, event); err != nil {
			return err
		}
		if err := f.store.MarkProcessed(ctx, event.ID); err != nil {
			return err
		}
	}
	return nil
}

func (f *Flusher) flushEvent(ctx context.Context, event domain.OutboxEvent) error {
	query, args, err := buildInsert(event)
	if err != nil {
		return err
	}
	_, err = f.ch.ExecContext(ctx, query, args...)
	return err
}

// buildInsert converts an outbox event into a ClickHouse INSERT statement.
func buildInsert(event domain.OutboxEvent) (string, []any, error) {
	switch event.Type {
	case "quest_completed":
		var e domain.QuestCompletedEvent
		if err := json.Unmarshal([]byte(event.Payload), &e); err != nil {
			return "", nil, err
		}
		return `INSERT INTO quest_completed (user_id, branch_id, roadmap_id, quest_id, xp, gold, hours, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			[]any{e.UserID, e.BranchID, e.RoadmapID, e.QuestID, e.XP, e.Gold, e.Hours, e.CompletedAt}, nil
	case "purchase":
		var e domain.PurchaseEvent
		if err := json.Unmarshal([]byte(event.Payload), &e); err != nil {
			return "", nil, err
		}
		return `INSERT INTO purchase (user_id, item_id, price, purchased_at) VALUES (?, ?, ?, ?)`,
			[]any{e.UserID, e.ItemID, e.Price, e.PurchasedAt}, nil
	default:
		return "", nil, fmt.Errorf("unknown outbox event type %q", event.Type)
	}
}
