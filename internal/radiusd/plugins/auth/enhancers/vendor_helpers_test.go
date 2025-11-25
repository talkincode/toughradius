package enhancers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
)

func TestMatchVendor(t *testing.T) {
	tests := []struct {
		name       string
		authCtx    *auth.AuthContext
		vendorCode string
		expected   bool
	}{
		{
			name:       "nil context",
			authCtx:    nil,
			vendorCode: vendors.CodeHuawei,
			expected:   false,
		},
		{
			name: "nil NAS",
			authCtx: &auth.AuthContext{
				Nas: nil,
			},
			vendorCode: vendors.CodeHuawei,
			expected:   false,
		},
		{
			name: "matching vendor",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeHuawei,
				},
			},
			vendorCode: vendors.CodeHuawei,
			expected:   true,
		},
		{
			name: "non-matching vendor",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeHuawei,
				},
			},
			vendorCode: vendors.CodeH3C,
			expected:   false,
		},
		{
			name: "huawei match",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeHuawei,
				},
			},
			vendorCode: vendors.CodeHuawei,
			expected:   true,
		},
		{
			name: "h3c match",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeH3C,
				},
			},
			vendorCode: vendors.CodeH3C,
			expected:   true,
		},
		{
			name: "zte match",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeZTE,
				},
			},
			vendorCode: vendors.CodeZTE,
			expected:   true,
		},
		{
			name: "mikrotik match",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeMikrotik,
				},
			},
			vendorCode: vendors.CodeMikrotik,
			expected:   true,
		},
		{
			name: "ikuai match",
			authCtx: &auth.AuthContext{
				Nas: &domain.NetNas{
					VendorCode: vendors.CodeIkuai,
				},
			},
			vendorCode: vendors.CodeIkuai,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchVendor(tt.authCtx, tt.vendorCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClampInt64(t *testing.T) {
	tests := []struct {
		name     string
		val      int64
		max      int64
		expected int64
	}{
		{
			name:     "below max",
			val:      100,
			max:      1000,
			expected: 100,
		},
		{
			name:     "at max",
			val:      1000,
			max:      1000,
			expected: 1000,
		},
		{
			name:     "above max",
			val:      2000,
			max:      1000,
			expected: 1000,
		},
		{
			name:     "zero value",
			val:      0,
			max:      1000,
			expected: 0,
		},
		{
			name:     "negative value",
			val:      -100,
			max:      1000,
			expected: -100,
		},
		{
			name:     "max int32",
			val:      math.MaxInt32 + 1000,
			max:      math.MaxInt32,
			expected: math.MaxInt32,
		},
		{
			name:     "exactly max int32",
			val:      math.MaxInt32,
			max:      math.MaxInt32,
			expected: math.MaxInt32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampInt64(tt.val, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}
