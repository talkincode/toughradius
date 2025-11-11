package adminapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Meta describes pagination information
type Meta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// Response represents the unified response structure
type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta *Meta       `json:"meta,omitempty"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func ok(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{Data: data})
}

func paged(c echo.Context, data interface{}, total int64, page, pageSize int) error {
	return c.JSON(http.StatusOK, Response{
		Data: data,
		Meta: &Meta{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

func fail(c echo.Context, status int, code, message string, details interface{}) error {
	if status == 0 {
		status = http.StatusBadRequest
	}
	return c.JSON(status, ErrorResponse{
		Error:   code,
		Message: message,
		Details: details,
	})
}
