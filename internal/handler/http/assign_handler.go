package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ran-jita/task-api/internal/usecase"
)

type AssignHandler struct {
	assignUsecase usecase.AssignUsecase
}

func NewAssignHandler(assignUsecase usecase.AssignUsecase) *AssignHandler {
	return &AssignHandler{assignUsecase: assignUsecase}
}

type assignRequest struct {
	AssigneeID string `json:"assignee_id"`
}

func (h *AssignHandler) Assign(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req assignRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.AssigneeID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "assignee_id is required")
	}

	err := h.assignUsecase.Assign(c.Request().Context(), usecase.AssignTaskInput{
		TaskID:      c.Param("id"),
		AssigneeID:  req.AssigneeID,
		RequestedBy: userID,
	})
	if err != nil {
		return mapUsecaseError(err)
	}
	return c.NoContent(http.StatusNoContent)
}
