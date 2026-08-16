package services

import (
	"context"
	"errors"
	"testing"

	"github.com/north-fy/levelup/internal/domain"
)

func TestUserServiceGetByID(t *testing.T) {
	users := newFakeUserStore()
	svc := NewUserService(users)

	created := &domain.User{Email: "user@example.com", Nickname: "Hero", XP: 500}
	if err := users.Create(context.Background(), created); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Nickname != "Hero" {
		t.Errorf("nickname = %q", got.Nickname)
	}
	if got.Level != domain.LevelFor(500) {
		t.Errorf("level = %d, want %d", got.Level, domain.LevelFor(500))
	}
}

func TestUserServiceGetByIDNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserStore())

	_, err := svc.GetByID(context.Background(), 42)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUserServiceUpdate(t *testing.T) {
	users := newFakeUserStore()
	svc := NewUserService(users)

	created := &domain.User{Email: "user@example.com", Nickname: "Hero"}
	if err := users.Create(context.Background(), created); err != nil {
		t.Fatalf("create: %v", err)
	}

	status := "online"
	avatar := "https://example.com/avatar.png"
	updated, err := svc.Update(context.Background(), created.ID, UpdateProfileInput{
		Status:    &status,
		AvatarURL: &avatar,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Status != status {
		t.Errorf("status = %q, want %q", updated.Status, status)
	}
	if updated.AvatarURL != avatar {
		t.Errorf("avatar = %q, want %q", updated.AvatarURL, avatar)
	}
	if updated.Nickname != "Hero" {
		t.Errorf("nickname should be untouched, got %q", updated.Nickname)
	}
}

func TestUserServiceUpdateEmptyNickname(t *testing.T) {
	users := newFakeUserStore()
	svc := NewUserService(users)

	created := &domain.User{Email: "user@example.com", Nickname: "Hero"}
	if err := users.Create(context.Background(), created); err != nil {
		t.Fatalf("create: %v", err)
	}

	nickname := "   "
	_, err := svc.Update(context.Background(), created.ID, UpdateProfileInput{Nickname: &nickname})
	if err == nil {
		t.Error("expected error for empty nickname")
	}
}
