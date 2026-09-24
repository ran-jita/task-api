package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

type logEntry struct {
	RequestID string `json:"request_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Level     string `json:"level"`
	UserID    string `json:"user_id,omitempty"`
}

func Logging() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			status := c.Response().Status

			if he, ok := err.(*echo.HTTPError); ok {
				status = he.Code
			} else if err != nil && status < 400 {
				status = 500
			}

			level := "INFO"
			switch {
			case status >= 500:
				level = "ERROR"
			case status >= 400:
				level = "WARN"
			}

			userID, _ := c.Get("user_id").(string)

			entry := logEntry{
				RequestID: getRequestID(c),
				Method:    c.Request().Method,
				Path:      c.Path(),
				Status:    status,
				LatencyMs: time.Since(start).Milliseconds(),
				Level:     level,
				UserID:    userID,
			}
			line, _ := json.Marshal(entry)
			log.Println(string(line))

			return err
		}
	}
}
