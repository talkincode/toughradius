package enhancers

import (
	"context"
	"math"

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

	_ = huawei.HuaweiInputAverageRate_Set(resp, huawei.HuaweiInputAverageRate(up))     //nolint:errcheck,gosec // G115: clamped to MaxInt32
	_ = huawei.HuaweiInputPeakRate_Set(resp, huawei.HuaweiInputPeakRate(upPeak))       //nolint:errcheck,gosec // G115: clamped to MaxInt32
	_ = huawei.HuaweiOutputAverageRate_Set(resp, huawei.HuaweiOutputAverageRate(down)) //nolint:errcheck,gosec // G115: clamped to MaxInt32
	_ = huawei.HuaweiOutputPeakRate_Set(resp, huawei.HuaweiOutputPeakRate(downPeak))   //nolint:errcheck,gosec // G115: clamped to MaxInt32

	// Set Huawei-Framed-IPv6-Address when the user has a single static IPv6 host
	// address (a bare address or an explicit /128). singleIPv6Host rejects IPv4
	// literals (including IPv4-mapped values that net.ParseIP would otherwise
	// accept) and multi-host prefixes, so a misconfigured IpV6Addr is skipped
	// instead of advertising a malformed IPv6 address to the Huawei NAS.
	if common.IsNotEmptyAndNA(user.IpV6Addr) {
		if ipv6Addr, ok := singleIPv6Host(user.IpV6Addr); ok {
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
