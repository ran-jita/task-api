package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ran-jita/task-api/internal/repository"
	"github.com/ran-jita/task-api/internal/usecase"
)

type mockIdempotencyRepo struct {
	mu      sync.Mutex
	records map[string]*repository.IdempotencyRecord
}

func newMockIdempotencyRepo() *mockIdempotencyRepo {
	return &mockIdempotencyRepo{records: make(map[string]*repository.IdempotencyRecord)}
}

func (m *mockIdempotencyRepo) Reserve(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.records[key]; exists {
		return repository.ErrDuplicateKey
	}
	m.records[key] = &repository.IdempotencyRecord{Key: key, CreatedAt: time.Now()}
	return nil
}

func (m *mockIdempotencyRepo) Complete(ctx context.Context, key string, body []byte, statusCode int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[key]
	if !ok {
		return errors.New("key not found")
	}
	rec.ResponseBody = body
	rec.StatusCode = statusCode
	return nil
}

func (m *mockIdempotencyRepo) Get(ctx context.Context, key string) (*repository.IdempotencyRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[key]
	if !ok {
		return nil, nil
	}
	copied := *rec
	return &copied, nil
}

func (m *mockIdempotencyRepo) Release(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, key)
	return nil
}

func TestIdempotencyUsecase_Execute_ConcurrentSameKey(t *testing.T) {
	repo := newMockIdempotencyRepo()
	uc := usecase.NewIdempotencyUsecase(repo)

	var executionCount int32
	const totalGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(totalGoroutine)
	for i := 0; i < totalGoroutine; i++ {
		go func() {
			defer wg.Done()
			_, _, _, err := uc.Execute(context.Background(), "same-key", func(ctx context.Context) (interface{}, int, error) {
				atomic.AddInt32(&executionCount, 1)
				time.Sleep(5 * time.Millisecond)
				return map[string]string{"id": "task-1"}, http.StatusCreated, nil
			})
			if err != nil && !errors.Is(err, usecase.ErrIdempotencyInProgress) {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&executionCount); got != 1 {
		t.Fatalf("expected fn executed exactly once, got %d", got)
	}
}

func TestIdempotencyUsecase_Execute_SequentialReplay(t *testing.T) {
	repo := newMockIdempotencyRepo()
	uc := usecase.NewIdempotencyUsecase(repo)

	var executionCount int32
	fn := func(ctx context.Context) (interface{}, int, error) {
		atomic.AddInt32(&executionCount, 1)
		return map[string]string{"id": "task-1"}, http.StatusCreated, nil
	}

	_, status1, replayed1, err1 := uc.Execute(context.Background(), "key-seq", fn)
	if err1 != nil {
		t.Fatalf("unexpected error on first call: %v", err1)
	}
	if replayed1 {
		t.Fatal("first call should NOT be replayed")
	}

	_, status2, replayed2, err2 := uc.Execute(context.Background(), "key-seq", fn)
	if err2 != nil {
		t.Fatalf("unexpected error on second call: %v", err2)
	}
	if !replayed2 {
		t.Fatal("second call with same key should be replayed")
	}
	if status1 != status2 {
		t.Fatalf("expected same status code, got %d vs %d", status1, status2)
	}
	if got := atomic.LoadInt32(&executionCount); got != 1 {
		t.Fatalf("expected fn executed exactly once across both calls, got %d", got)
	}
}

func TestIdempotencyUsecase_Execute_ReleaseOnFailure(t *testing.T) {
	repo := newMockIdempotencyRepo()
	uc := usecase.NewIdempotencyUsecase(repo)

	failingFn := func(ctx context.Context) (interface{}, int, error) {
		return nil, 0, errors.New("simulated failure")
	}
	_, _, _, err1 := uc.Execute(context.Background(), "key-fail", failingFn)
	if err1 == nil {
		t.Fatal("expected error from first call")
	}

	var executed bool
	successFn := func(ctx context.Context) (interface{}, int, error) {
		executed = true
		return map[string]string{"id": "task-2"}, http.StatusCreated, nil
	}
	_, _, replayed, err2 := uc.Execute(context.Background(), "key-fail", successFn)
	if err2 != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err2)
	}
	if replayed {
		t.Fatal("retry after failure should NOT be treated as replay — should execute fresh")
	}
	if !executed {
		t.Fatal("expected successFn to actually execute after previous failure released the key")
	}
}
