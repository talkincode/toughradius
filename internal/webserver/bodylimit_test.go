package webserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultBodyLimitIsValid ensures the configured limit string is accepted by
// the BodyLimit middleware (it panics on an invalid value).
func TestDefaultBodyLimitIsValid(t *testing.T) {
	require.NotPanics(t, func() {
		_ = middleware.BodyLimit(DefaultBodyLimit)
	})
}

// TestBodyLimitRejectsOversizedBody verifies that BodyLimit returns 413 for a
// request body larger than the configured limit and allows smaller ones.
func TestBodyLimitRejectsOversizedBody(t *testing.T) {
	e := echo.New()
	e.Use(middleware.BodyLimit("1K"))
	e.POST("/x", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// Body under the limit is accepted.
	small := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(make([]byte, 512)))
	recSmall := httptest.NewRecorder()
	e.ServeHTTP(recSmall, small)
	assert.Equal(t, http.StatusOK, recSmall.Code)

	// Body over the limit is rejected.
	big := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(make([]byte, 2048)))
	recBig := httptest.NewRecorder()
	e.ServeHTTP(recBig, big)
	assert.Equal(t, http.StatusRequestEntityTooLarge, recBig.Code)
}
