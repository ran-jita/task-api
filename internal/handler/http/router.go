package http

import (
	"github.com/labstack/echo/v4"

	"github.com/ran-jita/task-api/internal/middleware"
)

func NewRouter(
	authHandler *AuthHandler,
	taskHandler *TaskHandler,
	assignHandler *AssignHandler,
	jwtSecret []byte,
) *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = CustomHTTPErrorHandler

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.Logging())

	// Public routes
	e.POST("/auth/register", authHandler.Register)
	e.POST("/auth/login", authHandler.Login)

	// Protected routes
	tasks := e.Group("/tasks", middleware.JWTAuth(jwtSecret))
	tasks.POST("", taskHandler.Create)
	tasks.GET("", taskHandler.List)
	tasks.GET("/:id", taskHandler.GetByID)
	tasks.PUT("/:id", taskHandler.Update)
	tasks.DELETE("/:id", taskHandler.Delete)
	tasks.POST("/:id/assign", assignHandler.Assign)

	return e
}
