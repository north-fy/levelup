package domain

import "time"

// Branch groups quests into tracks such as finance or sports.
type Branch struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	Name        string    `gorm:"size:100" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Color       string    `gorm:"size:20" json:"color"`
	Icon        string    `gorm:"size:50" json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
