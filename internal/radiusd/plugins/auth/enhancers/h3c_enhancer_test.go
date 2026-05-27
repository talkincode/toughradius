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
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/h3c"
	"layeh.com/radius"
)

func TestH3CAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewH3CAcceptEnhancer()
	assert.Equal(t, "accept-h3c", enhancer.Name())
}

func TestH3CAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewH3CAcceptEnhancer()
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

func TestH3CAcceptEnhancer_Enhance_VendorMatch(t *testing.T) {
	enhancer := NewH3CAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name          string
		vendorCode    string
		shouldEnhance bool
	}{
		{
			name:          "h3c vendor",
			vendorCode:    vendors.CodeH3C,
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
				upAvg := h3c.H3CInputAverageRate_Get(response)
				assert.Greater(t, uint32(upAvg), uint32(0))

				upPeak := h3c.H3CInputPeakRate_Get(response)
				assert.Greater(t, uint32(upPeak), uint32(0))

				downAvg := h3c.H3COutputAverageRate_Get(response)
				assert.Greater(t, uint32(downAvg), uint32(0))

				downPeak := h3c.H3COutputPeakRate_Get(response)
				assert.Greater(t, uint32(downPeak), uint32(0))
			}
		})
	}
}

func TestH3CAcceptEnhancer_Enhance_RateCalculation(t *testing.T) {
	enhancer := NewH3CAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name            string
		upRate          int
		downRate        int
		expectedUpAvg   uint32
		expectedDownAvg uint32
	}{
		{
			name:            "normal rates",
			upRate:          100,
			downRate:        200,
			expectedUpAvg:   100 * 1024,
			expectedDownAvg: 200 * 1024,
		},
		{
			name:            "zero rates",
			upRate:          0,
			downRate:        0,
			expectedUpAvg:   0,
			expectedDownAvg: 0,
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
				VendorCode: vendors.CodeH3C,
			}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			upAvg := h3c.H3CInputAverageRate_Get(response)
			assert.Equal(t, tt.expectedUpAvg, uint32(upAvg))

			downAvg := h3c.H3COutputAverageRate_Get(response)
			assert.Equal(t, tt.expectedDownAvg, uint32(downAvg))

			// Peak rate should be four times the average rate
			upPeak := h3c.H3CInputPeakRate_Get(response)
			expectedUpPeak := clampInt64(int64(tt.expectedUpAvg)*4, math.MaxInt32)
			assert.Equal(t, uint32(expectedUpPeak), uint32(upPeak)) //nolint:gosec // G115: test comparison

			downPeak := h3c.H3COutputPeakRate_Get(response)
			expectedDownPeak := clampInt64(int64(tt.expectedDownAvg)*4, math.MaxInt32)
			assert.Equal(t, uint32(expectedDownPeak), uint32(downPeak)) //nolint:gosec // G115: test comparison
		})
	}
}
