package usecase

import "context"

type IdempotentFunc func(ctx context.Context) (data interface{}, statusCode int, err error)

type IdempotencyUsecase interface {
	Execute(ctx context.Context, key string, fn IdempotentFunc) (data interface{}, statusCode int, replayed bool, err error)
}
