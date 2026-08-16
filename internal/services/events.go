package services

import (
	"context"
	"encoding/json"

	"github.com/north-fy/levelup/internal/domain"
)

// NoopQuestEventPublisher drops completed-quest events.
// Replaced by the outbox-based publisher in production wiring.
type NoopQuestEventPublisher struct{}

// PublishQuestCompleted discards the event.
func (NoopQuestEventPublisher) PublishQuestCompleted(context.Context, domain.QuestCompletedEvent) error {
	return nil
}

// NoopPurchaseEventPublisher drops completed-purchase events.
// Replaced by the outbox-based publisher in production wiring.
type NoopPurchaseEventPublisher struct{}

// PublishPurchase discards the event.
func (NoopPurchaseEventPublisher) PublishPurchase(context.Context, domain.PurchaseEvent) error {
	return nil
}

// OutboxQuestPublisher stores completed-quest events in the outbox.
type OutboxQuestPublisher struct {
	outbox OutboxStore
}

// NewOutboxQuestPublisher creates the quest outbox publisher.
func NewOutboxQuestPublisher(outbox OutboxStore) *OutboxQuestPublisher {
	return &OutboxQuestPublisher{outbox: outbox}
}

// PublishQuestCompleted marshals and enqueues the event.
func (p *OutboxQuestPublisher) PublishQuestCompleted(ctx context.Context, event domain.QuestCompletedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.outbox.Insert(ctx, "quest_completed", string(payload))
}

// OutboxPurchasePublisher stores completed-purchase events in the outbox.
type OutboxPurchasePublisher struct {
	outbox OutboxStore
}

// NewOutboxPurchasePublisher creates the purchase outbox publisher.
func NewOutboxPurchasePublisher(outbox OutboxStore) *OutboxPurchasePublisher {
	return &OutboxPurchasePublisher{outbox: outbox}
}

// PublishPurchase marshals and enqueues the event.
func (p *OutboxPurchasePublisher) PublishPurchase(ctx context.Context, event domain.PurchaseEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.outbox.Insert(ctx, "purchase", string(payload))
}
