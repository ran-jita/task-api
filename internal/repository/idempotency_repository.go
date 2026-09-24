package repository

import (
	"context"
	"errors"
	"time"
)

var ErrDuplicateKey = errors.New("idempotency key already exists")

type IdempotencyRecord struct {
	Key          string
	ResponseBody []byte
	StatusCode   int
	CreatedAt    time.Time
}

type IdempotencyRepository interface {
	Reserve(ctx context.Context, key string) error
	Complete(ctx context.Context, key string, responseBody []byte, statusCode int) error
	Get(ctx context.Context, key string) (*IdempotencyRecord, error)
	Release(ctx context.Context, key string) error
}
