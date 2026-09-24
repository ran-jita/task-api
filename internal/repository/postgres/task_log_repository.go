package postgres

import (
	"context"
	"database/sql"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/repository"
)

type taskLogRepository struct {
	db *sql.DB
}

func NewTaskLogRepository(db *sql.DB) repository.TaskLogRepository {
	return &taskLogRepository{db: db}
}

func (r *taskLogRepository) Create(ctx context.Context, tx repository.Tx, log *entity.TaskLog) error {
	query := `INSERT INTO task_logs (id, task_id, changed_by, old_assignee, new_assignee, created_at)
	          VALUES (gen_random_uuid(), $1, $2, $3, $4, now())`
	_, err := unwrap(tx).ExecContext(ctx, query, log.TaskID, log.ChangedBy, log.OldAssignee, log.NewAssignee)
	return err
}
