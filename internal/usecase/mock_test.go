package usecase_test

import (
	"context"
	"errors"

	"github.com/ran-jita/task-api/internal/entity"
	"github.com/ran-jita/task-api/internal/repository"
)

// --- mockTaskRepo ---
type mockTaskRepo struct {
	tasks map[string]*entity.Task
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{tasks: make(map[string]*entity.Task)}
}

func (m *mockTaskRepo) Create(ctx context.Context, task *entity.Task) error {
	task.ID = "task-generated-id"
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id string) (*entity.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, nil
	}
	copied := *t
	return &copied, nil
}

func (m *mockTaskRepo) FindByUser(ctx context.Context, userID string, filter repository.TaskFilter) ([]entity.Task, int, error) {
	var result []entity.Task
	for _, t := range m.tasks {
		if t.UserID == userID {
			result = append(result, *t)
		}
	}
	return result, len(result), nil
}

func (m *mockTaskRepo) Update(ctx context.Context, task *entity.Task) error {
	if _, ok := m.tasks[task.ID]; !ok {
		return errors.New("not found")
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) UpdateAssignee(ctx context.Context, tx repository.Tx, taskID, assigneeID string) error {
	t, ok := m.tasks[taskID]
	if !ok {
		return errors.New("not found")
	}
	t.AssigneeID = &assigneeID
	return nil
}

// --- mockUserRepo ---
type mockUserRepo struct {
	users map[string]*entity.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*entity.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

// --- mockTaskLogRepo ---
type mockTaskLogRepo struct {
	logs []entity.TaskLog
}

func newMockTaskLogRepo() *mockTaskLogRepo {
	return &mockTaskLogRepo{}
}

func (m *mockTaskLogRepo) Create(ctx context.Context, tx repository.Tx, log *entity.TaskLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

// --- mockTx & mockTxManager ---
type mockTx struct {
	committed  bool
	rolledBack bool
	failCommit bool
}

func (t *mockTx) Commit() error {
	if t.failCommit {
		return errors.New("simulated commit failure")
	}
	t.committed = true
	return nil
}

func (t *mockTx) Rollback() error {
	if !t.committed {
		t.rolledBack = true
	}
	return nil
}

type mockTxManager struct {
	failCommit bool
	lastTx     *mockTx
}

func (m *mockTxManager) Begin(ctx context.Context) (repository.Tx, error) {
	tx := &mockTx{failCommit: m.failCommit}
	m.lastTx = tx
	return tx, nil
}
