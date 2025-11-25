package enhancers

import (
	"context"
	"math"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors"
	"github.com/talkincode/toughradius/v9/internal/radiusd/vendors/ikuai"
)

type IkuaiAcceptEnhancer struct{}

func NewIkuaiAcceptEnhancer() *IkuaiAcceptEnhancer {
	return &IkuaiAcceptEnhancer{}
}

func (e *IkuaiAcceptEnhancer) Name() string {
	return "accept-ikuai"
}

func (e *IkuaiAcceptEnhancer) Enhance(ctx context.Context, authCtx *auth.AuthContext) error {
	if authCtx == nil || authCtx.Response == nil || authCtx.User == nil {
		return nil
	}
	if !matchVendor(authCtx, vendors.CodeIkuai) {
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

	up := clampInt64(int64(upRate)*1024*8, math.MaxInt32)
	down := clampInt64(int64(downRate)*1024*8, math.MaxInt32)

	_ = ikuai.RPUpstreamSpeedLimit_Set(resp, ikuai.RPUpstreamSpeedLimit(up))       //nolint:errcheck
	_ = ikuai.RPDownstreamSpeedLimit_Set(resp, ikuai.RPDownstreamSpeedLimit(down)) //nolint:errcheck
	return nil
}
