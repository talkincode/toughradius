package enhancers

import (
	"context"
	"math"

	"github.com/talkincode/toughradius/v9/internal/radiusd/plugins/auth"
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
	if !matchVendor(authCtx, vendorIkuai) {
		return nil
	}

	user := authCtx.User
	resp := authCtx.Response

	up := clampInt64(int64(user.UpRate)*1024*8, math.MaxInt32)
	down := clampInt64(int64(user.DownRate)*1024*8, math.MaxInt32)

	ikuai.RPUpstreamSpeedLimit_Set(resp, ikuai.RPUpstreamSpeedLimit(up))
	ikuai.RPDownstreamSpeedLimit_Set(resp, ikuai.RPDownstreamSpeedLimit(down))
	return nil
}
