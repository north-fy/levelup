package domain

import "time"

// OutboxEvent is a deferred event awaiting export to ClickHouse.
type OutboxEvent struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Type        string     `gorm:"size:32" json:"type"`
	Payload     string     `gorm:"type:jsonb" json:"payload"`
	Status      string     `gorm:"size:16;index" json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

// Event statuses.
const (
	OutboxStatusPending   = "pending"
	OutboxStatusProcessed = "processed"
)
