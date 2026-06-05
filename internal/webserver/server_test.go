package webserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestJwtSkipFuncDoesNotBypassWithDevmode ensures the JWT skipper only skips
// the explicitly public routes and never disables auth globally, including when
// the removed TOUGHRADIUS_DEVMODE environment variable is set.
func TestJwtSkipFuncDoesNotBypassWithDevmode(t *testing.T) {
	e := echo.New()
	skip := jwtSkipFunc()

	newCtx := func(path string) echo.Context {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(path)
		return c
	}

	// Public routes are skipped.
	assert.True(t, skip(newCtx("/ready")))
	assert.True(t, skip(newCtx(apiBasePath+"/auth/login")))

	// Protected routes are never skipped.
	assert.False(t, skip(newCtx(apiBasePath+"/users")))

	// Setting the former bypass env var must not disable auth.
	t.Setenv("TOUGHRADIUS_DEVMODE", "true")
	assert.False(t, skip(newCtx(apiBasePath+"/users")))
}
