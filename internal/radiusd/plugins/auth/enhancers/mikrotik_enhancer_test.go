package enhancers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/mikrotik"
	"layeh.com/radius"
)

func TestMikrotikAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewMikrotikAcceptEnhancer()
	assert.Equal(t, "accept-mikrotik", enhancer.Name())
}

func TestMikrotikAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewMikrotikAcceptEnhancer()
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

func TestMikrotikAcceptEnhancer_Enhance_VendorMatch(t *testing.T) {
	enhancer := NewMikrotikAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name          string
		vendorCode    string
		shouldEnhance bool
	}{
		{
			name:          "mikrotik vendor",
			vendorCode:    vendors.CodeMikrotik,
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
				UpRate:   1024,
				DownRate: 2048,
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

			if tt.shouldEnhance {
				rateLimit := mikrotik.MikrotikRateLimit_GetString(response)
				assert.NotEmpty(t, rateLimit)
				// Format should be "UpRatek/DownRatek"
				assert.Equal(t, "1024k/2048k", rateLimit)
			}
		})
	}
}

func TestMikrotikAcceptEnhancer_Enhance_RateFormat(t *testing.T) {
	enhancer := NewMikrotikAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name           string
		upRate         int
		downRate       int
		expectedFormat string
	}{
		{
			name:           "normal rates",
			upRate:         100,
			downRate:       200,
			expectedFormat: "100k/200k",
		},
		{
			name:           "zero rates",
			upRate:         0,
			downRate:       0,
			expectedFormat: "0k/0k",
		},
		{
			name:           "asymmetric rates",
			upRate:         512,
			downRate:       2048,
			expectedFormat: "512k/2048k",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Username: "testuser",
				UpRate:   tt.upRate,
				DownRate: tt.downRate,
			}
			nas := &domain.NetNas{
				VendorCode: vendors.CodeMikrotik,
			}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			rateLimit := mikrotik.MikrotikRateLimit_GetString(response)
			assert.Equal(t, tt.expectedFormat, rateLimit)
		})
	}
}
