package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ran-jita/task-api/internal/repository"
)

var ErrIdempotencyInProgress = errors.New("request with this idempotency key is still being processed")

type idempotencyUsecase struct {
	repo repository.IdempotencyRepository
}

func NewIdempotencyUsecase(repo repository.IdempotencyRepository) IdempotencyUsecase {
	return &idempotencyUsecase{repo: repo}
}

func (u *idempotencyUsecase) Execute(ctx context.Context, key string, fn IdempotentFunc) (interface{}, int, bool, error) {
	err := u.repo.Reserve(ctx, key)

	if err == nil {
		data, statusCode, fnErr := fn(ctx)
		if fnErr != nil {
			_ = u.repo.Release(ctx, key)
			return nil, 0, false, fnErr
		}

		body, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			return nil, 0, false, marshalErr
		}
		if err := u.repo.Complete(ctx, key, body, statusCode); err != nil {
			return nil, 0, false, err
		}
		return data, statusCode, false, nil
	}

	if errors.Is(err, repository.ErrDuplicateKey) {
		record, getErr := u.repo.Get(ctx, key)
		if getErr != nil {
			return nil, 0, false, getErr
		}

		if record != nil && time.Since(record.CreatedAt) > 24*time.Hour {
			if err := u.repo.Release(ctx, key); err != nil {
				return nil, 0, false, err
			}
			return u.Execute(ctx, key, fn)
		}

		if record == nil || record.ResponseBody == nil {
			return nil, 0, false, ErrIdempotencyInProgress
		}

		var data interface{}
		if err := json.Unmarshal(record.ResponseBody, &data); err != nil {
			return nil, 0, false, err
		}
		return data, record.StatusCode, true, nil
	}

	return nil, 0, false, err
}
