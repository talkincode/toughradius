package enhancers

import (
	"context"
	"math"
	"net"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
	"layeh.com/radius/rfc3162"
)

// DefaultAcceptEnhancer sets standard RADIUS attributes
type DefaultAcceptEnhancer struct{}

func NewDefaultAcceptEnhancer() *DefaultAcceptEnhancer {
	return &DefaultAcceptEnhancer{}
}

func (e *DefaultAcceptEnhancer) Name() string {
	return "default-accept"
}

func (e *DefaultAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}

	user := authCtx.User
	response := authCtx.Response

	// Get profile cache from metadata
	var profileCache interface{}
	if authCtx.Metadata != nil {
		profileCache = authCtx.Metadata["profile_cache"]
	}

	timeout := int64(time.Until(user.ExpireTime).Seconds())
	if timeout > math.MaxInt32 {
		timeout = math.MaxInt32
	}
	if timeout < 0 {
		timeout = 0
	}

	interim := getIntConfig(authCtx, app.ConfigRadiusAcctInterimInterval, 120)

	_ = rfc2865.SessionTimeout_Set(response, rfc2865.SessionTimeout(timeout))           //nolint:errcheck
	_ = rfc2869.AcctInterimInterval_Set(response, rfc2869.AcctInterimInterval(interim)) //nolint:errcheck

	// Use getter method for AddrPool
	addrPool := user.GetAddrPool(profileCache)
	if common.IsNotEmptyAndNA(addrPool) {
		_ = rfc2869.FramedPool_SetString(response, addrPool) //nolint:errcheck
	}

	// User-specific IP address (always use direct access)
	if common.IsNotEmptyAndNA(user.IpAddr) {
		_ = rfc2865.FramedIPAddress_Set(response, net.ParseIP(user.IpAddr)) //nolint:errcheck
	}

	// Set FramedIPv6Prefix if user has a fixed IPv6 address
	if common.IsNotEmptyAndNA(user.IpV6Addr) {
		// IPv6 prefix format: address/prefix-length (e.g., "2001:db8::1/64")
		// If only address is provided, append /128 for single host
		ipv6Prefix := user.IpV6Addr
		if !strings.Contains(ipv6Prefix, "/") {
			ipv6Prefix = ipv6Prefix + "/128"
		}
		if _, ipnet, err := net.ParseCIDR(ipv6Prefix); err == nil {
			_ = rfc3162.FramedIPv6Prefix_Set(response, ipnet) //nolint:errcheck
		}
	}

	// Use getter method for IPv6PrefixPool
	ipv6Pool := user.GetIPv6PrefixPool(profileCache)
	if common.IsNotEmptyAndNA(ipv6Pool) {
		_ = rfc3162.FramedIPv6Pool_SetString(response, ipv6Pool) //nolint:errcheck
	}

	return nil
}

func getIntConfig(authCtx *auth.AuthContext, name string, def int64) int64 {
	// Get config manager from metadata
	if authCtx.Metadata != nil {
		if cfgMgr, ok := authCtx.Metadata["config_mgr"].(*app.ConfigManager); ok {
			val := cfgMgr.GetInt64("radius", name)
			if val == 0 {
				return def
			}
			return val
		}
	}
	return def
}
