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
	"gorm.io/gorm"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

const tokenTTL = 12 * time.Hour

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerAuthRoutes() {
	webserver.ApiPOST("/auth/login", loginHandler)
	webserver.ApiGET("/auth/me", currentUserHandler)
}

func loginHandler(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_REQUEST", "无法解析登录参数", nil)
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return fail(c, http.StatusBadRequest, "INVALID_CREDENTIALS", "用户名与密码不能为空", nil)
	}

	var operator domain.SysOpr
	err := app.GDB().Where("username = ?", req.Username).First(&operator).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误", nil)
	}
	if err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "查询用户失败", err.Error())
	}

	hashed := common.Sha256HashWithSalt(req.Password, common.SecretSalt)
	if hashed != operator.Password {
		return fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "用户名或密码错误", nil)
	}
	if strings.EqualFold(operator.Status, common.DISABLED) {
		return fail(c, http.StatusForbidden, "ACCOUNT_DISABLED", "账号已被禁用", nil)
	}

	token, err := issueToken(operator)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "TOKEN_ERROR", "生成登录令牌失败", nil)
	}

	go func(id int64) {
		app.GDB().Model(&domain.SysOpr{}).Where("id = ?", id).Update("last_login", time.Now())
	}(operator.ID)

	operator.Password = ""
	return ok(c, map[string]interface{}{
		"token":        token,
		"user":         operator,
		"permissions":  []string{},
		"tokenExpires": time.Now().Add(tokenTTL).Unix(),
	})
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

func issueToken(op domain.SysOpr) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      strconv.FormatInt(op.ID, 10),
		"username": op.Username,
		"role":     op.Level,
		"exp":      now.Add(tokenTTL).Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Add(-1 * time.Minute).Unix(),
		"iss":      "toughradius",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(app.GConfig().Web.Secret))
}

func resolveOperatorFromContext(c echo.Context) (*domain.SysOpr, error) {
	userVal := c.Get("user")
	if userVal == nil {
		return nil, errors.New("no user in context")
	}

	token, ok := userVal.(*jwt.Token)
	if !ok {
		return nil, errors.New(fmt.Sprintf("invalid token type, got: %T", userVal))
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
	err = app.GDB().Where("id = ?", id).First(&operator).Error
	if err != nil {
		return nil, err
	}
	operator.Password = ""
	return &operator, nil
}
