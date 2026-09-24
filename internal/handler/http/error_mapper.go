package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ran-jita/task-api/internal/usecase"
)

func mapUsecaseError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrTaskNotFound), errors.Is(err, usecase.ErrAssigneeNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, usecase.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, usecase.ErrEmailAlreadyUsed):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, usecase.ErrInvalidCredential):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, usecase.ErrIdempotencyInProgress):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
