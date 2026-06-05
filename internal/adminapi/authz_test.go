package adminapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/talkincode/toughradius/v9/internal/domain"
)

// authzNext is a sentinel handler used to detect whether the middleware allowed
// the request to proceed to the wrapped handler.
func authzNext(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func TestRequireLevel(t *testing.T) {
	tests := []struct {
		name           string
		operator       *domain.SysOpr
		allowedLevels  []string
		expectedStatus int
		expectNext     bool
	}{
		{
			name:           "admin allowed for requireAdmin",
			operator:       &domain.SysOpr{Level: LevelAdmin},
			allowedLevels:  []string{LevelSuper, LevelAdmin},
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "super allowed for requireAdmin",
			operator:       &domain.SysOpr{Level: LevelSuper},
			allowedLevels:  []string{LevelSuper, LevelAdmin},
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "operator denied for requireAdmin",
			operator:       &domain.SysOpr{Level: LevelOperator},
			allowedLevels:  []string{LevelSuper, LevelAdmin},
			expectedStatus: http.StatusForbidden,
			expectNext:     false,
		},
		{
			name:           "unknown level denied",
			operator:       &domain.SysOpr{Level: "guest"},
			allowedLevels:  []string{LevelSuper, LevelAdmin},
			expectedStatus: http.StatusForbidden,
			expectNext:     false,
		},
		{
			name:           "missing operator unauthorized",
			operator:       nil,
			allowedLevels:  []string{LevelSuper, LevelAdmin},
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/settings", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.operator != nil {
				c.Set("current_operator", tt.operator)
			}

			called := false
			handler := RequireLevel(tt.allowedLevels...)(func(c echo.Context) error {
				called = true
				return authzNext(c)
			})

			err := handler(c)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectNext, called)
		})
	}
}

func TestRequireAdminMatchesSuperAndAdmin(t *testing.T) {
	mw := requireAdmin()

	for _, level := range []string{LevelSuper, LevelAdmin} {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/network/nas", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("current_operator", &domain.SysOpr{Level: level})

		err := mw(authzNext)(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code, "level %s should be allowed", level)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/network/nas", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("current_operator", &domain.SysOpr{Level: LevelOperator})

	err := mw(authzNext)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code, "operator level should be denied")
}
