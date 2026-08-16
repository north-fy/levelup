package services

import (
	"context"
	"testing"

	"github.com/north-fy/levelup/internal/domain"
)

func TestBranchService_Create(t *testing.T) {
	t.Parallel()

	svc := NewBranchService(newFakeBranchStore())

	branch, err := svc.Create(context.Background(), 1, CreateBranchInput{
		Name:  "Finance",
		Color: "#ff0000",
		Icon:  "chart",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if branch.ID == 0 {
		t.Fatal("expected generated id")
	}
	if branch.UserID != 1 {
		t.Fatalf("expected user id 1, got %d", branch.UserID)
	}
	if branch.Name != "Finance" {
		t.Fatalf("expected name Finance, got %q", branch.Name)
	}
}

func TestBranchService_Create_EmptyName(t *testing.T) {
	t.Parallel()

	svc := NewBranchService(newFakeBranchStore())

	_, err := svc.Create(context.Background(), 1, CreateBranchInput{Name: "  "})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	var validationErr *domain.ValidationError
	if !asValidationError(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestBranchService_Get_WrongOwner(t *testing.T) {
	t.Parallel()

	svc := NewBranchService(newFakeBranchStore())
	branch, err := svc.Create(context.Background(), 1, CreateBranchInput{Name: "Sports"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = svc.Get(context.Background(), 2, branch.ID)
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBranchService_List(t *testing.T) {
	t.Parallel()

	svc := NewBranchService(newFakeBranchStore())
	if _, err := svc.Create(context.Background(), 1, CreateBranchInput{Name: "Finance"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(context.Background(), 1, CreateBranchInput{Name: "Sports"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	branches, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
}

func TestBranchService_Update(t *testing.T) {
	t.Parallel()

	svc := NewBranchService(newFakeBranchStore())
	branch, err := svc.Create(context.Background(), 1, CreateBranchInput{Name: "Finance"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	newName := "Money"
	updated, err := svc.Update(context.Background(), 1, branch.ID, UpdateBranchInput{Name: &newName})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Name != "Money" {
		t.Fatalf("expected name Money, got %q", updated.Name)
	}
}

func TestBranchService_Delete(t *testing.T) {
	t.Parallel()

	svc := NewBranchService(newFakeBranchStore())
	branch, err := svc.Create(context.Background(), 1, CreateBranchInput{Name: "Finance"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := svc.Delete(context.Background(), 1, branch.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := svc.Get(context.Background(), 1, branch.ID); err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func asValidationError(err error, target **domain.ValidationError) bool {
	ve, ok := err.(*domain.ValidationError)
	if ok {
		*target = ve
	}
	return ok
}
