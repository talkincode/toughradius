package checkers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/errors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
)

func TestMacBindChecker_Name(t *testing.T) {
	checker := &MacBindChecker{}
	assert.Equal(t, "mac_bind", checker.Name())
}

func TestMacBindChecker_Order(t *testing.T) {
	checker := &MacBindChecker{}
	assert.Equal(t, 20, checker.Order())
}

func TestMacBindChecker_Check(t *testing.T) {
	checker := &MacBindChecker{}
	ctx := context.Background()

	tests := []struct {
		name        string
		bindMac     int
		userMac     string
		requestMac  string
		expectError bool
	}{
		{
			name:        "bind disabled",
			bindMac:     0,
			userMac:     "00:11:22:33:44:55",
			requestMac:  "00:11:22:33:44:66",
			expectError: false,
		},
		{
			name:        "mac matches",
			bindMac:     1,
			userMac:     "00:11:22:33:44:55",
			requestMac:  "00:11:22:33:44:55",
			expectError: false,
		},
		{
			name:        "mac mismatch",
			bindMac:     1,
			userMac:     "00:11:22:33:44:55",
			requestMac:  "00:11:22:33:44:66",
			expectError: true,
		},
		{
			name:        "user mac is NA",
			bindMac:     1,
			userMac:     "N/A",
			requestMac:  "00:11:22:33:44:55",
			expectError: false,
		},
		{
			name:        "user mac is empty",
			bindMac:     1,
			userMac:     "",
			requestMac:  "00:11:22:33:44:55",
			expectError: false,
		},
		{
			name:        "request mac is empty",
			bindMac:     1,
			userMac:     "00:11:22:33:44:55",
			requestMac:  "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.RadiusUser{
				Username: "testuser",
				BindMac:  tt.bindMac,
				MacAddr:  tt.userMac,
			}

			vendorReq := &vendorparsers.VendorRequest{
				MacAddr: tt.requestMac,
			}

			authCtx := &auth.AuthContext{
				User:          user,
				VendorRequest: vendorReq,
			}

			err := checker.Check(ctx, authCtx)

			if tt.expectError {
				require.Error(t, err)
				authErr, ok := errors.GetAuthError(err)
				assert.True(t, ok)
				assert.Contains(t, authErr.Message, "mac")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMacBindChecker_Check_NoVendorRequest(t *testing.T) {
	checker := &MacBindChecker{}
	ctx := context.Background()

	user := &domain.RadiusUser{
		Username: "testuser",
		BindMac:  1,
		MacAddr:  "00:11:22:33:44:55",
	}

	// No VendorRequest
	authCtx := &auth.AuthContext{
		User:          user,
		VendorRequest: nil,
	}

	err := checker.Check(ctx, authCtx)
	require.NoError(t, err)
}

func TestMacBindChecker_Check_InvalidVendorRequest(t *testing.T) {
	checker := &MacBindChecker{}
	ctx := context.Background()

	user := &domain.RadiusUser{
		Username: "testuser",
		BindMac:  1,
		MacAddr:  "00:11:22:33:44:55",
	}

	// VendorRequest of an unexpected type
	authCtx := &auth.AuthContext{
		User:          user,
		VendorRequest: "invalid",
	}

	err := checker.Check(ctx, authCtx)
	require.NoError(t, err)
}
