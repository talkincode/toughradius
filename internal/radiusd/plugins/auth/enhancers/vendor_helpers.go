package enhancers

import (
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
)

func matchVendor(ctx *auth.AuthContext, vendorCode string) bool {
	if ctx == nil || ctx.Nas == nil {
		return false
	}
	return ctx.Nas.VendorCode == vendorCode
}

func clampInt64(val int64, max int64) int64 {
	if val > max {
		return max
	}
	return val
}
