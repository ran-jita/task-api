package usecase

import (
	"context"
	"fmt"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/repository"
)

type assignUsecase struct {
	taskRepo    repository.TaskRepository
	taskLogRepo repository.TaskLogRepository
	userRepo    repository.UserRepository
	txManager   repository.TxManager
}

func NewAssignUsecase(
	taskRepo repository.TaskRepository,
	taskLogRepo repository.TaskLogRepository,
	userRepo repository.UserRepository,
	txManager repository.TxManager,
) AssignUsecase {
	return &assignUsecase{
		taskRepo:    taskRepo,
		taskLogRepo: taskLogRepo,
		userRepo:    userRepo,
		txManager:   txManager,
	}
}

func (u *assignUsecase) Assign(ctx context.Context, input AssignTaskInput) error {
	task, err := u.taskRepo.FindByID(ctx, input.TaskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	assignee, err := u.userRepo.FindByID(ctx, input.AssigneeID)
	if err != nil {
		return err
	}
	if assignee == nil {
		return ErrAssigneeNotFound
	}

	oldAssignee := task.AssigneeID

	tx, err := u.txManager.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := u.taskRepo.UpdateAssignee(ctx, tx, input.TaskID, input.AssigneeID); err != nil {
		return fmt.Errorf("update assignee: %w", err)
	}

	log := &entity.TaskLog{
		TaskID:      input.TaskID,
		ChangedBy:   input.RequestedBy,
		OldAssignee: oldAssignee,
		NewAssignee: input.AssigneeID,
	}
	if err := u.taskLogRepo.Create(ctx, tx, log); err != nil {
		return fmt.Errorf("create task log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	sendMockNotification(assignee.Email, input.TaskID)

	return nil
}

func sendMockNotification(email, taskID string) {
	fmt.Printf("[MOCK NOTIFICATION] Sending notification to %s: you've been assigned task %s\n", email, taskID)
}
