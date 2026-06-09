package adminapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Meta carries pagination information for a list response: the total number of
// matching rows across all pages, the 1-based page index, and the page size that
// were applied to the query.
type Meta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// Response is the unified success envelope returned by admin API handlers. Data
// holds the payload (an object or a slice) and is omitted when nil; Meta is
// present only for paginated list responses. Handlers populate it through the
// internal ok and paged helpers rather than constructing it directly.
type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta *Meta       `json:"meta,omitempty"`
}

// ErrorResponse is the unified failure envelope returned by admin API handlers.
// Error is a stable, machine-readable code (for example "NOT_FOUND" or
// "VALIDATION_ERROR") that clients may branch on; Message is a human-readable
// explanation; Details carries optional structured context and is omitted for
// server-side (5xx) failures so internal information is never leaked to clients.
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
	// Never expose internal error details (e.g. raw database errors that contain
	// table/column names) to clients on server-side failures. Log them for
	// operators and return only the generic code and message.
	if status >= http.StatusInternalServerError && details != nil {
		zap.L().Error("adminapi internal error",
			zap.String("namespace", "adminapi"),
			zap.Int("status", status),
			zap.String("code", code),
			zap.String("path", c.Path()),
			zap.Any("details", details),
		)
		details = nil
	}
	return c.JSON(status, ErrorResponse{
		Error:   code,
		Message: message,
		Details: details,
	})
}

// handleValidationError normalizes validator responses into the unified error payload
func handleValidationError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	if he, ok := err.(*echo.HTTPError); ok {
		if payload, ok := he.Message.(map[string]interface{}); ok {
			errCode := "VALIDATION_ERROR"
			if code, ok := payload["error"].(string); ok && code != "" {
				errCode = code
			}
			message := "Request parameter validation failed"
			if msg, ok := payload["message"].(string); ok && msg != "" {
				message = msg
			}
			return fail(c, he.Code, errCode, message, payload["details"])
		}
		return fail(c, he.Code, "VALIDATION_ERROR", fmt.Sprint(he.Message), nil)
	}

	return fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
}
