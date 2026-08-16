package services

import (
	"context"

	"github.com/north-fy/levelup/internal/domain"
)

// NoopQuestEventPublisher drops completed-quest events.
// Replaced by the outbox-based publisher in the statistics phase.
type NoopQuestEventPublisher struct{}

// PublishQuestCompleted discards the event.
func (NoopQuestEventPublisher) PublishQuestCompleted(context.Context, domain.QuestCompletedEvent) error {
	return nil
}
