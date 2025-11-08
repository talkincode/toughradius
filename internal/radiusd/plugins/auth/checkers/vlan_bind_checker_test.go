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

func TestVlanBindChecker_Name(t *testing.T) {
	checker := &VlanBindChecker{}
	assert.Equal(t, "vlan_bind", checker.Name())
}

func TestVlanBindChecker_Order(t *testing.T) {
	checker := &VlanBindChecker{}
	assert.Equal(t, 21, checker.Order())
}

func TestVlanBindChecker_Check(t *testing.T) {
	checker := &VlanBindChecker{}
	ctx := context.Background()

	tests := []struct {
		name        string
		bindVlan    int
		userVlan1   int
		userVlan2   int
		reqVlan1    int64
		reqVlan2    int64
		expectError bool
	}{
		{
			name:        "bind disabled",
			bindVlan:    0,
			userVlan1:   100,
			userVlan2:   200,
			reqVlan1:    999,
			reqVlan2:    888,
			expectError: false,
		},
		{
			name:        "vlan1 matches",
			bindVlan:    1,
			userVlan1:   100,
			userVlan2:   0,
			reqVlan1:    100,
			reqVlan2:    0,
			expectError: false,
		},
		{
			name:        "vlan1 mismatch",
			bindVlan:    1,
			userVlan1:   100,
			userVlan2:   0,
			reqVlan1:    200,
			reqVlan2:    0,
			expectError: true,
		},
		{
			name:        "vlan2 matches",
			bindVlan:    1,
			userVlan1:   0,
			userVlan2:   200,
			reqVlan1:    0,
			reqVlan2:    200,
			expectError: false,
		},
		{
			name:        "vlan2 mismatch",
			bindVlan:    1,
			userVlan1:   0,
			userVlan2:   200,
			reqVlan1:    0,
			reqVlan2:    300,
			expectError: true,
		},
		{
			name:        "both vlans match",
			bindVlan:    1,
			userVlan1:   100,
			userVlan2:   200,
			reqVlan1:    100,
			reqVlan2:    200,
			expectError: false,
		},
		{
			name:        "user vlan1 is 0, no check",
			bindVlan:    1,
			userVlan1:   0,
			userVlan2:   200,
			reqVlan1:    999,
			reqVlan2:    200,
			expectError: false,
		},
		{
			name:        "request vlan1 is 0, no check",
			bindVlan:    1,
			userVlan1:   100,
			userVlan2:   0,
			reqVlan1:    0,
			reqVlan2:    0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.RadiusUser{
				Username: "testuser",
				BindVlan: tt.bindVlan,
				Vlanid1:  tt.userVlan1,
				Vlanid2:  tt.userVlan2,
			}

			vendorReq := &vendorparsers.VendorRequest{
				Vlanid1: tt.reqVlan1,
				Vlanid2: tt.reqVlan2,
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
				assert.Contains(t, authErr.Message, "vlan")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVlanBindChecker_Check_NoVendorRequest(t *testing.T) {
	checker := &VlanBindChecker{}
	ctx := context.Background()

	user := &domain.RadiusUser{
		Username: "testuser",
		BindVlan: 1,
		Vlanid1:  100,
		Vlanid2:  200,
	}

	authCtx := &auth.AuthContext{
		User:          user,
		VendorRequest: nil,
	}

	err := checker.Check(ctx, authCtx)
	require.NoError(t, err)
}
