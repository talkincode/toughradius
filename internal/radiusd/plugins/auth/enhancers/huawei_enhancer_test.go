package enhancers

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
	"layeh.com/radius"
)

func TestHuaweiAcceptEnhancer_Name(t *testing.T) {
	enhancer := NewHuaweiAcceptEnhancer()
	assert.Equal(t, "accept-huawei", enhancer.Name())
}

func TestHuaweiAcceptEnhancer_Enhance_NilSafety(t *testing.T) {
	enhancer := NewHuaweiAcceptEnhancer()
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

func TestHuaweiAcceptEnhancer_Enhance_VendorMatch(t *testing.T) {
	enhancer := NewHuaweiAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name        string
		vendorCode  string
		shouldEnhance bool
	}{
		{
			name:        "huawei vendor",
			vendorCode:  vendorHuawei,
			shouldEnhance: true,
		},
		{
			name:        "other vendor",
			vendorCode:  "9999",
			shouldEnhance: false,
		},
		{
			name:        "empty vendor",
			vendorCode:  "",
			shouldEnhance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := radius.New(radius.CodeAccessAccept, []byte("secret"))
			user := &domain.RadiusUser{
				Username: "testuser",
				UpRate:   1024,  // 1024 KB/s
				DownRate: 2048,  // 2048 KB/s
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

			// 检查是否添加了 Huawei 属性
			if tt.shouldEnhance {
				// 验证速率属性已设置
				upAvg := huawei.HuaweiInputAverageRate_Get(response)
				assert.Greater(t, uint32(upAvg), uint32(0))

				upPeak := huawei.HuaweiInputPeakRate_Get(response)
				assert.Greater(t, uint32(upPeak), uint32(0))

				downAvg := huawei.HuaweiOutputAverageRate_Get(response)
				assert.Greater(t, uint32(downAvg), uint32(0))

				downPeak := huawei.HuaweiOutputPeakRate_Get(response)
				assert.Greater(t, uint32(downPeak), uint32(0))
			}
		})
	}
}

func TestHuaweiAcceptEnhancer_Enhance_RateCalculation(t *testing.T) {
	enhancer := NewHuaweiAcceptEnhancer()
	ctx := context.Background()

	tests := []struct {
		name           string
		upRate         int
		downRate       int
		expectedUpAvg  uint32
		expectedDownAvg uint32
	}{
		{
			name:           "normal rates",
			upRate:         100,
			downRate:       200,
			expectedUpAvg:  100 * 1024,
			expectedDownAvg: 200 * 1024,
		},
		{
			name:           "zero rates",
			upRate:         0,
			downRate:       0,
			expectedUpAvg:  0,
			expectedDownAvg: 0,
		},
		{
			name:           "max int32 boundary",
			upRate:         2097152, // Will exceed MaxInt32 after * 1024
			downRate:       2097152,
			expectedUpAvg:  math.MaxInt32,
			expectedDownAvg: math.MaxInt32,
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
				VendorCode: vendorHuawei,
			}

			authCtx := &auth.AuthContext{
				Response: response,
				User:     user,
				Nas:      nas,
			}

			err := enhancer.Enhance(ctx, authCtx)
			require.NoError(t, err)

			upAvg := huawei.HuaweiInputAverageRate_Get(response)
			assert.Equal(t, tt.expectedUpAvg, uint32(upAvg))

			downAvg := huawei.HuaweiOutputAverageRate_Get(response)
			assert.Equal(t, tt.expectedDownAvg, uint32(downAvg))

			// Peak 速率应该是平均速率的 4 倍（有上限）
			upPeak := huawei.HuaweiInputPeakRate_Get(response)
			expectedUpPeak := clampInt64(int64(tt.expectedUpAvg)*4, math.MaxInt32)
			assert.Equal(t, uint32(expectedUpPeak), uint32(upPeak))

			downPeak := huawei.HuaweiOutputPeakRate_Get(response)
			expectedDownPeak := clampInt64(int64(tt.expectedDownAvg)*4, math.MaxInt32)
			assert.Equal(t, uint32(expectedDownPeak), uint32(downPeak))
		})
	}
}

func TestHuaweiAcceptEnhancer_Enhance_NoNas(t *testing.T) {
	enhancer := NewHuaweiAcceptEnhancer()
	ctx := context.Background()

	response := radius.New(radius.CodeAccessAccept, []byte("secret"))
	user := &domain.RadiusUser{
		Username: "testuser",
		UpRate:   1024,
		DownRate: 2048,
	}

	authCtx := &auth.AuthContext{
		Response: response,
		User:     user,
		Nas:      nil, // No NAS
	}

	err := enhancer.Enhance(ctx, authCtx)
	require.NoError(t, err)

	// 没有 NAS 时不应该添加属性
	upAvg := huawei.HuaweiInputAverageRate_Get(response)
	assert.Equal(t, huawei.HuaweiInputAverageRate(0), upAvg)
}
