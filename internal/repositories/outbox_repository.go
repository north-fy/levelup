package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// OutboxRepository persists deferred events for export to ClickHouse.
type OutboxRepository struct {
	db *gorm.DB
}

// NewOutboxRepository creates the repository.
func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// Insert stores a new pending event.
func (r *OutboxRepository) Insert(ctx context.Context, eventType, payload string) error {
	event := domain.OutboxEvent{
		Type:    eventType,
		Payload: payload,
		Status:  domain.OutboxStatusPending,
	}
	return r.db.WithContext(ctx).Create(&event).Error
}

// Pending returns up to limit events awaiting export, oldest first.
func (r *OutboxRepository) Pending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent
	if err := r.db.WithContext(ctx).
		Where("status = ?", domain.OutboxStatusPending).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// MarkProcessed marks an event as exported.
func (r *OutboxRepository) MarkProcessed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": domain.OutboxStatusProcessed, "processed_at": now}).Error
}
