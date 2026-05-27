package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/app"
)

func TestAuthError_Error(t *testing.T) {
	err := &AuthError{
		MetricsType: app.MetricsRadiusRejectNotExists,
		Message:     "test error message",
	}
	assert.Equal(t, "test error message", err.Error())
}

func TestNewAuthError(t *testing.T) {
	tests := []struct {
		name        string
		metricsType string
		message     string
	}{
		{
			name:        "user not exists error",
			metricsType: app.MetricsRadiusRejectNotExists,
			message:     "user not found",
		},
		{
			name:        "password error",
			metricsType: app.MetricsRadiusRejectPasswdError,
			message:     "invalid password",
		},
		{
			name:        "disabled error",
			metricsType: app.MetricsRadiusRejectDisable,
			message:     "account disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthError(tt.metricsType, tt.message)
			assert.NotNil(t, err)

			authErr, ok := err.(*AuthError)
			assert.True(t, ok)
			assert.Equal(t, tt.metricsType, authErr.MetricsType)
			assert.Equal(t, tt.message, authErr.Message)
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AuthError returns true",
			err:      NewAuthError(app.MetricsRadiusRejectNotExists, "test"),
			expected: true,
		},
		{
			name:     "standard error returns false",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAuthError(t *testing.T) {
	t.Run("returns AuthError when valid", func(t *testing.T) {
		originalErr := NewAuthError(app.MetricsRadiusRejectExpire, "expired")
		authErr, ok := GetAuthError(originalErr)
		assert.True(t, ok)
		assert.NotNil(t, authErr)
		assert.Equal(t, app.MetricsRadiusRejectExpire, authErr.MetricsType)
		assert.Equal(t, "expired", authErr.Message)
	})

	t.Run("returns false for standard error", func(t *testing.T) {
		authErr, ok := GetAuthError(errors.New("standard error"))
		assert.False(t, ok)
		assert.Nil(t, authErr)
	})

	t.Run("returns false for nil", func(t *testing.T) {
		authErr, ok := GetAuthError(nil)
		assert.False(t, ok)
		assert.Nil(t, authErr)
	})
}

func TestNewUserNotExistsError(t *testing.T) {
	err := NewUserNotExistsError()
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectNotExists, authErr.MetricsType)
	assert.Equal(t, "user not exists", authErr.Message)
}

func TestNewUserDisabledError(t *testing.T) {
	err := NewUserDisabledError()
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectDisable, authErr.MetricsType)
	assert.Equal(t, "user status is disabled", authErr.Message)
}

func TestNewUserExpiredError(t *testing.T) {
	err := NewUserExpiredError()
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectExpire, authErr.MetricsType)
	assert.Equal(t, "user expired", authErr.Message)
}

func TestNewPasswordMismatchError(t *testing.T) {
	err := NewPasswordMismatchError()
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectPasswdError, authErr.MetricsType)
	assert.Equal(t, "password mismatch", authErr.Message)
}

func TestNewOnlineLimitError(t *testing.T) {
	customMessage := "max sessions exceeded"
	err := NewOnlineLimitError(customMessage)
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectLimit, authErr.MetricsType)
	assert.Equal(t, customMessage, authErr.Message)
}

func TestNewMacBindError(t *testing.T) {
	err := NewMacBindError()
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectBindError, authErr.MetricsType)
	assert.Equal(t, "mac address binding failed", authErr.Message)
}

func TestNewVlanBindError(t *testing.T) {
	err := NewVlanBindError()
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectBindError, authErr.MetricsType)
	assert.Equal(t, "vlan binding failed", authErr.Message)
}

func TestNewUnauthorizedNasError(t *testing.T) {
	ip := "192.168.1.1"
	identifier := "nas-router-01"
	originalErr := errors.New("secret mismatch")

	err := NewUnauthorizedNasError(ip, identifier, originalErr)
	assert.NotNil(t, err)

	authErr, ok := GetAuthError(err)
	assert.True(t, ok)
	assert.Equal(t, app.MetricsRadiusRejectUnauthorized, authErr.MetricsType)
	assert.Contains(t, authErr.Message, ip)
	assert.Contains(t, authErr.Message, identifier)
	// The underlying cause should be wrapped
	assert.Equal(t, originalErr, authErr.Cause)
	// The full error message should include the cause
	assert.Contains(t, err.Error(), "secret mismatch")
}

func TestWrapError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := WrapError(app.MetricsRadiusRejectNotExists, nil)
		assert.Nil(t, result)
	})

	t.Run("AuthError passes through unchanged", func(t *testing.T) {
		originalErr := NewAuthError(app.MetricsRadiusRejectExpire, "original message")
		result := WrapError(app.MetricsRadiusRejectNotExists, originalErr)

		// Should return the same AuthError, not wrap it again
		assert.True(t, IsAuthError(result))
		authErr, _ := GetAuthError(result)
		assert.Equal(t, app.MetricsRadiusRejectExpire, authErr.MetricsType) // Original type preserved
		assert.Equal(t, "original message", authErr.Message)
	})

	t.Run("standard error gets wrapped", func(t *testing.T) {
		originalErr := errors.New("database connection failed")
		result := WrapError(app.MetricsRadiusRejectNotExists, originalErr)

		assert.True(t, IsAuthError(result))
		authErr, _ := GetAuthError(result)
		assert.Equal(t, app.MetricsRadiusRejectNotExists, authErr.MetricsType)
		assert.Equal(t, "database connection failed", authErr.Message)
	})
}

func TestNewError(t *testing.T) {
	message := "generic error occurred"
	err := NewError(message)

	assert.NotNil(t, err)
	assert.Equal(t, message, err.Error())
	// Should NOT be an AuthError
	assert.False(t, IsAuthError(err))
}

func TestAuthError_ErrorInterface(t *testing.T) {
	// Verify AuthError implements error interface
	var _ error = &AuthError{}
	_ = NewAuthError(app.MetricsRadiusRejectNotExists, "test")
}
