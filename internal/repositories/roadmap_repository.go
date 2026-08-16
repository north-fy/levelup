package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// RoadmapRepository persists roadmaps, their nodes and edges.
type RoadmapRepository struct {
	db *gorm.DB
}

// NewRoadmapRepository creates the repository.
func NewRoadmapRepository(db *gorm.DB) *RoadmapRepository {
	return &RoadmapRepository{db: db}
}

// Create stores a new roadmap.
func (r *RoadmapRepository) Create(ctx context.Context, roadmap *domain.Roadmap) error {
	return r.db.WithContext(ctx).Create(roadmap).Error
}

// GetByID returns a roadmap by id.
func (r *RoadmapRepository) GetByID(ctx context.Context, id uint) (*domain.Roadmap, error) {
	var roadmap domain.Roadmap
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&roadmap).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &roadmap, nil
}

// GetByIDAndUser returns a roadmap owned by the given user.
func (r *RoadmapRepository) GetByIDAndUser(ctx context.Context, id, userID uint) (*domain.Roadmap, error) {
	var roadmap domain.Roadmap
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&roadmap).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &roadmap, nil
}

// ListByUser returns all roadmaps of a user.
func (r *RoadmapRepository) ListByUser(ctx context.Context, userID uint) ([]domain.Roadmap, error) {
	var roadmaps []domain.Roadmap
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&roadmaps).Error; err != nil {
		return nil, err
	}
	return roadmaps, nil
}

// Update persists roadmap changes.
func (r *RoadmapRepository) Update(ctx context.Context, roadmap *domain.Roadmap) error {
	return r.db.WithContext(ctx).Save(roadmap).Error
}

// Delete removes a roadmap and cascades to its nodes and edges.
func (r *RoadmapRepository) Delete(ctx context.Context, roadmap *domain.Roadmap) error {
	return r.db.WithContext(ctx).Delete(roadmap).Error
}

// AddNode creates a node and its dependency edges atomically.
func (r *RoadmapRepository) AddNode(ctx context.Context, roadmapID uint, node *domain.RoadmapNode, deps []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(node).Error; err != nil {
			return err
		}
		return r.insertEdges(tx, roadmapID, node.ID, deps)
	})
}

// UpdateNodeDeps replaces a node's dependency edges atomically.
func (r *RoadmapRepository) UpdateNodeDeps(ctx context.Context, roadmapID, nodeID uint, deps []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("roadmap_id = ? AND to_node_id = ?", roadmapID, nodeID).Delete(&domain.RoadmapEdge{}).Error; err != nil {
			return err
		}
		return r.insertEdges(tx, roadmapID, nodeID, deps)
	})
}

// UpdateNode persists node changes.
func (r *RoadmapRepository) UpdateNode(ctx context.Context, node *domain.RoadmapNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

// GetNodeByIDAndUser returns a node whose roadmap belongs to the user.
func (r *RoadmapRepository) GetNodeByIDAndUser(ctx context.Context, nodeID, userID uint) (*domain.RoadmapNode, error) {
	var node domain.RoadmapNode
	err := r.db.WithContext(ctx).
		Joins("JOIN roadmaps ON roadmaps.id = roadmap_nodes.roadmap_id").
		Where("roadmap_nodes.id = ? AND roadmaps.user_id = ?", nodeID, userID).
		First(&node).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &node, nil
}

// ListNodesByRoadmap returns all nodes of a roadmap in position order.
func (r *RoadmapRepository) ListNodesByRoadmap(ctx context.Context, roadmapID uint) ([]domain.RoadmapNode, error) {
	var nodes []domain.RoadmapNode
	if err := r.db.WithContext(ctx).Where("roadmap_id = ?", roadmapID).Order("position ASC, id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListEdgesByRoadmap returns all edges of a roadmap.
func (r *RoadmapRepository) ListEdgesByRoadmap(ctx context.Context, roadmapID uint) ([]domain.RoadmapEdge, error) {
	var edges []domain.RoadmapEdge
	if err := r.db.WithContext(ctx).Where("roadmap_id = ?", roadmapID).Find(&edges).Error; err != nil {
		return nil, err
	}
	return edges, nil
}

// MarkNodeDone marks a node as completed.
func (r *RoadmapRepository) MarkNodeDone(ctx context.Context, nodeID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.RoadmapNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]any{"status": domain.QuestStatusDone, "completed_at": now}).Error
}

func (r *RoadmapRepository) insertEdges(tx *gorm.DB, roadmapID, nodeID uint, deps []uint) error {
	if len(deps) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(deps))
	for _, from := range deps {
		if from == nodeID {
			return domain.ErrGraphCycle
		}
		if _, ok := seen[from]; ok {
			continue
		}
		seen[from] = struct{}{}

		var count int64
		if err := tx.Model(&domain.RoadmapNode{}).
			Where("id = ? AND roadmap_id = ?", from, roadmapID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return domain.ErrNotFound
		}

		edge := domain.RoadmapEdge{
			RoadmapID:  roadmapID,
			FromNodeID: from,
			ToNodeID:   nodeID,
		}
		if err := tx.Create(&edge).Error; err != nil {
			return err
		}
	}
	return nil
}
