package domain

import (
	"math"
	"time"
)

// User represents a registered account.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;size:255" json:"email"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	GitHubID     string    `gorm:"column:github_id;uniqueIndex;size:64" json:"-"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Status       string    `gorm:"size:255" json:"status"`
	AvatarURL    string    `gorm:"size:512" json:"avatar_url"`
	Level        int       `json:"level"`
	XP           int       `json:"xp"`
	Gold         int       `json:"gold"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// HasGitHub reports whether the user has linked a GitHub account.
func (u *User) HasGitHub() bool {
	return u.GitHubID != ""
}

// LevelFor returns the level corresponding to the given experience points.
func LevelFor(xp int) int {
	if xp <= 0 {
		return 1
	}
	level := int(1 + math.Sqrt(float64(xp)/50))
	if level < 1 {
		return 1
	}
	return level
}
