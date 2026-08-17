package services

import (
	"context"
	"strings"
	"time"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
	"github.com/north-fy/levelup/internal/pkg/metrics"
)

// CreateRoadmapInput holds the fields for creating a roadmap.
type CreateRoadmapInput struct {
	Title       string
	Description string
}

// UpdateRoadmapInput holds the optional fields for updating a roadmap.
type UpdateRoadmapInput struct {
	Title       *string
	Description *string
}

// AddNodeInput holds the fields for adding a roadmap node.
type AddNodeInput struct {
	Title         string
	Description   string
	Type          domain.QuestType
	RewardXP      int
	RewardGold    int
	DurationHours int
	Dependencies  []uint
}

// UpdateNodeInput holds the optional fields for updating a roadmap node.
type UpdateNodeInput struct {
	Title         *string
	Description   *string
	RewardXP      *int
	RewardGold    *int
	DurationHours *int
	Dependencies  *[]uint
}

// RoadmapService manages roadmaps and their dependency graphs.
type RoadmapService struct {
	roadmaps RoadmapStore
	users    UserStore
	events   QuestEventPublisher
	cache    cache.Cache
}

// NewRoadmapService creates the roadmap service.
func NewRoadmapService(roadmaps RoadmapStore, users UserStore, events QuestEventPublisher, c cache.Cache) *RoadmapService {
	return &RoadmapService{
		roadmaps: roadmaps,
		users:    users,
		events:   events,
		cache:    c,
	}
}

// Create adds a personal roadmap for the user.
func (s *RoadmapService) Create(ctx context.Context, userID uint, input CreateRoadmapInput) (*domain.Roadmap, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, domain.NewValidationError("roadmap title is required")
	}
	if len(title) > 255 {
		return nil, domain.NewValidationError("roadmap title is too long")
	}

	roadmap := &domain.Roadmap{
		UserID:      userID,
		Title:       title,
		Description: input.Description,
		SourceType:  domain.RoadmapSourceOwn,
	}
	if err := s.roadmaps.Create(ctx, roadmap); err != nil {
		return nil, err
	}
	return roadmap, nil
}

// List returns all roadmaps of the user.
func (s *RoadmapService) List(ctx context.Context, userID uint) ([]domain.Roadmap, error) {
	return s.roadmaps.ListByUser(ctx, userID)
}

// Get returns a roadmap with its nodes and edges.
func (s *RoadmapService) Get(ctx context.Context, userID, roadmapID uint) (*domain.RoadmapDetail, error) {
	roadmap, err := s.roadmaps.GetByIDAndUser(ctx, roadmapID, userID)
	if err != nil {
		return nil, err
	}
	nodes, err := s.roadmaps.ListNodesByRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	edges, err := s.roadmaps.ListEdgesByRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	return &domain.RoadmapDetail{
		Roadmap: *roadmap,
		Nodes:   nodes,
		Edges:   edges,
	}, nil
}

// Update applies the provided changes to a roadmap.
func (s *RoadmapService) Update(ctx context.Context, userID, roadmapID uint, input UpdateRoadmapInput) (*domain.Roadmap, error) {
	roadmap, err := s.roadmaps.GetByIDAndUser(ctx, roadmapID, userID)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, domain.NewValidationError("roadmap title cannot be empty")
		}
		roadmap.Title = title
	}
	if input.Description != nil {
		roadmap.Description = *input.Description
	}
	if err := s.roadmaps.Update(ctx, roadmap); err != nil {
		return nil, err
	}
	return roadmap, nil
}

// Delete removes a roadmap owned by the user.
func (s *RoadmapService) Delete(ctx context.Context, userID, roadmapID uint) error {
	roadmap, err := s.roadmaps.GetByIDAndUser(ctx, roadmapID, userID)
	if err != nil {
		return err
	}
	return s.roadmaps.Delete(ctx, roadmap)
}

