package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/ran-jita/task-api/internal/repository"
)

type idempotencyRepository struct {
	db *sql.DB
}

func NewIdempotencyRepository(db *sql.DB) repository.IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

func (r *idempotencyRepository) Reserve(ctx context.Context, key string) error {
	query := `INSERT INTO idempotency_keys (key, created_at) VALUES ($1, now())`
	_, err := r.db.ExecContext(ctx, query, key)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return repository.ErrDuplicateKey
	}
	return err
}

func (r *idempotencyRepository) Complete(ctx context.Context, key string, responseBody []byte, statusCode int) error {
	query := `UPDATE idempotency_keys SET response_body=$1, status_code=$2 WHERE key=$3`
	_, err := r.db.ExecContext(ctx, query, responseBody, statusCode, key)
	return err
}

func (r *idempotencyRepository) Get(ctx context.Context, key string) (*repository.IdempotencyRecord, error) {
	query := `SELECT key, response_body, status_code, created_at FROM idempotency_keys WHERE key=$1`
	rec := &repository.IdempotencyRecord{}
	var body sql.NullString
	var code sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, key).Scan(&rec.Key, &body, &code, &rec.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if body.Valid {
		rec.ResponseBody = []byte(body.String)
		rec.StatusCode = int(code.Int64)
	}
	return rec, nil
}

func (r *idempotencyRepository) Release(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM idempotency_keys WHERE key=$1`, key)
	return err
}
