package usecase

import (
	"context"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/repository"
)

type taskUsecase struct {
	taskRepo repository.TaskRepository
}

func NewTaskUsecase(taskRepo repository.TaskRepository) TaskUsecase {
	return &taskUsecase{taskRepo: taskRepo}
}

func (u *taskUsecase) Create(ctx context.Context, input CreateTaskInput) (*entity.Task, error) {
	task := &entity.Task{
		UserID:      input.UserID,
		Title:       input.Title,
		Description: input.Description,
		Status:      entity.StatusPending,
	}
	if err := u.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (u *taskUsecase) List(ctx context.Context, input ListTaskInput) (*ListTaskOutput, error) {
	filter := repository.TaskFilter{
		Status: input.Status,
		Title:  input.Title,
		Limit:  input.Limit,
		Page:   input.Page,
	}
	tasks, total, err := u.taskRepo.FindByUser(ctx, input.UserID, filter)
	if err != nil {
		return nil, err
	}
	return &ListTaskOutput{Tasks: tasks, Total: total}, nil
}

func (u *taskUsecase) GetByID(ctx context.Context, id, userID string) (*entity.Task, error) {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	if task.UserID != userID {
		return nil, ErrForbidden
	}
	return task, nil
}

func (u *taskUsecase) Update(ctx context.Context, input UpdateTaskInput) (*entity.Task, error) {
	task, err := u.taskRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	if task.UserID != input.UserID {
		return nil, ErrForbidden
	}

	task.Title = input.Title
	task.Description = input.Description
	task.Status = entity.TaskStatus(input.Status)

	if err := u.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (u *taskUsecase) Delete(ctx context.Context, id, userID string) error {
	task, err := u.taskRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	if task.UserID != userID {
		return ErrForbidden
	}
	return u.taskRepo.Delete(ctx, id)
}
