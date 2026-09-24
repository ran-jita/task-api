package repository

import (
	"context"

	"github.com/ran-jita/task-api/internal/entity"
)

type TaskLogRepository interface {
	Create(ctx context.Context, tx Tx, log *entity.TaskLog) error
}
