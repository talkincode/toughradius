package enhancers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/aruba"
	"layeh.com/radius"
)

func TestArubaAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewArubaAcceptEnhancer()
	assert.Equal(t, "accept-aruba", enhancer.Name())
}

func TestArubaAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewArubaAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name    string
		authCtx *auth.AuthContext
	}{
		{
			name:    "nil context",
			authCtx: nil,
		},
		{
			name: "nil response",
			authCtx: &auth.AuthContext{
				User: &domain.RadiusUser{},
			},
		},
		{
			name: "nil user",
			authCtx: &auth.AuthContext{
				Response: radius.New(radius.CodeAccessAccept, []byte("secret")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enhancer.Enhance(ctx, tt.authCtx)
			require.NoError(t, err)
		})
	}
}

func TestArubaAcceptEnhancer_Enhance_VendorMatch(t *testing.T) {
	enhancer := NewArubaAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name          string
		vendorCode    string
		shouldEnhance bool
	}{
		{
			name:          "aruba vendor",
			vendorCode:    vendors.CodeAruba,
			shouldEnhance: true,
		},
		{
			name:          "other vendor",
			vendorCode:    vendors.CodeHuawei,
			shouldEnhance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Username: "testuser",
				Vlanid1:  120,
				Domain:   "guest-role",
			}
			nas := &domain.NetNas{
				VendorCode: tt.vendorCode,
			}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			vlan := aruba.ArubaUserVlan_Get(response)
			role := aruba.ArubaUserRole_GetString(response)
			if tt.shouldEnhance {
				assert.Equal(t, uint32(120), uint32(vlan))
				assert.Equal(t, "guest-role", role)
			} else {
				assert.Equal(t, uint32(0), uint32(vlan))
				assert.Equal(t, "", role)
			}
		})
	}
}

func TestArubaAcceptEnhancer_Enhance_UserVlan(t *testing.T) {
	enhancer := NewArubaAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name         string
		vlanid1      int
		expectVlan   uint32
		expectAbsent bool // attribute must not be present at all
	}{
		{name: "valid access vlan", vlanid1: 200, expectVlan: 200},
		{name: "lowest valid vlan", vlanid1: 1, expectVlan: 1},
		{name: "highest valid vlan", vlanid1: arubaMaxVLANID, expectVlan: arubaMaxVLANID},
		{name: "zero vlan skipped", vlanid1: 0, expectAbsent: true},
		{name: "reserved 4095 skipped", vlanid1: 4095, expectAbsent: true},
		{name: "negative vlan skipped", vlanid1: -1, expectAbsent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Username: "testuser",
				Vlanid1:  tt.vlanid1,
			}
			nas := &domain.NetNas{VendorCode: vendors.CodeAruba}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			// Use Lookup so a skipped VLAN asserts true absence, not the zero
			// value that Get cannot distinguish from a present "0".
			vlan, lookupErr := aruba.ArubaUserVlan_Lookup(response)
			if tt.expectAbsent {
				assert.ErrorIs(t, lookupErr, radius.ErrNoAttribute)
			} else {
				require.NoError(t, lookupErr)
				assert.Equal(t, tt.expectVlan, uint32(vlan))
			}
		})
	}
}

func TestArubaAcceptEnhancer_Enhance_UserRole(t *testing.T) {
	enhancer := NewArubaAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name         string
		domain       string
		expectedRole string // "" means the attribute must be absent
	}{
		{name: "role from domain", domain: "employee", expectedRole: "employee"},
		{name: "empty domain skipped", domain: "", expectedRole: ""},
		{name: "NA domain skipped", domain: "N/A", expectedRole: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Username: "testuser",
				Domain:   tt.domain,
			}
			nas := &domain.NetNas{VendorCode: vendors.CodeAruba}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedRole, aruba.ArubaUserRole_GetString(response))
		})
	}
}
