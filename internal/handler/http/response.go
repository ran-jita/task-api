package http

import "github.com/labstack/echo/v4"

type ErrorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

func writeSuccess(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, SuccessResponse{Data: data})
}

func writeSuccessWithMeta(c echo.Context, status int, data, meta interface{}) error {
	return c.JSON(status, SuccessResponse{Data: data, Meta: meta})
}
