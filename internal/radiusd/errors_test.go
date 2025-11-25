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
			name:     "User not found error",
			errType:  app.MetricsRadiusRejectNotExists,
			errMsg:   "user not exists",
			expected: "user not exists",
		},
		{
			name:     "User disabled error",
			errType:  app.MetricsRadiusRejectDisable,
			errMsg:   "user status is disabled",
			expected: "user status is disabled",
		},
		{
			name:     "User expired error",
			errType:  app.MetricsRadiusRejectExpire,
			errMsg:   "user expire",
			expected: "user expire",
		},
		{
			name:     "Unauthorized access error",
			errType:  app.MetricsRadiusRejectUnauthorized,
			errMsg:   "unauthorized access",
			expected: "unauthorized access",
		},
		{
			name:     "Rate limit error",
			errType:  app.MetricsRadiusRejectLimit,
			errMsg:   "there is a authentication still in process",
			expected: "there is a authentication still in process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authErr := NewAuthError(tt.errType, tt.errMsg)

			// Test creating an AuthError
			if authErr == nil {
				t.Fatal("NewAuthError returned nil")
			}

			// Test Type field
			if authErr.Type != tt.errType {
				t.Errorf("expected Type = %s, got %s", tt.errType, authErr.Type)
			}

			// Test the Error() method
			if authErr.Error() != tt.expected {
				t.Errorf("expected Error() = %s, got %s", tt.expected, authErr.Error())
			}

			// Test the internal Err field
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

	// Validate that AuthError implements the error interface
	var _ error = authErr

	// Test whether it can be used as a standard error
	err := error(authErr)
	if err.Error() != "test error" {
		t.Errorf("expected error message = 'test error', got %s", err.Error())
	}
}

func TestAuthErrorComparison(t *testing.T) {
	err1 := NewAuthError(app.MetricsRadiusRejectNotExists, "user not found")
	err2 := NewAuthError(app.MetricsRadiusRejectNotExists, "user not found")
	err3 := NewAuthError(app.MetricsRadiusRejectDisable, "user disabled")

	// Test errors with the same type and message
	if err1.Type != err2.Type {
		t.Error("errors with same type should have same Type field")
	}

	if err1.Error() != err2.Error() {
		t.Error("errors with same message should return same Error()")
	}

	// Test errors with different types
	if err1.Type == err3.Type {
		t.Error("errors with different types should have different Type field")
	}
}

func TestAuthErrorIsError(t *testing.T) {
	authErr := NewAuthError(app.MetricsRadiusRejectNotExists, "test error")

	// Test comparisons using errors.Is
	if !errors.Is(authErr, authErr) {
		t.Error("AuthError should be equal to itself using errors.Is")
	}

	// Test comparisons with other errors
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
	// Test all known error type constants
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

	// Test error wrapping
	wrappedErr := errors.New("wrapped: " + authErr.Error())
	if wrappedErr.Error() != "wrapped: user not exists" {
		t.Errorf("error wrapping failed, got: %s", wrappedErr.Error())
	}
}
