package domain

import "time"

// RoadmapSourceType distinguishes personal and installed roadmaps.
type RoadmapSourceType string

const (
	// RoadmapSourceOwn marks a roadmap created by the user.
	RoadmapSourceOwn RoadmapSourceType = "own"
	// RoadmapSourceWorkshop marks a roadmap installed from the workshop.
	RoadmapSourceWorkshop RoadmapSourceType = "workshop"
)

// Roadmap is a graph of roadmap nodes.
type Roadmap struct {
	ID          uint              `gorm:"primaryKey" json:"id"`
	UserID      uint              `gorm:"index" json:"user_id"`
	Title       string            `gorm:"size:255" json:"title"`
	Description string            `gorm:"size:1024" json:"description"`
	SourceType  RoadmapSourceType `gorm:"size:20" json:"source_type"`
	SourceID    uint              `json:"source_id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// RoadmapNode is a single task inside a roadmap graph.
type RoadmapNode struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	RoadmapID     uint        `gorm:"index" json:"roadmap_id"`
	Title         string      `gorm:"size:255" json:"title"`
	Description   string      `gorm:"size:1024" json:"description"`
	Position      int         `json:"position"`
	Type          QuestType   `gorm:"size:20" json:"type"`
	RewardXP      int         `json:"reward_xp"`
	RewardGold    int         `json:"reward_gold"`
	DurationHours int         `json:"duration_hours"`
	Status        QuestStatus `gorm:"size:20" json:"status"`
	CompletedAt   *time.Time  `json:"completed_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// RoadmapEdge is a dependency from one roadmap node to another.
type RoadmapEdge struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	RoadmapID  uint `gorm:"index" json:"roadmap_id"`
	FromNodeID uint `gorm:"index" json:"from_node_id"`
	ToNodeID   uint `gorm:"index" json:"to_node_id"`
}

// WorkshopRoadmap is a published roadmap available for installation.
type WorkshopRoadmap struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AuthorID        uint      `gorm:"index" json:"author_id"`
	SourceRoadmapID uint      `gorm:"index" json:"source_roadmap_id"`
	Title           string    `gorm:"size:255" json:"title"`
	Description     string    `gorm:"size:1024" json:"description"`
	IsPublished     bool      `json:"is_published"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RoadmapDetail is a roadmap with its nodes and edges.
type RoadmapDetail struct {
	Roadmap
	Nodes []RoadmapNode `json:"nodes"`
	Edges []RoadmapEdge `json:"edges"`
}
