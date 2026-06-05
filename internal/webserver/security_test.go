package webserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

// TestSecurityHeadersApplied verifies the baseline security headers are set on
// responses.
func TestSecurityHeadersApplied(t *testing.T) {
	e := echo.New()
	e.Use(securityHeaders())
	e.GET("/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, "nosniff", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "1; mode=block", rec.Header().Get(echo.HeaderXXSSProtection))
}

// TestCorsAllowedOriginsParsing verifies env-driven CORS configuration parsing.
func TestCorsAllowedOriginsParsing(t *testing.T) {
	t.Setenv(corsEnvVar, "")
	assert.Nil(t, corsAllowedOrigins(), "unset/empty yields no origins (CORS disabled)")

	t.Setenv(corsEnvVar, " https://a.example.com , https://b.example.com ,")
	assert.Equal(t,
		[]string{"https://a.example.com", "https://b.example.com"},
		corsAllowedOrigins(),
		"origins should be split and trimmed, empties dropped")
}

// TestCorsHonorsConfiguredOrigins verifies that, when configured, a preflight
// from an allowed origin is granted and a disallowed origin is not.
func TestCorsHonorsConfiguredOrigins(t *testing.T) {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://allowed.example.com"},
		AllowCredentials: true,
	}))
	e.GET("/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	// Allowed origin echoed back.
	reqOK := httptest.NewRequest(http.MethodOptions, "/x", nil)
	reqOK.Header.Set(echo.HeaderOrigin, "https://allowed.example.com")
	reqOK.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	recOK := httptest.NewRecorder()
	e.ServeHTTP(recOK, reqOK)
	assert.Equal(t, "https://allowed.example.com", recOK.Header().Get(echo.HeaderAccessControlAllowOrigin))

	// Disallowed origin is not granted.
	reqBad := httptest.NewRequest(http.MethodOptions, "/x", nil)
	reqBad.Header.Set(echo.HeaderOrigin, "https://evil.example.com")
	reqBad.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	recBad := httptest.NewRecorder()
	e.ServeHTTP(recBad, reqBad)
	assert.NotEqual(t, "https://evil.example.com", recBad.Header().Get(echo.HeaderAccessControlAllowOrigin))
}
