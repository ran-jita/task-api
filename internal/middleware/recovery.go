package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
)

type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf(`{"level":"ERROR","request_id":"%s","panic":"%v","stack":%q}`,
						getRequestID(c), r, string(debug.Stack()))

					_ = c.JSON(http.StatusInternalServerError, errorResponse{
						Status:  http.StatusInternalServerError,
						Code:    "INTERNAL_ERROR",
						Message: "internal server error",
					})
				}
			}()
			return next(c)
		}
	}
}
