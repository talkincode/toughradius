package enhancers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/cisco"
	"layeh.com/radius"
)

func TestCiscoAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewCiscoAcceptEnhancer()
	assert.Equal(t, "accept-cisco", enhancer.Name())
}

func TestCiscoAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewCiscoAcceptEnhancer()
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

func TestCiscoAcceptEnhancer_Enhance_VendorMatch(t *testing.T) {
	enhancer := NewCiscoAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name          string
		vendorCode    string
		shouldEnhance bool
	}{
		{
			name:          "cisco vendor",
			vendorCode:    vendors.CodeCisco,
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
				AddrPool: "pool-a",
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

			avpair := cisco.CiscoAVPair_GetString(response)
			if tt.shouldEnhance {
				assert.Equal(t, "ip:addr-pool=pool-a", avpair)
			} else {
				assert.Equal(t, "", avpair)
			}
		})
	}
}

func TestCiscoAcceptEnhancer_Enhance_AddrPool(t *testing.T) {
	enhancer := NewCiscoAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name         string
		addrPool     string
		expectAVPair string // "" means the attribute must be absent
	}{
		{name: "named pool", addrPool: "subscribers", expectAVPair: "ip:addr-pool=subscribers"},
		{name: "empty pool skipped", addrPool: "", expectAVPair: ""},
		{name: "NA pool skipped", addrPool: "N/A", expectAVPair: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Username: "testuser",
				AddrPool: tt.addrPool,
			}
			nas := &domain.NetNas{VendorCode: vendors.CodeCisco}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			// Use Lookup so a skipped pool asserts true absence rather than the
			// empty value that Get cannot distinguish from a present "".
			value, lookupErr := cisco.CiscoAVPair_LookupString(response)
			if tt.expectAVPair == "" {
				assert.ErrorIs(t, lookupErr, radius.ErrNoAttribute)
			} else {
				require.NoError(t, lookupErr)
				assert.Equal(t, tt.expectAVPair, value)
			}
		})
	}
}
