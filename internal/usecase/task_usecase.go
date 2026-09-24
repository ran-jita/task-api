package usecase

import (
	"context"

	"github.com/ran-jita/task-api/internal/entity"
)

type CreateTaskInput struct {
	UserID      string
	Title       string
	Description string
}

type ListTaskInput struct {
	UserID string
	Status string
	Title  string
	Limit  int
	Page   int
}

type ListTaskOutput struct {
	Tasks []entity.Task
	Total int
}

type UpdateTaskInput struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Status      string
}

type TaskUsecase interface {
	Create(ctx context.Context, input CreateTaskInput) (*entity.Task, error)
	List(ctx context.Context, input ListTaskInput) (*ListTaskOutput, error)
	GetByID(ctx context.Context, id, userID string) (*entity.Task, error)
	Update(ctx context.Context, input UpdateTaskInput) (*entity.Task, error)
	Delete(ctx context.Context, id, userID string) error
}
