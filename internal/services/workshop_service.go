package services

import (
	"context"
	"strings"

	"github.com/north-fy/levelup/internal/domain"
)

// CreateWorkshopInput holds the fields for publishing a roadmap.
type CreateWorkshopInput struct {
	RoadmapID   uint
	Title       string
	Description string
}

// UpdateWorkshopInput holds the optional fields for updating a workshop roadmap.
type UpdateWorkshopInput struct {
	Title       *string
	Description *string
	IsPublished *bool
}

// WorkshopService manages published roadmaps.
type WorkshopService struct {
	workshops WorkshopStore
	roadmaps  RoadmapStore
}

// NewWorkshopService creates the workshop service.
func NewWorkshopService(workshops WorkshopStore, roadmaps RoadmapStore) *WorkshopService {
	return &WorkshopService{
		workshops: workshops,
		roadmaps:  roadmaps,
	}
}

// Create publishes one of the author's roadmaps to the workshop.
func (s *WorkshopService) Create(ctx context.Context, authorID uint, input CreateWorkshopInput) (*domain.WorkshopRoadmap, error) {
	roadmap, err := s.roadmaps.GetByIDAndUser(ctx, input.RoadmapID, authorID)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = roadmap.Title
	}
	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = roadmap.Description
	}

	workshop := &domain.WorkshopRoadmap{
		AuthorID:        authorID,
		SourceRoadmapID: input.RoadmapID,
		Title:           title,
		Description:     description,
		IsPublished:     true,
	}
	if err := s.workshops.Create(ctx, workshop); err != nil {
		return nil, err
	}
	return workshop, nil
}

// List returns published workshop roadmaps, or the author's own when mine is set.
func (s *WorkshopService) List(ctx context.Context, userID uint, mine bool) ([]domain.WorkshopRoadmap, error) {
	if mine {
		return s.workshops.ListByAuthor(ctx, userID)
	}
	return s.workshops.ListPublished(ctx)
}

// Update applies the provided changes to the author's workshop roadmap.
func (s *WorkshopService) Update(ctx context.Context, authorID, workshopID uint, input UpdateWorkshopInput) (*domain.WorkshopRoadmap, error) {
	workshop, err := s.workshops.GetByIDAndAuthor(ctx, workshopID, authorID)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, domain.NewValidationError("workshop title cannot be empty")
		}
		workshop.Title = title
	}
	if input.Description != nil {
		workshop.Description = *input.Description
	}
	if input.IsPublished != nil {
		workshop.IsPublished = *input.IsPublished
	}
	if err := s.workshops.Update(ctx, workshop); err != nil {
		return nil, err
	}
	return workshop, nil
}

// Install deep-copies a published workshop roadmap into the user's roadmaps.
func (s *WorkshopService) Install(ctx context.Context, userID, workshopID uint) (*domain.RoadmapDetail, error) {
	workshop, err := s.workshops.GetByID(ctx, workshopID)
	if err != nil {
		return nil, err
	}
	if !workshop.IsPublished {
		return nil, domain.ErrWorkshopNotPublished
	}

	installed, err := s.workshops.InstallCopy(ctx, userID, workshop)
	if err != nil {
		return nil, err
	}

	nodes, err := s.roadmaps.ListNodesByRoadmap(ctx, installed.ID)
	if err != nil {
		return nil, err
	}
	edges, err := s.roadmaps.ListEdgesByRoadmap(ctx, installed.ID)
	if err != nil {
		return nil, err
	}
	return &domain.RoadmapDetail{
		Roadmap: *installed,
		Nodes:   nodes,
		Edges:   edges,
	}, nil
}
