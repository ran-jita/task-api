package usecase

import "context"

type AssignTaskInput struct {
	TaskID      string
	AssigneeID  string
	RequestedBy string
}

type AssignUsecase interface {
	Assign(ctx context.Context, input AssignTaskInput) error
}
