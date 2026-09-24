package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/repository"
)

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) repository.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *entity.Task) error {
	query := `INSERT INTO tasks (id, user_id, assignee_id, title, description, status, created_at, updated_at)
	          VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, now(), now())
	          RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		task.UserID, task.AssigneeID, task.Title, task.Description, task.Status,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*entity.Task, error) {
	query := `SELECT id, user_id, assignee_id, title, description, status, created_at, updated_at
	          FROM tasks WHERE id = $1`
	t := &entity.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.AssigneeID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *taskRepository) FindByUser(ctx context.Context, userID string, filter repository.TaskFilter) ([]entity.Task, int, error) {
	conditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argN := 2

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argN))
		args = append(args, filter.Status)
		argN++
	}
	if filter.Title != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argN))
		args = append(args, "%"+filter.Title+"%")
		argN++
	}
	where := strings.Join(conditions, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM tasks WHERE " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, assignee_id, title, description, status, created_at, updated_at
		 FROM tasks WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argN, argN+1,
	)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var t entity.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.AssigneeID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *taskRepository) Update(ctx context.Context, task *entity.Task) error {
	query := `UPDATE tasks SET title=$1, description=$2, status=$3, updated_at=now()
	          WHERE id=$4`
	_, err := r.db.ExecContext(ctx, query, task.Title, task.Description, task.Status, task.ID)
	return err
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	return err
}

func (r *taskRepository) UpdateAssignee(ctx context.Context, tx repository.Tx, taskID, assigneeID string) error {
	query := `UPDATE tasks SET assignee_id=$1, updated_at=now() WHERE id=$2`
	_, err := unwrap(tx).ExecContext(ctx, query, assigneeID, taskID)
	return err
}
