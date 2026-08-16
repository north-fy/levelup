package services

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
)

const (
	maxQuestTitleLen = 255
	maxQuestDescLen  = 1024
)

// CreateQuestInput holds the fields for creating a quest.
type CreateQuestInput struct {
	Title         string
	Description   string
	Type          domain.QuestType
	RewardXP      int
	RewardGold    int
	DurationHours int
}

// UpdateQuestInput holds the optional fields for updating a quest.
type UpdateQuestInput struct {
	Title         *string
	Description   *string
	RewardXP      *int
	RewardGold    *int
	DurationHours *int
}

// QuestService manages quests and their lifecycle.
type QuestService struct {
	quests   QuestStore
	branches BranchStore
	users    UserStore
	events   QuestEventPublisher
	cache    cache.Cache
}

// NewQuestService creates the quest service.
func NewQuestService(quests QuestStore, branches BranchStore, users UserStore, events QuestEventPublisher, c cache.Cache) *QuestService {
	return &QuestService{
		quests:   quests,
		branches: branches,
		users:    users,
		events:   events,
		cache:    c,
	}
}

// Create adds a quest to a branch owned by the user.
func (s *QuestService) Create(ctx context.Context, userID, branchID uint, input CreateQuestInput) (*domain.Quest, error) {
	if _, err := s.branches.GetByIDAndUser(ctx, branchID, userID); err != nil {
		return nil, err
	}
	if err := validateQuestInput(input); err != nil {
		return nil, err
	}

	quest := &domain.Quest{
		BranchID:      branchID,
		UserID:        userID,
		Title:         strings.TrimSpace(input.Title),
		Description:   input.Description,
		Type:          input.Type,
		RewardXP:      input.RewardXP,
		RewardGold:    input.RewardGold,
		DurationHours: input.DurationHours,
		Status:        domain.QuestStatusTodo,
	}
	if err := s.quests.Create(ctx, quest); err != nil {
		return nil, err
	}
	return quest, nil
}

// List returns all quests of a branch owned by the user.
func (s *QuestService) List(ctx context.Context, userID, branchID uint) ([]domain.Quest, error) {
	if _, err := s.branches.GetByIDAndUser(ctx, branchID, userID); err != nil {
		return nil, err
	}
	return s.quests.ListByBranchAndUser(ctx, branchID, userID)
}

// Get returns a quest owned by the user.
func (s *QuestService) Get(ctx context.Context, userID, questID uint) (*domain.Quest, error) {
	return s.quests.GetByIDAndUser(ctx, questID, userID)
}

// Update applies the provided changes to a quest.
func (s *QuestService) Update(ctx context.Context, userID, questID uint, input UpdateQuestInput) (*domain.Quest, error) {
	quest, err := s.quests.GetByIDAndUser(ctx, questID, userID)
	if err != nil {
		return nil, err
	}
	if quest.Status == domain.QuestStatusDone {
		return nil, domain.ErrQuestAlreadyDone
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, domain.NewValidationError("quest title cannot be empty")
		}
		if len(title) > maxQuestTitleLen {
			return nil, domain.NewValidationError("quest title is too long")
		}
		quest.Title = title
	}
	if input.Description != nil {
		quest.Description = *input.Description
	}
	if input.RewardXP != nil {
		if *input.RewardXP < 0 {
			return nil, domain.NewValidationError("reward xp cannot be negative")
		}
		quest.RewardXP = *input.RewardXP
	}
	if input.RewardGold != nil {
		if *input.RewardGold < 0 {
			return nil, domain.NewValidationError("reward gold cannot be negative")
		}
		quest.RewardGold = *input.RewardGold
	}
	if input.DurationHours != nil {
		if *input.DurationHours < 0 {
			return nil, domain.NewValidationError("duration hours cannot be negative")
		}
		quest.DurationHours = *input.DurationHours
	}

	if err := s.quests.Update(ctx, quest); err != nil {
		return nil, err
	}
	return quest, nil
}

// Delete removes a quest owned by the user.
func (s *QuestService) Delete(ctx context.Context, userID, questID uint) error {
	quest, err := s.quests.GetByIDAndUser(ctx, questID, userID)
	if err != nil {
		return err
	}
	return s.quests.Delete(ctx, quest)
}