// AddNode appends a node to a roadmap with its dependencies.
func (s *RoadmapService) AddNode(ctx context.Context, userID, roadmapID uint, input AddNodeInput) (*domain.RoadmapNode, error) {
	if _, err := s.roadmaps.GetByIDAndUser(ctx, roadmapID, userID); err != nil {
		return nil, err
	}
	if err := validateNodeInput(input.Title, input.Type, input.RewardXP, input.RewardGold, input.DurationHours); err != nil {
		return nil, err
	}

	nodes, err := s.roadmaps.ListNodesByRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	if err := validateDependencies(nodes, input.Dependencies); err != nil {
		return nil, err
	}

	node := &domain.RoadmapNode{
		RoadmapID:     roadmapID,
		Title:         strings.TrimSpace(input.Title),
		Description:   input.Description,
		Position:      len(nodes),
		Type:          input.Type,
		RewardXP:      input.RewardXP,
		RewardGold:    input.RewardGold,
		DurationHours: input.DurationHours,
		Status:        domain.QuestStatusTodo,
	}
	if err := s.roadmaps.AddNode(ctx, roadmapID, node, input.Dependencies); err != nil {
		return nil, err
	}
	return node, nil
}

// UpdateNode applies the provided changes to a roadmap node.
func (s *RoadmapService) UpdateNode(ctx context.Context, userID, roadmapID, nodeID uint, input UpdateNodeInput) (*domain.RoadmapNode, error) {
	if _, err := s.roadmaps.GetByIDAndUser(ctx, roadmapID, userID); err != nil {
		return nil, err
	}
	node, err := s.roadmaps.GetNodeByIDAndUser(ctx, nodeID, userID)
	if err != nil {
		return nil, err
	}
	if node.RoadmapID != roadmapID {
		return nil, domain.ErrNotFound
	}
	if node.Status == domain.QuestStatusDone {
		return nil, domain.ErrQuestAlreadyDone
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, domain.NewValidationError("node title cannot be empty")
		}
		node.Title = title
	}
	if input.Description != nil {
		node.Description = *input.Description
	}
	if input.RewardXP != nil {
		if *input.RewardXP < 0 {
			return nil, domain.NewValidationError("reward xp cannot be negative")
		}
		node.RewardXP = *input.RewardXP
	}
	if input.RewardGold != nil {
		if *input.RewardGold < 0 {
			return nil, domain.NewValidationError("reward gold cannot be negative")
		}
		node.RewardGold = *input.RewardGold
	}
	if input.DurationHours != nil {
		if *input.DurationHours < 0 {
			return nil, domain.NewValidationError("duration hours cannot be negative")
		}
		node.DurationHours = *input.DurationHours
	}

	if input.Dependencies != nil {
		nodes, err := s.roadmaps.ListNodesByRoadmap(ctx, roadmapID)
		if err != nil {
			return nil, err
		}
		if err := validateDependencies(nodes, *input.Dependencies); err != nil {
			return nil, err
		}
		if err := s.checkNoCycle(ctx, roadmapID, nodes, nodeID, *input.Dependencies); err != nil {
			return nil, err
		}
		if err := s.roadmaps.UpdateNodeDeps(ctx, roadmapID, nodeID, *input.Dependencies); err != nil {
			return nil, err
		}
	}

	if err := s.roadmaps.UpdateNode(ctx, node); err != nil {
		return nil, err
	}
	return node, nil
}

