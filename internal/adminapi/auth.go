package adminapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

const tokenTTL = 12 * time.Hour

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginRateLimiter throttles authentication attempts per client IP to slow down
// brute-force / credential-stuffing attacks. It allows a short burst of attempts
// and then refills slowly; exhausting the budget yields HTTP 429.
func loginRateLimiter() echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      rate.Every(3 * time.Second),
		Burst:     5,
		ExpiresIn: 3 * time.Minute,
	})
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return fail(c, http.StatusForbidden, "RATE_LIMIT_ERROR", "Unable to evaluate rate limit", nil)
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return fail(c, http.StatusTooManyRequests, "RATE_LIMITED",
				"Too many login attempts, please try again later", nil)
		},
	})
}

func registerAuthRoutes() {
	webserver.ApiPOST("/auth/login", loginHandler, loginRateLimiter())
	webserver.ApiGET("/auth/me", currentUserHandler)
}

func loginHandler(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Unable to parse login parameters", nil)
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return fail(c, http.StatusBadRequest, "INVALID_CREDENTIALS", "Username and password cannot be empty", nil)
	}

	var operator domain.SysOpr
	err := GetDB(c).Where("username = ?", req.Username).First(&operator).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect username or password", nil)
	}
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query user", err.Error())
	}

	if !common.VerifyPassword(req.Password, operator.Password) {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect username or password", nil)
	}
	if strings.EqualFold(operator.Status, common.DISABLED) {
		return fail(c, http.StatusForbidden, "ACCOUNT_DISABLED", "Account has been disabled", nil)
	}

	token, err := issueToken(c, operator)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate login token", nil)
	}

	go func(id int64) {
		GetDB(c).Model(&domain.SysOpr{}).Where("id = ?", id).Update("last_login", time.Now())
	}(operator.ID)

	operator.Password = ""
	return ok(c, map[string]interface{}{
		"token":        token,
		"user":         operator,
		"permissions":  []string{},
		"tokenExpires": time.Now().Add(tokenTTL).Unix(),
	})
}

func issueToken(c echo.Context, op domain.SysOpr) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", op.ID),
		"username": op.Username,
		"role":     op.Level,
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Add(-1 * time.Minute).Unix(),
		"iss":      "toughradius",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetAppContext(c).Config().Web.Secret))
}

func currentUserHandler(c echo.Context) error {
	operator, err := resolveOperatorFromContext(c)
	if err != nil {
		return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error(), nil)
	}
	return ok(c, map[string]interface{}{
		"user":        operator,
		"permissions": []string{},
	})
}

// testOperatorResolver is a test-only seam. It is nil in production builds and
// is assigned only by test code (see test_helpers_test.go), letting unit tests
// inject an already-resolved operator without standing up the full JWT
// middleware. Production code never assigns it, so the shipped binary always
// resolves caller identity from the signed JWT below and never trusts an
// operator placed directly into the request context.
var testOperatorResolver func(c echo.Context) (*domain.SysOpr, bool)

func resolveOperatorFromContext(c echo.Context) (*domain.SysOpr, error) {
	if testOperatorResolver != nil {
		if op, ok := testOperatorResolver(c); ok {
			return op, nil
		}
	}

	userVal := c.Get("user")
	if userVal == nil {
		return nil, errors.New("no user in context")
	}

	token, ok := userVal.(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("invalid token type, got: %T", userVal)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, errors.New("invalid token subject")
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return nil, errors.New("invalid token id")
	}
	var operator domain.SysOpr
	err = GetDB(c).Where("id = ?", id).First(&operator).Error
	if err != nil {
		return nil, err
	}
	// Reject tokens issued to accounts that have since been disabled so that
	// revoking an operator takes effect immediately, before the JWT expires.
	if strings.EqualFold(operator.Status, common.DISABLED) {
		return nil, errors.New("account disabled")
	}
	operator.Password = ""
	return &operator, nil
}
