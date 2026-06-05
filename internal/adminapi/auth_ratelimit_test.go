package adminapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestLoginRateLimiter verifies that repeated login attempts from the same client
// IP are throttled with HTTP 429 once the burst budget is exhausted.
func TestLoginRateLimiter(t *testing.T) {
	e := echo.New()
	e.POST("/login", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, loginRateLimiter())

	const clientIP = "203.0.113.7:40000"

	// The burst budget (5) allows the first attempts through.
	allowed := 0
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = clientIP
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			allowed++
		}
	}
	assert.Equal(t, 5, allowed, "burst attempts should be allowed")

	// The next immediate attempt is denied.
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = clientIP
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "exceeding the budget should be rate limited")

	// A different client IP is unaffected.
	req2 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req2.RemoteAddr = "198.51.100.9:40000"
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code, "a different IP should not be throttled")
}
