package radiusd

import (
	"errors"
	"testing"

	"github.com/talkincode/toughradius/v9/internal/app"
)

func TestNewAuthError(t *testing.T) {
	tests := []struct {
		name     string
		errType  string
		errMsg   string
		expected string
	}{
		{
			name:     "用户不存在错误",
			errType:  app.MetricsRadiusRejectNotExists,
			errMsg:   "user not exists",
			expected: "user not exists",
		},
		{
			name:     "用户已禁用错误",
			errType:  app.MetricsRadiusRejectDisable,
			errMsg:   "user status is disabled",
			expected: "user status is disabled",
		},
		{
			name:     "用户过期错误",
			errType:  app.MetricsRadiusRejectExpire,
			errMsg:   "user expire",
			expected: "user expire",
		},
		{
			name:     "未授权访问错误",
			errType:  app.MetricsRadiusRejectUnauthorized,
			errMsg:   "unauthorized access",
			expected: "unauthorized access",
		},
		{
			name:     "速率限制错误",
			errType:  app.MetricsRadiusRejectLimit,
			errMsg:   "there is a authentication still in process",
			expected: "there is a authentication still in process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authErr := NewAuthError(tt.errType, tt.errMsg)

			// 测试 AuthError 的创建
			if authErr == nil {
				t.Fatal("NewAuthError returned nil")
			}

			// 测试 Type 字段
			if authErr.Type != tt.errType {
				t.Errorf("expected Type = %s, got %s", tt.errType, authErr.Type)
			}

			// 测试 Error() 方法
			if authErr.Error() != tt.expected {
				t.Errorf("expected Error() = %s, got %s", tt.expected, authErr.Error())
			}

			// 测试内部 Err 字段
			if authErr.Err == nil {
				t.Error("Err field should not be nil")
			}

			if authErr.Err.Error() != tt.expected {
				t.Errorf("expected internal error = %s, got %s", tt.expected, authErr.Err.Error())
			}
		})
	}
}

func TestAuthErrorImplementsError(t *testing.T) {
	authErr := NewAuthError(app.MetricsRadiusRejectNotExists, "test error")

	// 验证 AuthError 实现了 error 接口
	var _ error = authErr

	// 测试是否可以作为普通 error 使用
	err := error(authErr)
	if err.Error() != "test error" {
		t.Errorf("expected error message = 'test error', got %s", err.Error())
	}
}

func TestAuthErrorComparison(t *testing.T) {
	err1 := NewAuthError(app.MetricsRadiusRejectNotExists, "user not found")
	err2 := NewAuthError(app.MetricsRadiusRejectNotExists, "user not found")
	err3 := NewAuthError(app.MetricsRadiusRejectDisable, "user disabled")

	// 测试相同类型和消息的错误
	if err1.Type != err2.Type {
		t.Error("errors with same type should have same Type field")
	}

	if err1.Error() != err2.Error() {
		t.Error("errors with same message should return same Error()")
	}

	// 测试不同类型的错误
	if err1.Type == err3.Type {
		t.Error("errors with different types should have different Type field")
	}
}

func TestAuthErrorIsError(t *testing.T) {
	authErr := NewAuthError(app.MetricsRadiusRejectNotExists, "test error")

	// 测试使用 errors.Is 比较
	if !errors.Is(authErr, authErr) {
		t.Error("AuthError should be equal to itself using errors.Is")
	}

	// 测试与其他错误比较
	otherErr := errors.New("test error")
	if errors.Is(authErr, otherErr) {
		t.Error("AuthError should not be equal to a different error")
	}
}

func TestAuthErrorEmptyMessage(t *testing.T) {
	authErr := NewAuthError(app.MetricsRadiusRejectNotExists, "")

	if authErr.Error() != "" {
		t.Errorf("expected empty error message, got %s", authErr.Error())
	}
}

func TestAuthErrorType(t *testing.T) {
	// 测试所有已知的错误类型常量
	errorTypes := []string{
		app.MetricsRadiusRejectNotExists,
		app.MetricsRadiusRejectDisable,
		app.MetricsRadiusRejectExpire,
		app.MetricsRadiusRejectUnauthorized,
		app.MetricsRadiusRejectLimit,
		app.MetricsRadiusRejectPasswdError,
		app.MetricsRadiusRejectBindError,
		app.MetricsRadiusRejectLdapError,
		app.MetricsRadiusRejectOther,
	}

	for _, errType := range errorTypes {
		authErr := NewAuthError(errType, "test error")
		if authErr.Type != errType {
			t.Errorf("expected Type = %s, got %s", errType, authErr.Type)
		}
	}
}

func TestAuthErrorWrapping(t *testing.T) {
	authErr := NewAuthError(app.MetricsRadiusRejectNotExists, "user not exists")

	// 测试错误包装
	wrappedErr := errors.New("wrapped: " + authErr.Error())
	if wrappedErr.Error() != "wrapped: user not exists" {
		t.Errorf("error wrapping failed, got: %s", wrappedErr.Error())
	}
}