// Complete finishes a simple quest and grants its reward.
func (s *QuestService) Complete(ctx context.Context, userID, questID uint) (*domain.Quest, error) {
	quest, err := s.quests.GetByIDAndUser(ctx, questID, userID)
	if err != nil {
		return nil, err
	}
	if quest.IsTimed() {
		return nil, domain.ErrTimedQuestIncomplete
	}
	if quest.Status != domain.QuestStatusTodo {
		return nil, domain.ErrQuestAlreadyDone
	}

	now := time.Now()
	quest.Status = domain.QuestStatusDone
	quest.CompletedAt = &now

	if err := s.quests.Update(ctx, quest); err != nil {
		return nil, err
	}
	if err := s.awardRewards(ctx, userID, quest.RewardXP, quest.RewardGold); err != nil {
		return nil, err
	}
	if err := s.events.PublishQuestCompleted(ctx, domain.QuestCompletedEvent{
		UserID:      userID,
		QuestID:     quest.ID,
		BranchID:    quest.BranchID,
		XP:          quest.RewardXP,
		Gold:        quest.RewardGold,
		Hours:       0,
		CompletedAt: now,
	}); err != nil {
		return nil, err
	}
	return quest, nil
}

// Start begins the time tracker for a timed quest.
func (s *QuestService) Start(ctx context.Context, userID, questID uint) (*domain.Quest, error) {
	quest, err := s.quests.GetByIDAndUser(ctx, questID, userID)
	if err != nil {
		return nil, err
	}
	if !quest.IsTimed() {
		return nil, domain.ErrForbidden
	}
	if quest.Status == domain.QuestStatusDone || quest.Status == domain.QuestStatusCancelled {
		return nil, domain.ErrQuestAlreadyDone
	}
	if quest.Status == domain.QuestStatusInProgress {
		return nil, domain.ErrQuestAlreadyStarted
	}

	active, err := s.quests.HasActiveTimer(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, domain.ErrActiveTimerConflict
	}

	now := time.Now()
	quest.Status = domain.QuestStatusInProgress
	quest.StartedAt = &now
	if err := s.quests.Update(ctx, quest); err != nil {
		return nil, err
	}
	return quest, nil
}

// Stop ends the time tracker, completing the quest and granting a proportional reward.
func (s *QuestService) Stop(ctx context.Context, userID, questID uint) (*domain.Quest, error) {
	quest, err := s.quests.GetByIDAndUser(ctx, questID, userID)
	if err != nil {
		return nil, err
	}
	if !quest.IsTimed() {
		return nil, domain.ErrForbidden
	}
	if quest.Status != domain.QuestStatusInProgress {
		return nil, domain.ErrQuestNotInProgress
	}
	if quest.StartedAt == nil {
		return nil, domain.ErrQuestNotInProgress
	}

	elapsed := time.Since(*quest.StartedAt)
	progress := progressRatio(elapsed, quest.DurationHours)

	xpAwarded := int(math.Round(float64(quest.RewardXP) * progress))
	goldAwarded := int(math.Round(float64(quest.RewardGold) * progress))
	hours := hoursSpent(elapsed)

	now := time.Now()
	quest.Status = domain.QuestStatusDone
	quest.CompletedAt = &now

	if err := s.quests.Update(ctx, quest); err != nil {
		return nil, err
	}
	if err := s.awardRewards(ctx, userID, xpAwarded, goldAwarded); err != nil {
		return nil, err
	}
	if err := s.events.PublishQuestCompleted(ctx, domain.QuestCompletedEvent{
		UserID:      userID,
		QuestID:     quest.ID,
		BranchID:    quest.BranchID,
		XP:          xpAwarded,
		Gold:        goldAwarded,
		Hours:       hours,
		CompletedAt: now,
	}); err != nil {
		return nil, err
	}
	return quest, nil
}

func (s *QuestService) awardRewards(ctx context.Context, userID uint, xp, gold int) error {
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
	return nil
}

func validateQuestInput(input CreateQuestInput) error {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return domain.NewValidationError("quest title is required")
	}
	if len(title) > maxQuestTitleLen {
		return domain.NewValidationError("quest title is too long")
	}
	if input.Type != domain.QuestTypeSimple && input.Type != domain.QuestTypeTimed {
		return domain.NewValidationError("quest type must be simple or timed")
	}
	if input.RewardXP < 0 {
		return domain.NewValidationError("reward xp cannot be negative")
	}
	if input.RewardGold < 0 {
		return domain.NewValidationError("reward gold cannot be negative")
	}
	if input.Type == domain.QuestTypeTimed && input.DurationHours <= 0 {
		return domain.NewValidationError("timed quest requires a positive duration")
	}
	return nil
}

func progressRatio(elapsed time.Duration, durationHours int) float64 {
	required := time.Duration(durationHours) * time.Hour
	if required <= 0 {
		return 1
	}
	progress := float64(elapsed) / float64(required)
	if progress > 1 {
		return 1
	}
	if progress < 0 {
		return 0
	}
	return progress
}

func hoursSpent(elapsed time.Duration) int {
	hours := int(math.Round(elapsed.Hours()))
	if hours < 1 && elapsed > 0 {
		return 1
	}
	return hours
}
