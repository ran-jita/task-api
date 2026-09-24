package usecase

import "errors"

var (
	ErrEmailAlreadyUsed  = errors.New("email already used")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrTaskNotFound      = errors.New("task not found")
	ErrForbidden         = errors.New("forbidden: not the task owner")
	ErrAssigneeNotFound  = errors.New("assignee not found")
)
