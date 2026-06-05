package adminapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Operator privilege levels, ordered from most to least privileged.
const (
	LevelSuper    = "super"
	LevelAdmin    = "admin"
	LevelOperator = "operator"
)

// RequireLevel returns middleware that authorizes a request only when the
// authenticated operator's level is one of the allowed levels.
//
// It must be chained after the JWT middleware so the operator can be resolved
// from the request context. Requests from operators whose level is not allowed
// receive HTTP 403; requests without a resolvable operator receive HTTP 401.
func RequireLevel(levels ...string) echo.MiddlewareFunc {
	allowed := make(map[string]bool, len(levels))
	for _, l := range levels {
		allowed[l] = true
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			operator, err := resolveOperatorFromContext(c)
			if err != nil {
				return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
			}
			if !allowed[operator.Level] {
				return fail(c, http.StatusForbidden, "PERMISSION_DENIED",
					"Insufficient privileges for this operation", nil)
			}
			return next(c)
		}
	}
}

// requireAdmin restricts an endpoint to admin and super operators. It is used to
// guard configuration writes and system backup/restore, which are not safe for
// regular operator-level accounts.
func requireAdmin() echo.MiddlewareFunc {
	return RequireLevel(LevelSuper, LevelAdmin)
}
