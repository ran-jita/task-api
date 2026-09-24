package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	status := http.StatusInternalServerError
	message := "internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		status = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		}
	}

	if !c.Response().Committed {
		_ = c.JSON(status, ErrorResponse{
			Status:  status,
			Code:    http.StatusText(status),
			Message: message,
		})
	}
}
