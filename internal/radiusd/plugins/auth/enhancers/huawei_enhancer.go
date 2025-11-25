package enhancers

import (
	"context"
	"math"
	"net"
	"strings"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/huawei"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

type HuaweiAcceptEnhancer struct{}

func NewHuaweiAcceptEnhancer() *HuaweiAcceptEnhancer {
	return &HuaweiAcceptEnhancer{}
}

func (e *HuaweiAcceptEnhancer) Name() string {
	return "accept-huawei"
}

func (e *HuaweiAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeHuawei) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	// Use getter methods for bandwidth rates
	upRate := user.GetUpRate(profileCache)
	downRate := user.GetDownRate(profileCache)

	up := clampInt64(int64(upRate)*1024, math.MaxInt32)
	down := clampInt64(int64(downRate)*1024, math.MaxInt32)
	upPeak := clampInt64(up*4, math.MaxInt32)
	downPeak := clampInt64(down*4, math.MaxInt32)

	_ = huawei.HuaweiInputAverageRate_Set(resp, huawei.HuaweiInputAverageRate(up))     //nolint:errcheck
	_ = huawei.HuaweiInputPeakRate_Set(resp, huawei.HuaweiInputPeakRate(upPeak))       //nolint:errcheck
	_ = huawei.HuaweiOutputAverageRate_Set(resp, huawei.HuaweiOutputAverageRate(down)) //nolint:errcheck
	_ = huawei.HuaweiOutputPeakRate_Set(resp, huawei.HuaweiOutputPeakRate(downPeak))   //nolint:errcheck

	// Set Huawei FramedIPv6Address if user has a fixed IPv6 address
	if common.IsNotEmptyAndNA(user.IpV6Addr) {
		// Parse IPv6 address (without prefix length)
		ipv6Str := user.IpV6Addr
		if strings.Contains(ipv6Str, "/") {
			// Remove prefix length if present
			parts := strings.SplitN(ipv6Str, "/", 2)
			ipv6Str = parts[0]
		}
		if ipv6Addr := net.ParseIP(ipv6Str); ipv6Addr != nil {
			_ = huawei.HuaweiFramedIPv6Address_Set(resp, ipv6Addr) //nolint:errcheck
		}
	}

	// Use getter method for Domain
	domain := user.GetDomain(profileCache)
	if common.IsNotEmptyAndNA(domain) {
		_ = huawei.HuaweiDomainName_SetString(resp, domain) //nolint:errcheck
	}

	return nil
}
