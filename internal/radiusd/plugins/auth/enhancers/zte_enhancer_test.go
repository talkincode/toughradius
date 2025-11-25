package enhancers

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/zte"
	"layeh.com/radius"
)

func TestZTEAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewZTEAcceptEnhancer()
	assert.Equal(t, "accept-zte", enhancer.Name())
}

func TestZTEAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewZTEAcceptEnhancer()
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

func TestZTEAcceptEnhancer_Enhance_VendorMatch(t *testing.T) {
	enhancer := NewZTEAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name          string
		vendorCode    string
		shouldEnhance bool
	}{
		{
			name:          "zte vendor",
			vendorCode:    vendors.CodeZTE,
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
				upRate := zte.ZTERateCtrlSCRUp_Get(response)
				assert.Greater(t, uint32(upRate), uint32(0))

				downRate := zte.ZTERateCtrlSCRDown_Get(response)
				assert.Greater(t, uint32(downRate), uint32(0))
			}
		})
	}
}

func TestZTEAcceptEnhancer_Enhance_RateCalculation(t *testing.T) {
	enhancer := NewZTEAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name         string
		upRate       int
		downRate     int
		expectedUp   uint32
		expectedDown uint32
	}{
		{
			name:         "normal rates",
			upRate:       100,
			downRate:     200,
			expectedUp:   100 * 1024,
			expectedDown: 200 * 1024,
		},
		{
			name:         "zero rates",
			upRate:       0,
			downRate:     0,
			expectedUp:   0,
			expectedDown: 0,
		},
		{
			name:         "max boundary",
			upRate:       2097152,
			downRate:     2097152,
			expectedUp:   math.MaxInt32,
			expectedDown: math.MaxInt32,
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
				VendorCode: vendors.CodeZTE,
			}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			upRate := zte.ZTERateCtrlSCRUp_Get(response)
			assert.Equal(t, tt.expectedUp, uint32(upRate))

			downRate := zte.ZTERateCtrlSCRDown_Get(response)
			assert.Equal(t, tt.expectedDown, uint32(downRate))
		})
	}
}
