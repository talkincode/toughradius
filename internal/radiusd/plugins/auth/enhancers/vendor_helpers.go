package enhancers

import "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"

const (
	vendorHuawei   = "2011"
	vendorH3C      = "25506"
	vendorZTE      = "3902"
	vendorMikrotik = "14988"
	vendorIkuai    = "10055"
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
