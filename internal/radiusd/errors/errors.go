package errors

import (
	"errors"
	"fmt"

	"github.com/talkincode/toughradius/v9/internal/app"
)

// AuthError RADIUS认证错误
type AuthError struct {
	MetricsType string
	Message     string
}

func (e *AuthError) Error() string {
	return e.Message
}

// NewAuthError 创建认证错误
// 注意：不在这里记录metrics，由调用方决定
func NewAuthError(metricsType string, message string) error {
	return &AuthError{
		MetricsType: metricsType,
		Message:     message,
	}
}

// IsAuthError 判断是否为认证错误
func IsAuthError(err error) bool {
	_, ok := err.(*AuthError)
	return ok
}

// GetAuthError 获取认证错误详情
func GetAuthError(err error) (*AuthError, bool) {
	authErr, ok := err.(*AuthError)
	return authErr, ok
}

// Common auth error constructors for convenience

func NewUserNotExistsError() error {
	return NewAuthError(app.MetricsRadiusRejectNotExists, "user not exists")
}

func NewUserDisabledError() error {
	return NewAuthError(app.MetricsRadiusRejectDisable, "user status is disabled")
}

func NewUserExpiredError() error {
	return NewAuthError(app.MetricsRadiusRejectExpire, "user expired")
}

func NewPasswordMismatchError() error {
	return NewAuthError(app.MetricsRadiusRejectPasswdError, "password mismatch")
}

func NewOnlineLimitError(message string) error {
	return NewAuthError(app.MetricsRadiusRejectLimit, message)
}

func NewMacBindError() error {
	return NewAuthError(app.MetricsRadiusRejectBindError, "mac address binding failed")
}

func NewVlanBindError() error {
	return NewAuthError(app.MetricsRadiusRejectBindError, "vlan binding failed")
}

func NewUnauthorizedNasError(ip, identifier string, err error) error {
	return NewAuthError(app.MetricsRadiusRejectUnauthorized,
		fmt.Sprintf("unauthorized access to device, Ip=%s, Identifier=%s, %s",
			ip, identifier, err.Error()))
}

// WrapError 将普通error包装为AuthError
func WrapError(metricsType string, err error) error {
	if err == nil {
		return nil
	}
	if IsAuthError(err) {
		return err
	}
	return NewAuthError(metricsType, err.Error())
}

// NewError 创建普通错误（非AuthError）
func NewError(message string) error {
	return errors.New(message)
}
