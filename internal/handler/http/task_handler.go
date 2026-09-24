package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ran-jita/task-api/internal/usecase"
)

type TaskHandler struct {
	taskUsecase        usecase.TaskUsecase
	idempotencyUsecase usecase.IdempotencyUsecase
}

func NewTaskHandler(taskUsecase usecase.TaskUsecase, idempotencyUsecase usecase.IdempotencyUsecase) *TaskHandler {
	return &TaskHandler{taskUsecase: taskUsecase, idempotencyUsecase: idempotencyUsecase}
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *TaskHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req createTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Idempotency-Key header is required")
	}

	data, statusCode, replayed, err := h.idempotencyUsecase.Execute(
		c.Request().Context(),
		idempotencyKey,
		func(ctx context.Context) (interface{}, int, error) {
			task, err := h.taskUsecase.Create(ctx, usecase.CreateTaskInput{
				UserID:      userID,
				Title:       req.Title,
				Description: req.Description,
			})
			if err != nil {
				return nil, 0, err
			}
			return task, http.StatusCreated, nil
		},
	)
	if err != nil {
		return mapUsecaseError(err)
	}

	if replayed {
		c.Response().Header().Set("X-Idempotent-Replayed", "true")
	}
	return c.JSON(statusCode, SuccessResponse{Data: data})
}

type listTaskResponse struct {
	Tasks []taskDTO `json:"tasks"`
}

type taskDTO struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
}

func (h *TaskHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)

	limit := parseIntQuery(c, "limit", 10)
	page := parseIntQuery(c, "page", 1)

	result, err := h.taskUsecase.List(c.Request().Context(), usecase.ListTaskInput{
		UserID: userID,
		Status: c.QueryParam("status"),
		Title:  c.QueryParam("title"),
		Limit:  limit,
		Page:   page,
	})
	if err != nil {
		return mapUsecaseError(err)
	}

	meta := map[string]interface{}{
		"total": result.Total,
		"page":  page,
		"limit": limit,
	}
	return writeSuccessWithMeta(c, http.StatusOK, result.Tasks, meta)
}

func (h *TaskHandler) GetByID(c echo.Context) error {
	userID := c.Get("user_id").(string)
	taskID := c.Param("id")

	task, err := h.taskUsecase.GetByID(c.Request().Context(), taskID, userID)
	if err != nil {
		return mapUsecaseError(err)
	}
	return writeSuccess(c, http.StatusOK, task)
}

func (h *TaskHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(string)
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	task, err := h.taskUsecase.Update(c.Request().Context(), usecase.UpdateTaskInput{
		ID:          c.Param("id"),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		return mapUsecaseError(err)
	}
	return writeSuccess(c, http.StatusOK, task)
}

func (h *TaskHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(string)
	if err := h.taskUsecase.Delete(c.Request().Context(), c.Param("id"), userID); err != nil {
		return mapUsecaseError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func parseIntQuery(c echo.Context, key string, fallback int) int {
	val := c.QueryParam(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
