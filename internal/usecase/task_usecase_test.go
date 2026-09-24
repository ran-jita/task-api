package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/usecase"
)

func TestTaskUsecase_Create(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	task, err := uc.Create(context.Background(), usecase.CreateTaskInput{
		UserID:      "user-1",
		Title:       "Test Task",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != entity.StatusPending {
		t.Fatalf("expected new task status to be pending, got %s", task.Status)
	}
	if task.UserID != "user-1" {
		t.Fatalf("expected user_id user-1, got %s", task.UserID)
	}
}

func TestTaskUsecase_GetByID_Forbidden(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	task, _ := uc.Create(context.Background(), usecase.CreateTaskInput{UserID: "user-1", Title: "T"})

	// user-2 coba akses task milik user-1 — harus ditolak
	_, err := uc.GetByID(context.Background(), task.ID, "user-2")
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskUsecase_GetByID_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	_, err := uc.GetByID(context.Background(), "nonexistent-id", "user-1")
	if !errors.Is(err, usecase.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskUsecase_Update_OwnerCanUpdate(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	task, _ := uc.Create(context.Background(), usecase.CreateTaskInput{UserID: "user-1", Title: "Old Title"})

	updated, err := uc.Update(context.Background(), usecase.UpdateTaskInput{
		ID:     task.ID,
		UserID: "user-1",
		Title:  "New Title",
		Status: "done",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "New Title" {
		t.Fatalf("expected title updated, got %s", updated.Title)
	}
}

func TestTaskUsecase_Update_NonOwnerForbidden(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	task, _ := uc.Create(context.Background(), usecase.CreateTaskInput{UserID: "user-1", Title: "Title"})

	_, err := uc.Update(context.Background(), usecase.UpdateTaskInput{
		ID:     task.ID,
		UserID: "user-2", // bukan pemilik
		Title:  "Hacked Title",
	})
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskUsecase_Delete_NonOwnerForbidden(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	task, _ := uc.Create(context.Background(), usecase.CreateTaskInput{UserID: "user-1", Title: "Title"})

	err := uc.Delete(context.Background(), task.ID, "user-2")
	if !errors.Is(err, usecase.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskUsecase_Delete_OwnerCanDelete(t *testing.T) {
	repo := newMockTaskRepo()
	uc := usecase.NewTaskUsecase(repo)

	task, _ := uc.Create(context.Background(), usecase.CreateTaskInput{UserID: "user-1", Title: "Title"})

	if err := uc.Delete(context.Background(), task.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := uc.GetByID(context.Background(), task.ID, "user-1")
	if !errors.Is(err, usecase.ErrTaskNotFound) {
		t.Fatalf("expected task to be deleted, got err: %v", err)
	}
}
