package enhancers

import (
	"context"
	"math"
	"net"
	"time"

	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
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

	timeout := int64(time.Until(user.ExpireTime).Seconds())
	if timeout > math.MaxInt32 {
		timeout = math.MaxInt32
	}
	if timeout < 0 {
		timeout = 0
	}

	interim := getIntConfig(app.ConfigRadiusAcctInterimInterval, 120)

	rfc2865.SessionTimeout_Set(response, rfc2865.SessionTimeout(timeout))
	rfc2869.AcctInterimInterval_Set(response, rfc2869.AcctInterimInterval(interim))

	if common.IsNotEmptyAndNA(user.AddrPool) {
		rfc2869.FramedPool_SetString(response, user.AddrPool)
	}
	if common.IsNotEmptyAndNA(user.IpAddr) {
		rfc2865.FramedIPAddress_Set(response, net.ParseIP(user.IpAddr))
	}

	return nil
}

func getIntConfig(name string, def int64) int64 {
	val := app.GApp().ConfigMgr().GetInt64("radius", name)
	if val == 0 {
		return def
	}
	return val
}
