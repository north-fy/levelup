package services

import (
	"context"
	"strings"

	"github.com/north-fy/levelup/internal/domain"
)

const (
	maxBranchNameLen  = 100
	maxBranchDescLen  = 512
	maxBranchColorLen = 20
	maxBranchIconLen  = 50
)

// CreateBranchInput holds the fields for creating a branch.
type CreateBranchInput struct {
	Name        string
	Description string
	Color       string
	Icon        string
}

// UpdateBranchInput holds the optional fields for updating a branch.
type UpdateBranchInput struct {
	Name        *string
	Description *string
	Color       *string
	Icon        *string
}

// BranchService manages user branches.
type BranchService struct {
	branches BranchStore
}

// NewBranchService creates the branch service.
func NewBranchService(branches BranchStore) *BranchService {
	return &BranchService{branches: branches}
}

// Create adds a new branch for the user.
func (s *BranchService) Create(ctx context.Context, userID uint, input CreateBranchInput) (*domain.Branch, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.NewValidationError("branch name is required")
	}
	if len(name) > maxBranchNameLen {
		return nil, domain.NewValidationError("branch name is too long")
	}

	branch := &domain.Branch{
		UserID:      userID,
		Name:        name,
		Description: input.Description,
		Color:       input.Color,
		Icon:        input.Icon,
	}
	if err := s.branches.Create(ctx, branch); err != nil {
		return nil, err
	}
	return branch, nil
}

// List returns all branches of the user.
func (s *BranchService) List(ctx context.Context, userID uint) ([]domain.Branch, error) {
	return s.branches.ListByUser(ctx, userID)
}

// Get returns a branch owned by the user.
func (s *BranchService) Get(ctx context.Context, userID, branchID uint) (*domain.Branch, error) {
	return s.branches.GetByIDAndUser(ctx, branchID, userID)
}

// Update applies the provided changes to a branch.
func (s *BranchService) Update(ctx context.Context, userID, branchID uint, input UpdateBranchInput) (*domain.Branch, error) {
	branch, err := s.branches.GetByIDAndUser(ctx, branchID, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.NewValidationError("branch name cannot be empty")
		}
		if len(name) > maxBranchNameLen {
			return nil, domain.NewValidationError("branch name is too long")
		}
		branch.Name = name
	}
	if input.Description != nil {
		branch.Description = *input.Description
	}
	if input.Color != nil {
		branch.Color = *input.Color
	}
	if input.Icon != nil {
		branch.Icon = *input.Icon
	}

	if err := s.branches.Update(ctx, branch); err != nil {
		return nil, err
	}
	return branch, nil
}

// Delete removes a branch owned by the user.
func (s *BranchService) Delete(ctx context.Context, userID, branchID uint) error {
	branch, err := s.branches.GetByIDAndUser(ctx, branchID, userID)
	if err != nil {
		return err
	}
	return s.branches.Delete(ctx, branch)
}
