package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/usecase"
)

func setupAssignUsecase(txManager *mockTxManager) (usecase.AssignUsecase, *mockTaskRepo, *mockUserRepo, *mockTaskLogRepo) {
	taskRepo := newMockTaskRepo()
	userRepo := newMockUserRepo()
	taskLogRepo := newMockTaskLogRepo()

	uc := usecase.NewAssignUsecase(taskRepo, taskLogRepo, userRepo, txManager)
	return uc, taskRepo, userRepo, taskLogRepo
}

func TestAssignUsecase_Assign_Success(t *testing.T) {
	txManager := &mockTxManager{}
	uc, taskRepo, userRepo, taskLogRepo := setupAssignUsecase(txManager)

	taskRepo.tasks["task-1"] = &entity.Task{ID: "task-1", UserID: "owner-1", Title: "T"}
	userRepo.users["assignee-1"] = &entity.User{ID: "assignee-1", Email: "assignee@example.com"}

	err := uc.Assign(context.Background(), usecase.AssignTaskInput{
		TaskID:      "task-1",
		AssigneeID:  "assignee-1",
		RequestedBy: "owner-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Transaction harus COMMIT, bukan rollback
	if !txManager.lastTx.committed {
		t.Fatal("expected transaction to be committed")
	}
	if txManager.lastTx.rolledBack {
		t.Fatal("transaction should not be rolled back on success")
	}

	// task_logs harus tercatat
	if len(taskLogRepo.logs) != 1 {
		t.Fatalf("expected 1 task log entry, got %d", len(taskLogRepo.logs))
	}
	if taskLogRepo.logs[0].NewAssignee != "assignee-1" {
		t.Fatalf("expected new_assignee assignee-1, got %s", taskLogRepo.logs[0].NewAssignee)
	}
}

func TestAssignUsecase_Assign_TaskNotFound(t *testing.T) {
	txManager := &mockTxManager{}
	uc, _, userRepo, _ := setupAssignUsecase(txManager)

	userRepo.users["assignee-1"] = &entity.User{ID: "assignee-1"}

	err := uc.Assign(context.Background(), usecase.AssignTaskInput{
		TaskID:      "nonexistent-task",
		AssigneeID:  "assignee-1",
		RequestedBy: "owner-1",
	})
	if !errors.Is(err, usecase.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
	// Transaction TIDAK BOLEH pernah dibuka — validasi gagal sebelum Begin()
	if txManager.lastTx != nil {
		t.Fatal("expected no transaction to be started when task not found")
	}
}

func TestAssignUsecase_Assign_AssigneeNotFound(t *testing.T) {
	txManager := &mockTxManager{}
	uc, taskRepo, _, _ := setupAssignUsecase(txManager)

	taskRepo.tasks["task-1"] = &entity.Task{ID: "task-1", UserID: "owner-1"}

	err := uc.Assign(context.Background(), usecase.AssignTaskInput{
		TaskID:      "task-1",
		AssigneeID:  "nonexistent-user",
		RequestedBy: "owner-1",
	})
	if !errors.Is(err, usecase.ErrAssigneeNotFound) {
		t.Fatalf("expected ErrAssigneeNotFound, got %v", err)
	}
	if txManager.lastTx != nil {
		t.Fatal("expected no transaction to be started when assignee not found")
	}
}

// Ini test PALING PENTING untuk requirement "transaction dengan rollback":
// kalau Commit gagal, task_logs yang sudah di-insert TIDAK BOLEH dianggap final.
func TestAssignUsecase_Assign_RollbackOnCommitFailure(t *testing.T) {
	txManager := &mockTxManager{failCommit: true}
	uc, taskRepo, userRepo, _ := setupAssignUsecase(txManager)

	taskRepo.tasks["task-1"] = &entity.Task{ID: "task-1", UserID: "owner-1"}
	userRepo.users["assignee-1"] = &entity.User{ID: "assignee-1", Email: "a@example.com"}

	err := uc.Assign(context.Background(), usecase.AssignTaskInput{
		TaskID:      "task-1",
		AssigneeID:  "assignee-1",
		RequestedBy: "owner-1",
	})
	if err == nil {
		t.Fatal("expected error when commit fails, got nil")
	}
	if txManager.lastTx.committed {
		t.Fatal("transaction should NOT be marked committed when Commit() fails")
	}
	if !txManager.lastTx.rolledBack {
		t.Fatal("expected rollback to be triggered after commit failure")
	}
}
