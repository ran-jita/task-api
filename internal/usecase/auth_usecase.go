package usecase

import (
	"context"

	"github.com/ran-jita/task-api/internal/entity"
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token string
	User  entity.User
}

type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*entity.User, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
}
