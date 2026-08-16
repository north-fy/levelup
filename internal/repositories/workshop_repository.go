package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/domain"
)

// WorkshopRepository persists published roadmaps and copies graphs on install.
type WorkshopRepository struct {
	db *gorm.DB
}

// NewWorkshopRepository creates the repository.
func NewWorkshopRepository(db *gorm.DB) *WorkshopRepository {
	return &WorkshopRepository{db: db}
}

// Create stores a new workshop roadmap.
func (r *WorkshopRepository) Create(ctx context.Context, workshop *domain.WorkshopRoadmap) error {
	return r.db.WithContext(ctx).Create(workshop).Error
}

// GetByID returns a workshop roadmap by id.
func (r *WorkshopRepository) GetByID(ctx context.Context, id uint) (*domain.WorkshopRoadmap, error) {
	var workshop domain.WorkshopRoadmap
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&workshop).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &workshop, nil
}

// GetByIDAndAuthor returns a workshop roadmap authored by the given user.
func (r *WorkshopRepository) GetByIDAndAuthor(ctx context.Context, id, authorID uint) (*domain.WorkshopRoadmap, error) {
	var workshop domain.WorkshopRoadmap
	if err := r.db.WithContext(ctx).Where("id = ? AND author_id = ?", id, authorID).First(&workshop).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &workshop, nil
}

// ListPublished returns all published workshop roadmaps.
func (r *WorkshopRepository) ListPublished(ctx context.Context) ([]domain.WorkshopRoadmap, error) {
	var workshops []domain.WorkshopRoadmap
	if err := r.db.WithContext(ctx).Where("is_published = ?", true).Order("created_at DESC").Find(&workshops).Error; err != nil {
		return nil, err
	}
	return workshops, nil
}

// ListByAuthor returns all workshop roadmaps authored by the user.
func (r *WorkshopRepository) ListByAuthor(ctx context.Context, authorID uint) ([]domain.WorkshopRoadmap, error) {
	var workshops []domain.WorkshopRoadmap
	if err := r.db.WithContext(ctx).Where("author_id = ?", authorID).Order("created_at DESC").Find(&workshops).Error; err != nil {
		return nil, err
	}
	return workshops, nil
}

// Update persists workshop roadmap changes.
func (r *WorkshopRepository) Update(ctx context.Context, workshop *domain.WorkshopRoadmap) error {
	return r.db.WithContext(ctx).Save(workshop).Error
}

// InstallCopy deep-copies the workshop graph into a personal roadmap.
func (r *WorkshopRepository) InstallCopy(ctx context.Context, installerID uint, workshop *domain.WorkshopRoadmap) (*domain.Roadmap, error) {
	var installed domain.Roadmap

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var source domain.Roadmap
		if err := tx.Where("id = ?", workshop.SourceRoadmapID).First(&source).Error; err != nil {
			return err
		}

		var nodes []domain.RoadmapNode
		if err := tx.Where("roadmap_id = ?", source.ID).Order("position ASC, id ASC").Find(&nodes).Error; err != nil {
			return err
		}

		var edges []domain.RoadmapEdge
		if err := tx.Where("roadmap_id = ?", source.ID).Find(&edges).Error; err != nil {
			return err
		}

		installed = domain.Roadmap{
			UserID:      installerID,
			Title:       workshop.Title,
			Description: workshop.Description,
			SourceType:  domain.RoadmapSourceWorkshop,
			SourceID:    workshop.ID,
		}
		if err := tx.Create(&installed).Error; err != nil {
			return err
		}

		idMap := make(map[uint]uint, len(nodes))
		for i := range nodes {
			oldID := nodes[i].ID
			nodes[i].ID = 0
			nodes[i].RoadmapID = installed.ID
			nodes[i].Status = domain.QuestStatusTodo
			nodes[i].CompletedAt = nil
			if err := tx.Create(&nodes[i]).Error; err != nil {
				return err
			}
			idMap[oldID] = nodes[i].ID
		}

		for _, edge := range edges {
			from, fromOK := idMap[edge.FromNodeID]
			to, toOK := idMap[edge.ToNodeID]
			if !fromOK || !toOK {
				return domain.ErrGraphCycle
			}
			edge.ID = 0
			edge.RoadmapID = installed.ID
			edge.FromNodeID = from
			edge.ToNodeID = to
			if err := tx.Create(&edge).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &installed, nil
}