// CompleteNode finishes a node once all its dependencies are done.
func (s *RoadmapService) CompleteNode(ctx context.Context, userID, roadmapID, nodeID uint) (*domain.RoadmapNode, error) {
	if _, err := s.roadmaps.GetByIDAndUser(ctx, roadmapID, userID); err != nil {
		return nil, err
	}
	node, err := s.roadmaps.GetNodeByIDAndUser(ctx, nodeID, userID)
	if err != nil {
		return nil, err
	}
	if node.RoadmapID != roadmapID {
		return nil, domain.ErrNotFound
	}
	if node.Status == domain.QuestStatusDone {
		return nil, domain.ErrQuestAlreadyDone
	}

	nodes, err := s.roadmaps.ListNodesByRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	edges, err := s.roadmaps.ListEdgesByRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}

	graph := buildGraph(nodes, edges)
	for _, dep := range graph.Predecessors(nodeID) {
		if statusOfNode(nodes, dep) != domain.QuestStatusDone {
			return nil, domain.ErrPrerequisitesNotMet
		}
	}

	if err := s.roadmaps.MarkNodeDone(ctx, nodeID); err != nil {
		return nil, err
	}
	node.Status = domain.QuestStatusDone
	now := time.Now()
	node.CompletedAt = &now

	if err := s.awardRewards(ctx, userID, node.RewardXP, node.RewardGold); err != nil {
		return nil, err
	}
	if err := s.events.PublishQuestCompleted(ctx, domain.QuestCompletedEvent{
		UserID:      userID,
		BranchID:    0,
		RoadmapID:   roadmapID,
		XP:          node.RewardXP,
		Gold:        node.RewardGold,
		Hours:       0,
		CompletedAt: now,
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *RoadmapService) checkNoCycle(ctx context.Context, roadmapID uint, nodes []domain.RoadmapNode, nodeID uint, deps []uint) error {
	edges, err := s.roadmaps.ListEdgesByRoadmap(ctx, roadmapID)
	if err != nil {
		return err
	}

	graph := domain.NewGraph()
	for _, n := range nodes {
		graph.AddNode(n.ID)
	}
	for _, e := range edges {
		if e.ToNodeID == nodeID {
			continue
		}
		if err := graph.AddEdge(e.FromNodeID, e.ToNodeID); err != nil {
			return err
		}
	}
	for _, dep := range deps {
		if err := graph.AddEdge(dep, nodeID); err != nil {
			return err
		}
	}
	return nil
}

func (s *RoadmapService) awardRewards(ctx context.Context, userID uint, xp, gold int) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.XP += xp
	user.Gold += gold
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}
	invalidateUser(ctx, s.cache, userID)
	metrics.NodesCompleted.Inc()
	return nil
}

func validateNodeInput(title string, nodeType domain.QuestType, xp, gold, duration int) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.NewValidationError("node title is required")
	}
	if len(title) > 255 {
		return domain.NewValidationError("node title is too long")
	}
	if nodeType != domain.QuestTypeSimple && nodeType != domain.QuestTypeTimed {
		return domain.NewValidationError("node type must be simple or timed")
	}
	if xp < 0 {
		return domain.NewValidationError("reward xp cannot be negative")
	}
	if gold < 0 {
		return domain.NewValidationError("reward gold cannot be negative")
	}
	if nodeType == domain.QuestTypeTimed && duration <= 0 {
		return domain.NewValidationError("timed node requires a positive duration")
	}
	return nil
}

func validateDependencies(nodes []domain.RoadmapNode, deps []uint) error {
	valid := make(map[uint]struct{}, len(nodes))
	for _, n := range nodes {
		valid[n.ID] = struct{}{}
	}
	for _, dep := range deps {
		if _, ok := valid[dep]; !ok {
			return domain.NewValidationError("dependency node does not exist in this roadmap")
		}
	}
	return nil
}

func buildGraph(nodes []domain.RoadmapNode, edges []domain.RoadmapEdge) *domain.Graph {
	graph := domain.NewGraph()
	for _, n := range nodes {
		graph.AddNode(n.ID)
	}
	for _, e := range edges {
		_ = graph.AddEdge(e.FromNodeID, e.ToNodeID)
	}
	return graph
}

func statusOfNode(nodes []domain.RoadmapNode, id uint) domain.QuestStatus {
	for _, n := range nodes {
		if n.ID == id {
			return n.Status
		}
	}
	return domain.QuestStatusCancelled
}
