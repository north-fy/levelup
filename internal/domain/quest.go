package domain

import "time"

// QuestType distinguishes instant and timed quests.
type QuestType string

const (
	// QuestTypeSimple grants a fixed reward when completed.
	QuestTypeSimple QuestType = "simple"
	// QuestTypeTimed tracks hours spent and rewards proportionally.
	QuestTypeTimed QuestType = "timed"
)

// QuestStatus is the lifecycle state of a quest.
type QuestStatus string

const (
	// QuestStatusTodo is the initial state of a quest.
	QuestStatusTodo QuestStatus = "todo"
	// QuestStatusInProgress means a timed quest timer is running.
	QuestStatusInProgress QuestStatus = "in_progress"
	// QuestStatusDone means the quest was completed and rewarded.
	QuestStatusDone QuestStatus = "done"
	// QuestStatusCancelled means the quest was abandoned.
	QuestStatusCancelled QuestStatus = "cancelled"
)

// Quest is a task inside a branch that rewards experience and gold.
type Quest struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	BranchID      uint        `gorm:"index" json:"branch_id"`
	UserID        uint        `gorm:"index" json:"user_id"`
	Title         string      `gorm:"size:255" json:"title"`
	Description   string      `gorm:"size:1024" json:"description"`
	Type          QuestType   `gorm:"size:20" json:"type"`
	RewardXP      int         `json:"reward_xp"`
	RewardGold    int         `json:"reward_gold"`
	DurationHours int         `json:"duration_hours"`
	Status        QuestStatus `gorm:"size:20;index" json:"status"`
	StartedAt     *time.Time  `json:"started_at"`
	CompletedAt   *time.Time  `json:"completed_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// IsTimed reports whether the quest uses the time tracker.
func (q *Quest) IsTimed() bool {
	return q.Type == QuestTypeTimed
}

// QuestCompletedEvent is emitted when a quest is finished.
type QuestCompletedEvent struct {
	UserID      uint
	QuestID     uint
	BranchID    uint
	XP          int
	Gold        int
	Hours       int
	CompletedAt time.Time
}
