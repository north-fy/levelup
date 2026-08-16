package services

import (
	"context"
	"time"

	"github.com/north-fy/levelup/internal/domain"
)

// UserStore abstracts user persistence.
type UserStore interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	IsEmailTaken(ctx context.Context, email string, excludeID uint) (bool, error)
}

// BranchStore abstracts branch persistence.
type BranchStore interface {
	Create(ctx context.Context, branch *domain.Branch) error
	GetByIDAndUser(ctx context.Context, id, userID uint) (*domain.Branch, error)
	ListByUser(ctx context.Context, userID uint) ([]domain.Branch, error)
	Update(ctx context.Context, branch *domain.Branch) error
	Delete(ctx context.Context, branch *domain.Branch) error
}

// QuestStore abstracts quest persistence.
type QuestStore interface {
	Create(ctx context.Context, quest *domain.Quest) error
	GetByIDAndUser(ctx context.Context, id, userID uint) (*domain.Quest, error)
	ListByBranchAndUser(ctx context.Context, branchID, userID uint) ([]domain.Quest, error)
	Update(ctx context.Context, quest *domain.Quest) error
	Delete(ctx context.Context, quest *domain.Quest) error
	HasActiveTimer(ctx context.Context, userID uint) (bool, error)
}

// QuestEventPublisher forwards completed-quest events for statistics.
type QuestEventPublisher interface {
	PublishQuestCompleted(ctx context.Context, event domain.QuestCompletedEvent) error
}

// ShopItemStore abstracts shop item persistence and the buy transaction.
type ShopItemStore interface {
	Create(ctx context.Context, item *domain.ShopItem) error
	GetByID(ctx context.Context, id uint) (*domain.ShopItem, error)
	GetByIDAndSeller(ctx context.Context, id, sellerID uint) (*domain.ShopItem, error)
	ListActive(ctx context.Context) ([]domain.ShopItem, error)
	ListByUser(ctx context.Context, userID uint) ([]domain.ShopItem, error)
	Update(ctx context.Context, item *domain.ShopItem) error
	Deactivate(ctx context.Context, item *domain.ShopItem) error
	Buy(ctx context.Context, itemID, buyerID uint) (*domain.Purchase, error)
}

// PurchaseStore abstracts purchase history persistence.
type PurchaseStore interface {
	ListByBuyer(ctx context.Context, buyerID uint) ([]domain.Purchase, error)
}

// PurchaseEventPublisher forwards completed-purchase events for statistics.
type PurchaseEventPublisher interface {
	PublishPurchase(ctx context.Context, event domain.PurchaseEvent) error
}

// RoadmapStore abstracts roadmap, node and edge persistence.
type RoadmapStore interface {
	Create(ctx context.Context, roadmap *domain.Roadmap) error
	GetByID(ctx context.Context, id uint) (*domain.Roadmap, error)
	GetByIDAndUser(ctx context.Context, id, userID uint) (*domain.Roadmap, error)
	ListByUser(ctx context.Context, userID uint) ([]domain.Roadmap, error)
	Update(ctx context.Context, roadmap *domain.Roadmap) error
	Delete(ctx context.Context, roadmap *domain.Roadmap) error
	AddNode(ctx context.Context, roadmapID uint, node *domain.RoadmapNode, deps []uint) error
	UpdateNode(ctx context.Context, node *domain.RoadmapNode) error
	UpdateNodeDeps(ctx context.Context, roadmapID, nodeID uint, deps []uint) error
	GetNodeByIDAndUser(ctx context.Context, nodeID, userID uint) (*domain.RoadmapNode, error)
	ListNodesByRoadmap(ctx context.Context, roadmapID uint) ([]domain.RoadmapNode, error)
	ListEdgesByRoadmap(ctx context.Context, roadmapID uint) ([]domain.RoadmapEdge, error)
	MarkNodeDone(ctx context.Context, nodeID uint) error
}

// WorkshopStore abstracts published roadmap persistence.
type WorkshopStore interface {
	Create(ctx context.Context, workshop *domain.WorkshopRoadmap) error
	GetByID(ctx context.Context, id uint) (*domain.WorkshopRoadmap, error)
	GetByIDAndAuthor(ctx context.Context, id, authorID uint) (*domain.WorkshopRoadmap, error)
	ListPublished(ctx context.Context) ([]domain.WorkshopRoadmap, error)
	ListByAuthor(ctx context.Context, authorID uint) ([]domain.WorkshopRoadmap, error)
	Update(ctx context.Context, workshop *domain.WorkshopRoadmap) error
	InstallCopy(ctx context.Context, installerID uint, workshop *domain.WorkshopRoadmap) (*domain.Roadmap, error)
}

// OutboxStore abstracts deferred event persistence.
type OutboxStore interface {
	Insert(ctx context.Context, eventType, payload string) error
	Pending(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uint) error
}

// TokenStore abstracts ephemeral auth data persistence.
type TokenStore interface {
	SaveRefresh(ctx context.Context, tokenID string, userID uint, ttl time.Duration) error
	GetRefreshUser(ctx context.Context, tokenID string) (uint, error)
	DeleteRefresh(ctx context.Context, tokenID string) error
	BlacklistAccess(ctx context.Context, tokenID string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	SaveOAuthState(ctx context.Context, state string, ttl time.Duration) error
	ValidateOAuthState(ctx context.Context, state string) error
}
