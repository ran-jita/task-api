package repository

import (
	"context"

	"github.com/ran-jita/task-api/internal/entity"
)

type TaskFilter struct {
	Status string
	Title  string
	Limit  int
	Page   int
}

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	FindByID(ctx context.Context, id string) (*entity.Task, error)
	FindByUser(ctx context.Context, userID string, filter TaskFilter) ([]entity.Task, int, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id string) error

	UpdateAssignee(ctx context.Context, tx Tx, taskID, assigneeID string) error
}
