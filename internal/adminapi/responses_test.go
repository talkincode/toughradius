package adminapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestFailSanitizesInternalDetails verifies that server-side (>=500) failures
// never leak raw error details (such as database errors that expose table and
// column names) to the client. The generic code/message must still be present.
func TestFailSanitizesInternalDetails(t *testing.T) {
	e := echo.New()
	dbErr := errors.New(`pq: column "secret_token" of relation "rad_account" does not exist`)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query data", dbErr.Error())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	body := rec.Body.String()
	assert.NotContains(t, body, "secret_token", "raw column name must not leak to client")
	assert.NotContains(t, body, "rad_account", "raw table name must not leak to client")

	var resp ErrorResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "DATABASE_ERROR", resp.Error)
	assert.Equal(t, "Failed to query data", resp.Message)
	assert.Nil(t, resp.Details, "internal details must be stripped from the response")
}

// TestFailPreservesClientDetails verifies that client-side (4xx) failures keep
// their details, which are intended to help the caller correct the request.
func TestFailPreservesClientDetails(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input", "field 'name' is required")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "field 'name' is required"),
		"client-facing validation details should be preserved")
}
