package validators

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func TestPAPValidator_Name(t *testing.T) {
	validator := &PAPValidator{}
	assert.Equal(t, "pap", validator.Name())
}

func TestPAPValidator_CanHandle(t *testing.T) {
	validator := &PAPValidator{}

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "with password",
			password: "testpass123",
			expected: true,
		},
		{
			name:     "empty password",
			password: "",
			expected: false,
		},
		{
			name:     "whitespace only",
			password: "   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			if tt.password != "" {
				rfc2865.UserPassword_SetString(packet, tt.password)
			}

			req := &radius.Request{Packet: packet}
			authCtx := &auth.AuthContext{Request: req}

			result := validator.CanHandle(authCtx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPAPValidator_Validate(t *testing.T) {
	validator := &PAPValidator{}
	ctx := context.Background()

	tests := []struct {
		name            string
		requestPassword string
		userPassword    string
		expectError     bool
	}{
		{
			name:            "correct password",
			requestPassword: "testpass123",
			userPassword:    "testpass123",
			expectError:     false,
		},
		{
			name:            "wrong password",
			requestPassword: "wrongpass",
			userPassword:    "testpass123",
			expectError:     true,
		},
		{
			name:            "empty request password",
			requestPassword: "",
			userPassword:    "testpass123",
			expectError:     true,
		},
		{
			name:            "password with spaces",
			requestPassword: "test pass",
			userPassword:    "test pass",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
			rfc2865.UserPassword_SetString(packet, tt.requestPassword)

			req := &radius.Request{Packet: packet}
			user := &domain.RadiusUser{Username: "testuser", Password: tt.userPassword}
			authCtx := &auth.AuthContext{
				Request: req,
				User:    user,
			}

			err := validator.Validate(ctx, authCtx, tt.userPassword)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
